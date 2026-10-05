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
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
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

const (
	fixtureTimezone      = "Asia/Kolkata"
	fixtureMaxResultRows = "4242"
	usersConfig          = "<clickhouse><profiles><default><max_result_rows>" + fixtureMaxResultRows + "</max_result_rows></default></profiles></clickhouse>"
)

func TestSingleInstance(t *testing.T) {
	for _, secure := range []bool{false, true} {
		name := "plain"
		if secure {
			name = "tls"
		}
		t.Run(name, func(t *testing.T) {
			pool := dockerPool(t)
			options := []Opt{WithUser("fixture_user"), WithPassword("a:b/?&+#"), WithDatabase("fixture_db"), WithPrintLogsOnError(true), WithMemory(1 << 30), WithEnv("RESOURCE_TZ=" + fixtureTimezone), WithConfig(`<clickhouse><timezone from_env="RESOURCE_TZ"/></clickhouse>`), WithUsersConfig(usersConfig)}
			if secure {
				options = append(options, WithTLS())
			}
			result, err := Setup(pool, t, options...)
			require.NoError(t, err)
			require.Equal(t, "fixture_user", result.User)
			require.Equal(t, "a:b/?&+#", result.Password)
			require.Equal(t, "fixture_db", result.Database)
			require.Len(t, result.Nodes, 1)
			require.Empty(t, result.ClusterName)
			require.Equal(t, "127.0.0.1", result.Host)
			for _, protocol := range []ch.Protocol{ch.Native, ch.HTTP} {
				db, err := result.OpenDB(protocol)
				require.NoError(t, err)
				var one uint8
				require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
				require.EqualValues(t, 1, one)
				var user, database, timezone, maxResultRows string
				require.NoError(t, db.QueryRow("SELECT currentUser(), currentDatabase(), timezone(), toString(getSetting('max_result_rows'))").Scan(&user, &database, &timezone, &maxResultRows))
				require.Equal(t, "fixture_user", user)
				require.Equal(t, "fixture_db", database)
				require.Equal(t, fixtureTimezone, timezone, "the timezone comes from WithEnv through WithConfig")
				require.Equal(t, fixtureMaxResultRows, maxResultRows, "the profile comes from WithUsersConfig")
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

func TestDigestOnlyImage(t *testing.T) {
	pool := dockerPool(t)
	_, digest, _ := strings.Cut(DefaultImage, "@")
	result, err := Setup(pool, t, WithImage("clickhouse/clickhouse-server@"+digest), WithPrintLogsOnError(true))
	require.NoError(t, err)
	var version string
	require.NoError(t, result.DB.QueryRow("SELECT version()").Scan(&version))
	require.True(t, strings.HasPrefix(version, "26.3."), version)
}

func TestProvidedNetwork(t *testing.T) {
	pool := dockerPool(t)
	network, err := pool.CreateNetwork("clickhouse-test-" + filepath.Base(t.TempDir()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.RemoveNetwork(network)) })
	var containerIDs []string
	for name, options := range map[string][]Opt{"single": nil, "cluster": {WithCluster(1, 2)}} {
		t.Run(name, func(t *testing.T) {
			result, err := Setup(pool, t, append(options, WithNetwork(network.Network), WithPrintLogsOnError(true))...)
			require.NoError(t, err)
			for _, node := range result.Nodes {
				container, err := pool.Client.InspectContainer(node.ContainerID)
				require.NoError(t, err)
				attached, ok := container.NetworkSettings.Networks[network.Network.Name]
				require.True(t, ok, "node %s joins the caller network", node.Hostname)
				require.Equal(t, network.Network.ID, attached.NetworkID)
				containerIDs = append(containerIDs, node.ContainerID)
			}
		})
	}
	require.Len(t, containerIDs, 3)
	for _, id := range containerIDs {
		_, err := pool.Client.InspectContainer(id)
		var missing *docker.NoSuchContainer
		require.ErrorAs(t, err, &missing, "cleanup purges container %s", id)
	}
	_, err = pool.Client.NetworkInfo(network.Network.ID)
	require.NoError(t, err, "caller-owned network survives resource cleanup")
}

func TestWriteConfig(t *testing.T) {
	config := defaultConfig()
	WithTLS()(&config)
	WithConfig("<clickhouse><timezone>UTC</timezone></clickhouse>")(&config)
	WithUsersConfig(usersConfig)(&config)
	nodes := topology("fixture", config)
	mounts, err := writeConfig(t.TempDir(), config, nodes, 0, []byte("cert"), []byte("key"))
	require.NoError(t, err)
	got := make(map[string]string, len(mounts))
	for _, mount := range mounts {
		parts := strings.Split(mount, ":")
		require.Len(t, parts, 3, mount)
		require.Equal(t, "ro", parts[2])
		info, err := os.Stat(parts[0])
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o644), info.Mode().Perm())
		content, err := os.ReadFile(parts[0])
		require.NoError(t, err)
		got[parts[1]] = string(content)
	}
	require.Equal(t, map[string]string{
		"/etc/clickhouse-server/config.d/resource.xml": serverConfig(config, nodes, 0),
		"/etc/clickhouse-server/config.d/zz_extra.xml": config.ConfigXML,
		"/etc/clickhouse-server/users.d/zz_extra.xml":  usersConfig,
		"/etc/clickhouse-server/certs/server.pem":      "cert",
		"/etc/clickhouse-server/certs/server.key":      "key",
	}, got)
}

func TestReadinessRejectsHTTPFailure(t *testing.T) {
	node, _ := mockNode(t)
	node.Host, node.HTTPPort, node.User, node.Password = "127.0.0.1", "8123", "user", "password"
	var path, user, password string
	var authenticated bool
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		path = request.URL.Path
		user, password, authenticated = request.BasicAuth()
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
	})}
	err := ready(client, node)
	require.ErrorContains(t, err, "503")
	require.Equal(t, "/ping", path)
	require.True(t, authenticated)
	require.Equal(t, "user", user)
	require.Equal(t, "password", password)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func mockNode(t *testing.T) (*Node, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, db.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	})
	return &Node{DB: db}, mock
}

