package clickhouse

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"net/url"
	"strings"
	"testing"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	config := defaultConfig()
	require.NoError(t, config.validate())
	require.Equal(t, DefaultImage, config.Image)
	require.Zero(t, config.Shards)
	require.Zero(t, config.Replicas)
	WithTag("25.8")(&config)
	require.Equal(t, "clickhouse/clickhouse-server:25.8", config.Image)
	WithImage("localhost:5000/clickhouse:test")(&config)
	require.Equal(t, "localhost:5000/clickhouse:test", config.Image)

	valid := map[string][]Opt{
		"cluster":       {WithCluster(2, 2)},
		"tls":           {WithTLS()},
		"env":           {WithEnv("TZ=UTC", "EMPTY=")},
		"config xml":    {WithConfig("<?xml version=\"1.0\"?><clickhouse><timezone>UTC</timezone></clickhouse>")},
		"users xml":     {WithUsersConfig("<clickhouse><profiles/></clickhouse>")},
		"names":         {WithUser("user-1_x"), WithDatabase("_db_1")},
		"no password":   {WithPassword("")},
		"custom digest": {WithImage("localhost:5000/clickhouse@sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893")},
	}
	for name, options := range valid {
		t.Run("valid "+name, func(t *testing.T) {
			config := defaultConfig()
			for _, option := range options {
				option(&config)
			}
			require.NoError(t, config.validate())
		})
	}

	invalid := []struct {
		name   string
		option Opt
		err    string
	}{
		{"no cluster", WithCluster(0, 0), "shards and replicas"},
		{"no shards", WithCluster(0, 2), "shards and replicas"},
		{"no replicas", WithCluster(2, 0), "shards and replicas"},
		{"negative shards", WithCluster(-1, 1), "shards and replicas"},
		{"node overflow", WithCluster(2, int(^uint(0)>>1)), "overflows"},
		{"negative memory", WithMemory(-1), "memory"},
		{"empty user", WithUser(""), "user"},
		{"user with at sign", WithUser("user@x"), "user"},
		{"user with dot", WithUser("user.x"), "user"},
		{"user with leading digit", WithUser("1user"), "user"},
		{"empty database", WithDatabase(""), "database"},
		{"database with hyphen", WithDatabase("my-db"), "database"},
		{"database with space", WithDatabase("my db"), "database"},
		{"config root", WithConfig("<yandex/>"), "root must be clickhouse"},
		{"users root", WithUsersConfig("<users/>"), "root must be clickhouse"},
		{"two roots", WithConfig("<clickhouse/><clickhouse/>"), "one clickhouse document root"},
		{"empty document", WithConfig("<!-- only -->"), "one clickhouse document root"},
		{"text outside root", WithConfig("text<clickhouse/>"), "text outside"},
		{"env without value", WithEnv("TZ"), "KEY=value"},
		{"env without key", WithEnv("=UTC"), "KEY=value"},
		{"env user", WithEnv("CLICKHOUSE_USER=other"), "CLICKHOUSE_USER"},
		{"env password", WithEnv("CLICKHOUSE_PASSWORD=other"), "CLICKHOUSE_PASSWORD"},
		{"env password file", WithEnv("CLICKHOUSE_PASSWORD_FILE=/run/secrets/password"), "CLICKHOUSE_PASSWORD_FILE"},
		{"env database", WithEnv("CLICKHOUSE_DB=other"), "CLICKHOUSE_DB"},
		{"env access management", WithEnv("CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=0"), "CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT"},
		{"env skip user setup", WithEnv("CLICKHOUSE_SKIP_USER_SETUP=1"), "CLICKHOUSE_SKIP_USER_SETUP"},
		{"image", WithImage("not an image"), "parsing ClickHouse image"},
	}
	for _, tc := range invalid {
		t.Run("invalid "+tc.name, func(t *testing.T) {
			config := defaultConfig()
			tc.option(&config)
			require.ErrorContains(t, config.validate(), tc.err)
		})
	}

	t.Run("invalid xml syntax", func(t *testing.T) {
		for _, option := range []Opt{WithConfig("<clickhouse>"), WithUsersConfig("<clickhouse></users>")} {
			config := defaultConfig()
			option(&config)
			var syntaxErr *xml.SyntaxError
			require.ErrorAs(t, config.validate(), &syntaxErr)
		}
	})
}

