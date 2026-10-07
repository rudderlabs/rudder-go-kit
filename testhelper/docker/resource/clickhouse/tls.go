package clickhouse

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// newTLSFixture issues a server certificate signed by ca, or by a new throwaway CA when ca is nil.
// It returns the client configuration, the CA certificate, the server certificate and the server key.
func newTLSFixture(ca *tls.Certificate, bindIP string, hostnames []string, ipSANs bool) (*tls.Config, []byte, []byte, []byte, error) {
	now := time.Now()
	if ca == nil {
		var err error
		if ca, err = newCertificateAuthority(now); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// A nil SerialNumber gets a random serial, so fixtures that share a caller CA do not reuse one.
	server := &x509.Certificate{
		Subject:   pkix.Name{CommonName: "localhost"},
		DNSNames:  append([]string{"localhost"}, hostnames...),
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(7 * 24 * time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if ipSANs {
		server.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
		if address := net.ParseIP(bindIP); address != nil && !address.IsUnspecified() {
			server.IPAddresses = append(server.IPAddresses, address)
		}
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, server, ca.Leaf, &serverKey.PublicKey, ca.PrivateKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	leaf, err := x509.ParseCertificate(serverDER)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca.Leaf)
	// A caller CA that is expired, limited to other key usages or name-constrained signs a certificate that
	// every client rejects, so Setup fails here instead of after pool.MaxWait.
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "localhost", CurrentTime: now}); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("verifying the server certificate against the ClickHouse certificate authority: %w", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	config := &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	if !ipSANs {
		// A client dials an IP address, so it verifies the DNS name instead.
		config.ServerName = "localhost"
	}
	return config, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Leaf.Raw}),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), nil
}

func newCertificateAuthority(now time.Time) (*tls.Certificate, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "rudder-go-kit ClickHouse test CA"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(7 * 24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	leaf, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{Certificate: [][]byte{caDER}, PrivateKey: caKey, Leaf: leaf}, nil
}
