package clickhouse

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func dockerPool(t *testing.T) *dockertest.Pool {
	t.Helper()
	if testing.Short() {
		t.Skip("ClickHouse integration tests require Docker")
	}
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	if err := pool.Client.Ping(); err != nil {
		t.Skipf("Docker unavailable: %v", err)
	}
	pool.MaxWait = 2 * time.Minute
	return pool
}

func TestSingleInstance(t *testing.T) {
	for _, secure := range []bool{false, true} {
		name := "plain"
		if secure {
			name = "tls"
		}
		t.Run(name, func(t *testing.T) {
			pool := dockerPool(t)
			options := []Opt{WithUser("fixture_user"), WithPassword("a:b/?&+#"), WithDatabase("fixture_db"), WithPrintLogsOnError(true), WithMemory(1 << 30), WithEnv("TZ=UTC"), WithConfig("<clickhouse><timezone>UTC</timezone></clickhouse>"), WithUsersConfig("<clickhouse><profiles><default><max_threads>2</max_threads></default></profiles></clickhouse>")}
			if secure {
				options = append(options, WithTLS())
			}
			result, err := Setup(pool, t, options...)
			require.NoError(t, err)
			require.Len(t, result.Nodes, 1)
			require.Empty(t, result.ClusterName)
			require.Equal(t, "127.0.0.1", result.Host)
			for _, protocol := range []ch.Protocol{ch.Native, ch.HTTP} {
				db, err := result.OpenDB(protocol)
				require.NoError(t, err)
				var one uint8
				require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
				require.EqualValues(t, 1, one)
				var user, database, timezone string
				require.NoError(t, db.QueryRow("SELECT currentUser(), currentDatabase(), timezone()").Scan(&user, &database, &timezone))
				require.Equal(t, result.User, user)
				require.Equal(t, result.Database, database)
				require.Equal(t, "UTC", timezone)
				require.NoError(t, db.Close())
			}
			container, err := pool.Client.InspectContainer(result.ContainerID)
			require.NoError(t, err)
			require.EqualValues(t, 1<<30, container.HostConfig.Memory)
			published := 0
			for port, bindings := range container.NetworkSettings.Ports {
				if len(bindings) > 0 {
					published++
					if secure {
						require.Contains(t, []string{"8443/tcp", "9440/tcp"}, string(port))
					}
				}
				for _, binding := range bindings {
					require.Equal(t, "127.0.0.1", binding.HostIP)
				}
			}
			require.Equal(t, 2, published)
			if secure {
				require.NotEmpty(t, result.CAPEM)
				untrusted, err := ch.ParseDSN(result.HTTPDSN())
				require.NoError(t, err)
				untrusted.TLS = &tls.Config{RootCAs: x509.NewCertPool(), MinVersion: tls.VersionTLS12}
				db := ch.OpenDB(untrusted)
				require.Error(t, db.Ping())
				require.NoError(t, db.Close())
			}
		})
	}
}

func TestCluster(t *testing.T) {
	pool := dockerPool(t)
	result, err := Setup(pool, t, WithCluster(2, 2), WithTLS(), WithMemory(1<<30), WithPrintLogsOnError(true))
	require.NoError(t, err)
	require.Len(t, result.Nodes, 4)
	require.NotEmpty(t, result.NetworkID)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, node := range result.Nodes {
		var count uint64
		require.NoError(t, node.DB.QueryRowContext(ctx, "SELECT count() FROM system.clusters WHERE cluster = ?", result.ClusterName).Scan(&count))
		require.EqualValues(t, 4, count)
		var shard, replica string
		require.NoError(t, node.DB.QueryRowContext(ctx, "SELECT getMacro('shard'), getMacro('replica')").Scan(&shard, &replica))
		require.Equal(t, node.Macros.Shard, shard)
		require.Equal(t, node.Macros.Replica, replica)
	}
	_, err = result.DB.ExecContext(ctx, "CREATE TABLE replicated ON CLUSTER "+result.ClusterName+" (id UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/integration', '{replica}') ORDER BY id")
	require.NoError(t, err)
	_, err = result.Nodes[0].DB.ExecContext(ctx, "INSERT INTO replicated VALUES (42)")
	require.NoError(t, err)
	_, err = result.Nodes[1].DB.ExecContext(ctx, "SYSTEM SYNC REPLICA replicated")
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		var value uint64
		err := result.Nodes[1].DB.QueryRowContext(ctx, "SELECT id FROM replicated").Scan(&value)
		return err == nil && value == 42
	}, 30*time.Second, 100*time.Millisecond)
	var count uint64
	require.NoError(t, result.Nodes[2].DB.QueryRowContext(ctx, "SELECT count() FROM replicated").Scan(&count))
	require.Zero(t, count, "replication stays within a shard")
	_, err = result.DB.ExecContext(ctx, "DROP TABLE replicated ON CLUSTER "+result.ClusterName+" SYNC")
	require.NoError(t, err)
}

