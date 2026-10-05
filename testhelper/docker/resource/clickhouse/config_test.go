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
	WithCluster(2, 2)(&config)
	require.NoError(t, config.validate())
	for _, option := range []Opt{WithCluster(0, 0), WithCluster(0, 2), WithCluster(2, 0), WithCluster(-1, 1), WithMemory(-1), WithUser(""), WithDatabase(""), WithConfig("<broken>"), WithUsersConfig("<users/>"), WithEnv("CLICKHOUSE_USER=other"), WithImage("not an image")} {
		invalid := defaultConfig()
		option(&invalid)
		require.Error(t, invalid.validate())
	}
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
	for _, image := range []string{DefaultImage, "localhost:5000/clickhouse:test", "clickhouse/clickhouse-server@sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893", "clickhouse/clickhouse-server"} {
		repository, tag, err := splitImage(image)
		require.NoError(t, err)
		require.NotEmpty(t, repository)
		require.NotEmpty(t, tag)
		if strings.Contains(image, "@") {
			require.Contains(t, repository+":"+tag, "@sha256:")
		}
	}
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
