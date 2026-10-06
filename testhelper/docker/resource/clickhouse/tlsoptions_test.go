package clickhouse

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

// testCA returns a new CA as PEM, the form a caller passes to WithCertificateAuthority.
func testCA(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	ca, err := newCertificateAuthority(time.Now())
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(ca.PrivateKey)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Leaf.Raw}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
}

func trustOnly(t *testing.T, caPEM []byte) *tls.Config {
	t.Helper()
	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(caPEM))
	return &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
}

func TestTLSOptionsConfig(t *testing.T) {
	certPEM, keyPEM := testCA(t)
	_, otherKeyPEM := testCA(t)
	_, _, serverCertPEM, serverKeyPEM, err := newTLSFixture(nil, "", nil, true)
	require.NoError(t, err)

	config := defaultConfig()
	require.Nil(t, config.CACertPEM)
	require.False(t, config.PlainHTTPPort)
	require.False(t, config.NoIPSANs)
	WithCertificateAuthority(certPEM, keyPEM)(&config)
	WithPlainHTTPPort()(&config)
	WithoutIPSANs()(&config)
	require.Equal(t, certPEM, config.CACertPEM)
	require.Equal(t, keyPEM, config.CAKeyPEM)
	require.True(t, config.PlainHTTPPort)
	require.True(t, config.NoIPSANs)

	valid := map[string][]Opt{
		"caller CA":            {WithTLS(), WithCertificateAuthority(certPEM, keyPEM)},
		"plain port with TLS":  {WithTLS(), WithPlainHTTPPort()},
		"plain port alone":     {WithPlainHTTPPort()},
		"no IP SANs with TLS":  {WithTLS(), WithoutIPSANs()},
		"all TLS options used": {WithTLS(), WithCertificateAuthority(certPEM, keyPEM), WithPlainHTTPPort(), WithoutIPSANs()},
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

	invalid := map[string]struct {
		options []Opt
		err     string
	}{
		"caller CA without TLS":    {[]Opt{WithCertificateAuthority(certPEM, keyPEM)}, "require WithTLS"},
		"no IP SANs without TLS":   {[]Opt{WithoutIPSANs()}, "require WithTLS"},
		"CA without key":           {[]Opt{WithTLS(), WithCertificateAuthority(certPEM, nil)}, "parsing ClickHouse certificate authority"},
		"key without CA":           {[]Opt{WithTLS(), WithCertificateAuthority(nil, keyPEM)}, "parsing ClickHouse certificate authority"},
		"CA with another key":      {[]Opt{WithTLS(), WithCertificateAuthority(certPEM, otherKeyPEM)}, "parsing ClickHouse certificate authority"},
		"CA that is not PEM":       {[]Opt{WithTLS(), WithCertificateAuthority([]byte("cert"), []byte("key"))}, "parsing ClickHouse certificate authority"},
		"server certificate as CA": {[]Opt{WithTLS(), WithCertificateAuthority(serverCertPEM, serverKeyPEM)}, "must be a CA certificate"},
	}
	for name, tc := range invalid {
		t.Run("invalid "+name, func(t *testing.T) {
			config := defaultConfig()
			for _, option := range tc.options {
				option(&config)
			}
			require.ErrorContains(t, config.validate(), tc.err)
		})
	}
}

func TestTLSFixtureCallerCA(t *testing.T) {
	certPEM, keyPEM := testCA(t)
	ca, err := parseCertificateAuthority(certPEM, keyPEM)
	require.NoError(t, err)
	config, caPEM, firstPEM, firstKeyPEM, err := newTLSFixture(ca, "", []string{"fixture-s1-r1"}, true)
	require.NoError(t, err)
	require.Equal(t, certPEM, caPEM, "CAPEM returns the caller CA")
	_, _, secondPEM, secondKeyPEM, err := newTLSFixture(ca, "", []string{"fixture-s1-r1"}, true)
	require.NoError(t, err)

	callerTrust := trustOnly(t, certPEM)
	otherCertPEM, _ := testCA(t)
	var serials []string
	for _, pair := range [][2][]byte{{firstPEM, firstKeyPEM}, {secondPEM, secondKeyPEM}} {
		certificate, err := tls.X509KeyPair(pair[0], pair[1])
		require.NoError(t, err)
		verifyCertificate(t, certificate, config, "127.0.0.1")
		verifyCertificate(t, certificate, callerTrust, "fixture-s1-r1")
		_, err = certificate.Leaf.Verify(x509.VerifyOptions{Roots: trustOnly(t, otherCertPEM).RootCAs, DNSName: "localhost"})
		var unknown x509.UnknownAuthorityError
		require.ErrorAs(t, err, &unknown)
		serials = append(serials, certificate.Leaf.SerialNumber.String())
	}
	require.NotEqual(t, serials[0], serials[1], "fixtures that share a CA get distinct serial numbers")
}

func TestTLSFixtureWithoutIPSANs(t *testing.T) {
	config, _, certPEM, keyPEM, err := newTLSFixture(nil, "127.0.0.2", []string{"fixture-s1-r1"}, false)
	require.NoError(t, err)
	require.Equal(t, "localhost", config.ServerName, "clients verify the DNS name because the certificate has no IP")
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	require.Empty(t, certificate.Leaf.IPAddresses)
	verifyCertificate(t, certificate, config, "localhost")
	verifyCertificate(t, certificate, config, "fixture-s1-r1")
	for _, address := range []string{"127.0.0.1", "127.0.0.2"} {
		_, err = certificate.Leaf.Verify(x509.VerifyOptions{Roots: config.RootCAs, DNSName: address})
		var hostname x509.HostnameError
		require.ErrorAs(t, err, &hostname, address)
	}

	config, _, _, _, err = newTLSFixture(nil, "127.0.0.2", nil, true)
	require.NoError(t, err)
	require.Empty(t, config.ServerName, "the default configuration verifies the dialed address")
}

func boundContainer(ports ...string) *docker.Container {
	bindings := map[docker.Port][]docker.PortBinding{}
	for _, port := range ports {
		bindings[docker.Port(port+"/tcp")] = []docker.PortBinding{{HostIP: "127.0.0.1", HostPort: "4" + port}}
	}
	return &docker.Container{ID: "fixture", NetworkSettings: &docker.NetworkSettings{Ports: bindings}}
}

func TestWaitForPortBindings(t *testing.T) {
	ports := []string{"8443", "9440", "8123"}

	t.Run("bound at start", func(t *testing.T) {
		started := boundContainer(ports...)
		got, err := waitForPortBindings(context.Background(), func(string) (*docker.Container, error) {
			t.Fatal("a bound container needs no inspect")
			return nil, errors.New("unexpected inspect")
		}, started, ports, time.Second)
		require.NoError(t, err)
		require.Same(t, started, got)
	})

	t.Run("bindings appear later", func(t *testing.T) {
		responses := []*docker.Container{nil, boundContainer("8443", "9440"), {ID: "fixture"}, boundContainer(ports...)}
		var ids []string
		got, err := waitForPortBindings(context.Background(), func(id string) (*docker.Container, error) {
			ids = append(ids, id)
			response := responses[0]
			responses = responses[1:]
			if response == nil {
				return nil, errors.New("transient inspect failure")
			}
			return response, nil
		}, &docker.Container{ID: "fixture", NetworkSettings: &docker.NetworkSettings{}}, ports, 10*time.Second)
		require.NoError(t, err)
		require.Equal(t, "48123", got.NetworkSettings.Ports["8123/tcp"][0].HostPort)
		require.Equal(t, []string{"fixture", "fixture", "fixture", "fixture"}, ids)
	})

	t.Run("empty host port is not a binding", func(t *testing.T) {
		container := boundContainer(ports...)
		container.NetworkSettings.Ports["8123/tcp"] = []docker.PortBinding{{HostIP: "127.0.0.1"}}
		require.False(t, hasPortBindings(container, ports))
	})

	t.Run("deadline", func(t *testing.T) {
		inspects := 0
		_, err := waitForPortBindings(context.Background(), func(string) (*docker.Container, error) {
			inspects++
			return &docker.Container{ID: "fixture"}, nil
		}, &docker.Container{ID: "fixture"}, ports, 300*time.Millisecond)
		require.ErrorContains(t, err, "reading ClickHouse node port bindings")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Positive(t, inspects)
		require.Less(t, inspects, 10, "the backoff grows between inspects")
	})
}

func getPing(t *testing.T, client *http.Client, url string) error {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { httputil.CloseResponse(response) }()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "Ok.\n", string(body))
	return nil
}

