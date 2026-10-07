package clickhouse

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
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

// customCA returns a self-signed CA as PEM after change edits the template of a valid CA.
func customCA(t *testing.T, change func(*x509.Certificate)) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "custom test CA"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	change(template)
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
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
	signOnlyPEM, signOnlyKeyPEM := customCA(t, func(ca *x509.Certificate) { ca.IsCA, ca.BasicConstraintsValid = false, false })
	caOnlyPEM, caOnlyKeyPEM := customCA(t, func(ca *x509.Certificate) { ca.KeyUsage = x509.KeyUsageDigitalSignature })
	rootPEM, _ := testCA(t)

	config := defaultConfig()
	require.Nil(t, config.CACertPEM)
	require.False(t, config.PublishPlainHTTPPort)
	require.False(t, config.WithoutIPSANs)
	WithCertificateAuthority(certPEM, keyPEM)(&config)
	WithPlainHTTPPort()(&config)
	WithoutIPSANs()(&config)
	require.Equal(t, certPEM, config.CACertPEM)
	require.Equal(t, keyPEM, config.CAKeyPEM)
	require.True(t, config.PublishPlainHTTPPort)
	require.True(t, config.WithoutIPSANs)

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
		"certSign without CA flag": {[]Opt{WithTLS(), WithCertificateAuthority(signOnlyPEM, signOnlyKeyPEM)}, "must be a CA certificate"},
		"CA flag without certSign": {[]Opt{WithTLS(), WithCertificateAuthority(caOnlyPEM, caOnlyKeyPEM)}, "must be a CA certificate"},
		"CA followed by a root":    {[]Opt{WithTLS(), WithCertificateAuthority(append(append([]byte{}, certPEM...), rootPEM...), keyPEM)}, "must be one certificate"},
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

func TestTLSFixtureRejectsCallerCAThatClientsReject(t *testing.T) {
	for name, change := range map[string]func(*x509.Certificate){
		"expired": func(ca *x509.Certificate) {
			ca.NotBefore, ca.NotAfter = time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour)
		},
		"not yet valid": func(ca *x509.Certificate) {
			ca.NotBefore, ca.NotAfter = time.Now().Add(time.Hour), time.Now().Add(2*time.Hour)
		},
		"client auth only": func(ca *x509.Certificate) { ca.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth} },
		"name constrained": func(ca *x509.Certificate) {
			ca.PermittedDNSDomainsCritical, ca.PermittedDNSDomains = true, []string{"example.com"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			certPEM, keyPEM := customCA(t, change)
			ca, err := parseCertificateAuthority(certPEM, keyPEM)
			require.NoError(t, err, "the CA itself parses")
			_, _, _, _, err = newTLSFixture(ca, "", []string{"fixture-s1-r1"}, true)
			require.ErrorContains(t, err, "verifying the server certificate against the ClickHouse certificate authority")
		})
	}
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
		got, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
			t.Fatal("a bound container needs no inspect")
			return nil, errors.New("unexpected inspect")
		}, started, ports, time.Second)
		require.NoError(t, err)
		require.Same(t, started, got)
	})

	t.Run("bindings appear later", func(t *testing.T) {
		responses := []*docker.Container{nil, boundContainer("8443", "9440"), {ID: "fixture"}, boundContainer(ports...)}
		var ids []string
		got, err := waitForPortBindings(context.Background(), func(_ context.Context, id string) (*docker.Container, error) {
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
		_, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
			inspects++
			return &docker.Container{ID: "fixture"}, nil
		}, &docker.Container{ID: "fixture"}, ports, time.Second)
		require.ErrorContains(t, err, "reading ClickHouse node port bindings")
		require.ErrorContains(t, err, "no host binding for every port of container fixture")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Positive(t, inspects)
		require.LessOrEqual(t, inspects, 5, "the backoff doubles between inspects")
	})

	for _, status := range []string{"exited", "dead"} {
		t.Run(status+" container stops the wait", func(t *testing.T) {
			stopped := &docker.Container{ID: "fixture", State: docker.State{Status: status, ExitCode: 137, OOMKilled: true}}
			start := time.Now()
			_, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
				return stopped, nil
			}, &docker.Container{ID: "fixture", State: docker.State{Status: "running", Running: true}}, ports, 10*time.Second)
			require.ErrorContains(t, err, "container fixture is "+status+" (exit code 137, OOM killed: true)")
			require.Less(t, time.Since(start), 5*time.Second)
		})
	}

	t.Run("removed container stops the wait", func(t *testing.T) {
		inspects := 0
		_, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
			inspects++
			return nil, &docker.NoSuchContainer{ID: "fixture"}
		}, &docker.Container{ID: "fixture"}, ports, 10*time.Second)
		var missing *docker.NoSuchContainer
		require.ErrorAs(t, err, &missing)
		require.Equal(t, 1, inspects)
	})

	t.Run("inspect receives the deadline", func(t *testing.T) {
		_, err := waitForPortBindings(context.Background(), func(ctx context.Context, _ string) (*docker.Container, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}, &docker.Container{ID: "fixture"}, ports, 300*time.Millisecond)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("deadline reports the last inspect error", func(t *testing.T) {
		_, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
			return nil, errors.New("no such container: fixture")
		}, &docker.Container{ID: "fixture"}, ports, 300*time.Millisecond)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.ErrorContains(t, err, "no such container: fixture")
	})

	t.Run("a later successful inspect clears the inspect error", func(t *testing.T) {
		calls := 0
		_, err := waitForPortBindings(context.Background(), func(context.Context, string) (*docker.Container, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("transient inspect failure")
			}
			return &docker.Container{ID: "fixture"}, nil
		}, &docker.Container{ID: "fixture"}, ports, 300*time.Millisecond)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.NotContains(t, err.Error(), "transient inspect failure")
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

// queryStatus runs SELECT currentUser() over plain HTTP and returns the status code.
func queryStatus(t *testing.T, base, user, password string) int {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, base+"/?query=SELECT%20currentUser()", nil)
	require.NoError(t, err)
	if user != "" {
		request.SetBasicAuth(user, password)
	}
	response, err := httpClient(t, nil).Do(request)
	require.NoError(t, err)
	defer func() { httputil.CloseResponse(response) }()
	return response.StatusCode
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
		plain := "http://" + net.JoinHostPort(result.Host, result.PlainHTTPPort)
		require.NoError(t, getPing(t, httpClient(t, nil), plain+"/ping"))
		require.Equal(t, http.StatusOK, queryStatus(t, plain, result.User, result.Password), "the administrator authenticates over the plain port")
		require.Equal(t, http.StatusUnauthorized, queryStatus(t, plain, "", ""), "the plain port requires credentials")
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