func TestProvidedNetwork(t *testing.T) {
	pool := dockerPool(t)
	network, err := pool.CreateNetwork("clickhouse-test-" + filepath.Base(t.TempDir()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.RemoveNetwork(network)) })
	t.Run("resource", func(t *testing.T) {
		result, err := Setup(pool, t, WithCluster(1, 2), WithNetwork(network.Network), WithPrintLogsOnError(true))
		require.NoError(t, err)
		require.Equal(t, network.Network.ID, result.NetworkID)
	})
	_, err = pool.Client.NetworkInfo(network.Network.ID)
	require.NoError(t, err, "caller-owned network survives resource cleanup")
}

func TestWriteConfig(t *testing.T) {
	config := defaultConfig()
	WithTLS()(&config)
	WithConfig("<clickhouse><timezone>UTC</timezone></clickhouse>")(&config)
	WithUsersConfig("<clickhouse><profiles><default><max_threads>2</max_threads></default></profiles></clickhouse>")(&config)
	nodes := topology("fixture", config)
	mounts, err := writeConfig(t.TempDir(), config, nodes, 0, []byte("cert"), []byte("key"))
	require.NoError(t, err)
	require.Len(t, mounts, 5)
	for _, mount := range mounts {
		source, _, found := strings.Cut(mount, ":")
		require.True(t, found)
		require.True(t, strings.HasSuffix(mount, ":ro"))
		info, err := os.Stat(source)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o644), info.Mode().Perm())
		require.NotEqual(t, "config.d", filepath.Base(source))
	}
}

func TestReadinessRejectsHTTPFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "/ping", request.URL.Path)
		user, password, ok := request.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "user", user)
		require.Equal(t, "password", password)
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
	})}
	err := ready(client, &Node{Host: "127.0.0.1", HTTPPort: "8123", User: "user", Password: "password"})
	require.ErrorContains(t, err, "503")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func TestReadinessQueriesConfiguredDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("Ok.\n"))}, nil
	})}
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	require.NoError(t, ready(client, &Node{DB: db, Host: "127.0.0.1", HTTPPort: "8123"}))
	mock.ExpectQuery("SELECT 1").WillReturnError(context.DeadlineExceeded)
	require.ErrorIs(t, ready(client, &Node{DB: db, Host: "127.0.0.1", HTTPPort: "8123"}), context.DeadlineExceeded)
	require.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestOpenDBInvalidProtocol(t *testing.T) {
	db, err := (&Node{}).OpenDB(ch.Protocol(255))
	require.Error(t, err)
	require.Nil(t, db)
}

func TestClusterReadiness(t *testing.T) {
	config := defaultConfig()
	WithCluster(1, 2)(&config)
	result := &Resource{Nodes: topology("fixture", config), ClusterName: clusterName}
	result.Node = result.Nodes[0]
	mocks := make([]sqlmock.Sqlmock, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		node.DB = db
		t.Cleanup(func() { require.NoError(t, db.Close()) })
		mock.ExpectQuery("SELECT count\\(\\) FROM system.clusters").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mocks = append(mocks, mock)
	}
	mocks[0].ExpectExec("CREATE TABLE IF NOT EXISTS.*ON CLUSTER rudder_cluster.*ReplicatedMergeTree").WillReturnResult(sqlmock.NewResult(0, 0))
	for _, mock := range mocks {
		mock.ExpectQuery("SELECT count\\(\\) FROM system.replicas").WithArgs(config.Database).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	mocks[0].ExpectExec("DROP TABLE IF EXISTS.*ON CLUSTER rudder_cluster SYNC").WillReturnResult(sqlmock.NewResult(0, 0))
	require.NoError(t, clusterReady(result))
	for _, mock := range mocks {
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
	}
}

func TestClusterReadinessRejectsIncompleteTopology(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	node := &Node{DB: db}
	mock.ExpectQuery("SELECT count\\(\\) FROM system.clusters").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	require.ErrorContains(t, clusterReady(&Resource{Node: node, Nodes: []*Node{node, {}}, ClusterName: clusterName}), "expected 2")
	require.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestClusterReadinessRejectsReadOnlyReplica(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	node := &Node{DB: db, Database: "fixture", Hostname: "node"}
	mock.ExpectQuery("SELECT count\\(\\) FROM system.clusters").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT count\\(\\) FROM system.replicas").WithArgs(node.Database).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	require.ErrorContains(t, clusterReady(&Resource{Node: node, Nodes: []*Node{node}, ClusterName: clusterName}), "not ready")
	require.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectClose()
}

func TestQuoteIdentifier(t *testing.T) {
	require.Equal(t, "`db\\`name\\\\path`", quoteIdentifier("db`name\\path"))
	require.Equal(t, "`database`", quoteIdentifier("database"))
}