func httpClient(t *testing.T, config *tls.Config) *http.Client {
	t.Helper()
	transport := &http.Transport{TLSClientConfig: config}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}

func TestCallerCertificateAuthority(t *testing.T) {
	pool := dockerPool(t)
	certPEM, keyPEM := testCA(t)
	callerTrust := trustOnly(t, certPEM)
	otherCertPEM, _ := testCA(t)
	otherTrust := trustOnly(t, otherCertPEM)
	// Two live fixtures in one process share trust through the caller CA.
	var results []*Resource
	for range 2 {
		result, err := Setup(pool, t, WithTLS(), WithCertificateAuthority(certPEM, keyPEM), WithMemory(1<<30), WithPrintLogsOnError(true))
		require.NoError(t, err)
		results = append(results, result)
	}
	trusted, untrusted := httpClient(t, callerTrust), httpClient(t, otherTrust)
	for _, result := range results {
		require.Equal(t, certPEM, result.CAPEM)
		require.Empty(t, result.PlainHTTPPort, "TLS mode publishes no plain port by default")
		ping := "https://" + net.JoinHostPort(result.Host, result.HTTPPort) + "/ping"
		require.NoError(t, getPing(t, trusted, ping))

		options, err := ch.ParseDSN(result.HTTPDSN())
		require.NoError(t, err)
		options.TLS = callerTrust.Clone()
		db := ch.OpenDB(options)
		var one uint8
		require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
		require.NoError(t, db.Close())

		var unknown x509.UnknownAuthorityError
		require.ErrorAs(t, getPing(t, untrusted, ping), &unknown, "a client that trusts another CA rejects the server certificate")
	}
}