func TestReadinessQueriesSQLAfterPing(t *testing.T) {
	node, mock := mockNode(t)
	node.Host, node.HTTPPort = "127.0.0.1", "8123"
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("Ok.\n"))}, nil
	})}
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	require.NoError(t, ready(client, node))
	mock.ExpectQuery("SELECT 1").WillReturnError(context.DeadlineExceeded)
	require.ErrorIs(t, ready(client, node), context.DeadlineExceeded)
}

func TestOpenDBInvalidProtocol(t *testing.T) {
	db, err := (&Node{}).OpenDB(ch.Protocol(255))
	require.Error(t, err)
	require.Nil(t, db)
}

// quotedCluster is the exact form the readiness DDL must contain.
const quotedCluster = "`" + clusterName + "`"

func TestClusterReadiness(t *testing.T) {
	config := defaultConfig()
	WithCluster(2, 2)(&config)
	result := &Resource{Nodes: topology("fixture", config), ClusterName: clusterName}
	result.Node = result.Nodes[0]
	mocks := make([]sqlmock.Sqlmock, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		mocked, mock := mockNode(t)
		node.DB = mocked.DB
		mock.ExpectQuery("SELECT count() FROM system.clusters WHERE cluster = ?").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
		mocks = append(mocks, mock)
	}
	table := "`rudderdb`.`__dockertest_readiness`"
	mocks[0].ExpectExec("CREATE TABLE IF NOT EXISTS " + table + " ON CLUSTER " + quotedCluster + " (id UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/dockertest_readiness', '{replica}') ORDER BY id").WillReturnResult(sqlmock.NewResult(0, 0))
	for _, mock := range mocks {
		mock.ExpectQuery("SELECT count() FROM system.replicas WHERE database = ? AND table = '__dockertest_readiness' AND is_readonly = 0 AND is_session_expired = 0").WithArgs(config.Database).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	mocks[0].ExpectExec("INSERT INTO " + table + " (id) VALUES (?)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mocks[2].ExpectExec("INSERT INTO " + table + " (id) VALUES (?)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	for index, mock := range mocks {
		if index == 1 {
			mock.ExpectQuery("SELECT count() FROM " + table + " WHERE id = ?").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		}
		mock.ExpectQuery("SELECT count() FROM " + table + " WHERE id = ?").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	mocks[0].ExpectExec("DROP TABLE IF EXISTS " + table + " ON CLUSTER " + quotedCluster + " SYNC").WillReturnResult(sqlmock.NewResult(0, 0))
	require.NoError(t, clusterReady(result))
}

func TestReadinessRowTimesOutForReplica(t *testing.T) {
	node, mock := mockNode(t)
	node.Hostname = "replica-2"
	mock.ExpectQuery("SELECT count() FROM `fixture`.`readiness` WHERE id = ?").WithArgs(uint64(42)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := waitForReadinessRow(ctx, node, "`fixture`.`readiness`", 42)
	require.ErrorContains(t, err, "replica-2")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestClusterReadinessRejectsIncompleteTopology(t *testing.T) {
	node, mock := mockNode(t)
	mock.ExpectQuery("SELECT count() FROM system.clusters WHERE cluster = ?").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// The second node has a DB with no expectations, so a missing count check fails on a query, not a nil pointer.
	unexpected, _ := mockNode(t)
	require.ErrorContains(t, clusterReady(&Resource{Node: node, Nodes: []*Node{node, unexpected}, ClusterName: clusterName}), "expected 2")
}

func TestClusterReadinessRejectsReadOnlyReplica(t *testing.T) {
	node, mock := mockNode(t)
	node.Database, node.Hostname = "fixture", "node"
	mock.ExpectQuery("SELECT count() FROM system.clusters WHERE cluster = ?").WithArgs(clusterName).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS `fixture`.`__dockertest_readiness` ON CLUSTER " + quotedCluster + " (id UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/dockertest_readiness', '{replica}') ORDER BY id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT count() FROM system.replicas WHERE database = ? AND table = '__dockertest_readiness' AND is_readonly = 0 AND is_session_expired = 0").WithArgs(node.Database).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	require.ErrorContains(t, clusterReady(&Resource{Node: node, Nodes: []*Node{node}, ClusterName: clusterName}), "not ready")
}

func TestQuoteIdentifier(t *testing.T) {
	require.Equal(t, "`db\\`name\\\\path`", quoteIdentifier("db`name\\path"))
	require.Equal(t, "`database`", quoteIdentifier("database"))
}

type countingCleaner struct{ cleanups int }

func (c *countingCleaner) Cleanup(func())    { c.cleanups++ }
func (*countingCleaner) Log(...any)          {}
func (*countingCleaner) Logf(string, ...any) {}
func (*countingCleaner) Failed() bool        { return false }

func TestSetupFailureRegistersNoCleanup(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	pool, err := dockertest.NewPool("tcp://127.0.0.1:1")
	require.NoError(t, err)
	cleaner := &countingCleaner{}
	_, err = Setup(pool, cleaner)
	require.ErrorContains(t, err, "pulling ClickHouse image")
	require.Zero(t, cleaner.cleanups, "a failed Setup tears down in place")
	entries, err := os.ReadDir(tmp)
	require.NoError(t, err)
	require.Empty(t, entries, "a failed Setup removes its configuration directory")
}

func TestSetupPanicTearsDownAndRepanics(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	cleaner := &countingCleaner{}
	// A pool without a client panics at the first Docker call, after Setup has created its directory.
	require.Panics(t, func() { _, _ = Setup(&dockertest.Pool{}, cleaner) })
	require.Zero(t, cleaner.cleanups, "a panicking Setup tears down in place")
	entries, err := os.ReadDir(tmp)
	require.NoError(t, err)
	require.Empty(t, entries, "a panicking Setup removes its configuration directory")
}

func TestSetupRejectsNilPoolOrCleanerBeforeOptions(t *testing.T) {
	pool, err := dockertest.NewPool("tcp://127.0.0.1:1")
	require.NoError(t, err)
	for name, args := range map[string]struct {
		pool    *dockertest.Pool
		cleaner resource.Cleaner
	}{
		"nil pool":    {cleaner: &countingCleaner{}},
		"nil cleaner": {pool: pool},
	} {
		t.Run(name, func(t *testing.T) {
			var applied bool
			_, err := Setup(args.pool, args.cleaner, func(*Config) { applied = true })
			require.EqualError(t, err, "pool and cleaner must not be nil")
			require.False(t, applied, "Setup checks its arguments before it applies options")
		})
	}
}