func TestTLSRequiresPassword(t *testing.T) {
	config := defaultConfig()
	WithTLS()(&config)
	WithPassword("")(&config)
	require.ErrorContains(t, config.validate(), "TLS requires a non-empty password")
}

func verifyCertificate(t *testing.T, certificate tls.Certificate, config *tls.Config, hostname string) {
	t.Helper()
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	require.NoError(t, err)
	_, err = leaf.Verify(x509.VerifyOptions{Roots: config.RootCAs, DNSName: hostname})
	require.NoError(t, err)
}

func TestImageReference(t *testing.T) {
	const digest = "sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893"
	// A digest-only reference keeps tag "latest" on purpose: dockertest joins Repository and Tag with ":",
	// Docker ignores the tag once a digest is present, and verifyImage checks the digest.
	for image, want := range map[string][2]string{
		DefaultImage:                               {"clickhouse/clickhouse-server", "26.3@" + digest},
		"localhost:5000/clickhouse:test":           {"localhost:5000/clickhouse", "test"},
		"localhost:5000/clickhouse":                {"localhost:5000/clickhouse", "latest"},
		"clickhouse/clickhouse-server":             {"clickhouse/clickhouse-server", "latest"},
		"clickhouse/clickhouse-server@" + digest:   {"clickhouse/clickhouse-server", "latest@" + digest},
		"localhost:5000/clickhouse:test@" + digest: {"localhost:5000/clickhouse", "test@" + digest},
	} {
		repository, tag, err := splitImage(image)
		require.NoError(t, err, image)
		require.Equal(t, want, [2]string{repository, tag}, image)
	}
	_, _, err := splitImage("clickhouse/clickhouse-server@sha256:short")
	require.ErrorContains(t, err, "parsing ClickHouse image")
}

func TestDigestRepositoryAliases(t *testing.T) {
	const digest = "sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893"
	require.True(t, matchesDigest([]string{"clickhouse/clickhouse-server@" + digest}, "docker.io/clickhouse/clickhouse-server", digest))
	require.True(t, matchesDigest([]string{"docker.io/clickhouse/clickhouse-server@" + digest}, "clickhouse/clickhouse-server", digest))
	require.False(t, matchesDigest([]string{"other/server@" + digest}, "clickhouse/clickhouse-server", digest))
	require.False(t, matchesDigest([]string{"invalid digest"}, "clickhouse/clickhouse-server", digest))
}

func TestNodeDSN(t *testing.T) {
	for _, secure := range []bool{false, true} {
		node := &Node{Host: "::1", NativePort: "9000", HTTPPort: "8123", User: "user@name", Password: "a:b/?&+#", Database: "db name", secure: secure}
		for _, protocol := range []ch.Protocol{ch.Native, ch.HTTP} {
			dsn := node.NativeDSN()
			if protocol == ch.HTTP {
				dsn = node.HTTPDSN()
			}
			parsed, err := ch.ParseDSN(dsn)
			require.NoError(t, err)
			require.Equal(t, node.User, parsed.Auth.Username)
			require.Equal(t, node.Password, parsed.Auth.Password)
			require.Equal(t, node.Database, parsed.Auth.Database)
			require.Equal(t, protocol, parsed.Protocol)
			require.Equal(t, secure, parsed.TLS != nil)
			if secure {
				require.False(t, parsed.TLS.InsecureSkipVerify)
			}
			uri, err := url.Parse(dsn)
			require.NoError(t, err)
			require.Equal(t, "::1", uri.Hostname())
		}
	}
}