func TestPlainHTTPPort(t *testing.T) {
	pool := dockerPool(t)

	t.Run("tls", func(t *testing.T) {
		result, err := Setup(pool, t, WithTLS(), WithPlainHTTPPort(), WithoutIPSANs(), WithMemory(1<<30), WithPrintLogsOnError(true))
		require.NoError(t, err)
		require.NotEmpty(t, result.PlainHTTPPort)
		require.NotEqual(t, result.HTTPPort, result.PlainHTTPPort)
		require.NoError(t, getPing(t, httpClient(t, nil), "http://"+net.JoinHostPort(result.Host, result.PlainHTTPPort)+"/ping"))
		secure := "https://" + net.JoinHostPort(result.Host, result.HTTPPort) + "/ping"
		require.NoError(t, getPing(t, httpClient(t, result.TLSConfig), secure))

		addressCheck := result.TLSConfig.Clone()
		addressCheck.ServerName = ""
		var hostname x509.HostnameError
		require.ErrorAs(t, getPing(t, httpClient(t, addressCheck), secure), &hostname, "WithoutIPSANs leaves the loopback address off the certificate")

		container, err := pool.Client.InspectContainer(result.ContainerID)
		require.NoError(t, err)
		var published []string
		for port, bindings := range container.NetworkSettings.Ports {
			if len(bindings) > 0 {
				published = append(published, string(port))
			}
		}
		require.ElementsMatch(t, []string{"8123/tcp", "8443/tcp", "9440/tcp"}, published)
	})

	t.Run("plain", func(t *testing.T) {
		result, err := Setup(pool, t, WithPlainHTTPPort(), WithMemory(1<<30), WithPrintLogsOnError(true))
		require.NoError(t, err)
		require.NotEmpty(t, result.PlainHTTPPort)
		require.Equal(t, result.HTTPPort, result.PlainHTTPPort, "without TLS the plain port is the HTTP port")
	})
}