func TestClusterConfig(t *testing.T) {
	config := defaultConfig()
	WithCluster(2, 2)(&config)
	WithPassword("a<&'\"")(&config)
	nodes := topology("fixture", config)
	require.Len(t, nodes, 4)
	require.Equal(t, nodes[0].Macros.Shard, nodes[1].Macros.Shard)
	require.NotEqual(t, nodes[0].Macros.Shard, nodes[2].Macros.Shard)
	for index, node := range nodes {
		data := serverConfig(config, nodes, index)
		var parsed struct {
			RemoteServers struct {
				Cluster struct {
					Shards []struct {
						Replicas []struct {
							Host     string `xml:"host"`
							Password string `xml:"password"`
						} `xml:"replica"`
					} `xml:"shard"`
				} `xml:"rudder_cluster"`
			} `xml:"remote_servers"`
			Macros Macros `xml:"macros"`
		}
		require.NoError(t, xml.Unmarshal([]byte(data), &parsed))
		require.Len(t, parsed.RemoteServers.Cluster.Shards, 2)
		for _, shard := range parsed.RemoteServers.Cluster.Shards {
			require.Len(t, shard.Replicas, 2)
			for _, replica := range shard.Replicas {
				require.Equal(t, config.Password, replica.Password)
			}
		}
		require.Equal(t, node.Macros, parsed.Macros)
		require.Equal(t, index == 0, strings.Contains(data, "<keeper_server>"))
		require.Contains(t, data, "<host>fixture-s1-r1</host><port>9181</port>")
	}
}

// testTuningSettings are the server settings the resource sizes for tests by default.
// TestTestTuningApplied reads the same values back from system.server_settings.
var testTuningSettings = map[string]string{
	"background_pool_size":                          "5",
	"background_merges_mutations_concurrency_ratio": "5",
	"background_schedule_pool_size":                 "4",
	"background_move_pool_size":                     "1",
	"background_fetches_pool_size":                  "1",
	"background_common_pool_size":                   "2",
	"background_distributed_schedule_pool_size":     "1",
	"mark_cache_size":                               "16777216",
	"uncompressed_cache_size":                       "0",
	"index_mark_cache_size":                         "0",
	"asynchronous_metrics_update_period_s":          "600",
	"asynchronous_heavy_metrics_update_period_s":    "600",
}

// testTuningRemovedLogs are the system log tables the resource disables by default. query_log stays,
// because tests read it after SYSTEM FLUSH LOGS.
var testTuningRemovedLogs = []string{
	"metric_log", "trace_log", "text_log", "asynchronous_metric_log", "part_log", "processors_profile_log",
	"opentelemetry_span_log", "query_thread_log", "query_views_log", "crash_log", "background_schedule_pool_log",
}

func TestServerConfigTestTuning(t *testing.T) {
	config := defaultConfig()
	data := serverConfig(config, topology("fixture", config), 0)
	require.NoError(t, validateXML(data))
	for name, value := range testTuningSettings {
		require.Contains(t, data, "<"+name+">"+value+"</"+name+">", "the default config sets %s for tests", name)
	}
	require.Contains(t, data, "<mlock_executable>false</mlock_executable>")
	for _, table := range testTuningRemovedLogs {
		require.Contains(t, data, "<"+table+` remove="1"/>`, "the default config disables system.%s", table)
	}
	require.NotContains(t, data, "<query_log", "tests read system.query_log, so the default config keeps it")
	require.NotContains(t, data, "max_thread_pool_size", "ClickHouse hangs at startup when the global thread pool is capped below its idle need")
	require.NotContains(t, data, "max_server_memory_usage", "ClickHouse already derives the limit from the container memory")
}

func TestServerConfigProductionDefaults(t *testing.T) {
	config := defaultConfig()
	WithProductionDefaults()(&config)
	data := serverConfig(config, topology("fixture", config), 0)
	require.NoError(t, validateXML(data))
	for name := range testTuningSettings {
		require.NotContains(t, data, "<"+name+">", "WithProductionDefaults keeps the ClickHouse default for %s", name)
	}
	require.NotContains(t, data, `remove="1"`, "WithProductionDefaults keeps every system log table")
}

func TestTLSFixture(t *testing.T) {
	config, caPEM, certPEM, keyPEM, err := newTLSFixture("127.0.0.2", []string{"fixture-s1-r1"})
	require.NoError(t, err)
	require.NotEmpty(t, caPEM)
	require.Equal(t, uint16(tls.VersionTLS12), config.MinVersion)
	require.False(t, config.InsecureSkipVerify)
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	verifyCertificate(t, certificate, config, "127.0.0.2")
	verifyCertificate(t, certificate, config, "localhost")
	verifyCertificate(t, certificate, config, "fixture-s1-r1")
}
