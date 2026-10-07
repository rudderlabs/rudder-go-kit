package clickhouse

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// WithCertificateAuthority signs the server certificate with the caller CA instead of a throwaway CA per Setup.
// CAPEM and TLSConfig then return that CA, so fixtures that share one CA share trust.
// It requires WithTLS. certPEM must hold one CA certificate with the certSign key usage, valid now, that may
// issue serverAuth certificates for localhost. The server key is readable by the container user, so use a
// test-only CA.
func WithCertificateAuthority(certPEM, keyPEM []byte) Opt {
	return func(config *Config) { config.CACertPEM, config.CAKeyPEM = certPEM, keyPEM }
}

// WithPlainHTTPPort also publishes the plain HTTP port 8123 in TLS mode and sets Node.PlainHTTPPort.
// The plain port carries the credentials in clear text.
// Without TLS the plain port is published anyway, so the option changes nothing.
func WithPlainHTTPPort() Opt {
	return func(config *Config) { config.PublishPlainHTTPPort = true }
}

// WithoutIPSANs issues the server certificate for DNS names only. TLSConfig then verifies the name localhost,
// and a client that verifies the dialed IP address fails. It requires WithTLS.
func WithoutIPSANs() Opt {
	return func(config *Config) { config.WithoutIPSANs = true }
}

func (config Config) validateTLSOptions() error {
	if config.CACertPEM == nil && config.CAKeyPEM == nil && !config.WithoutIPSANs {
		return nil
	}
	if !config.TLS {
		return fmt.Errorf("WithCertificateAuthority and WithoutIPSANs require WithTLS")
	}
	if config.CACertPEM == nil && config.CAKeyPEM == nil {
		return nil
	}
	_, err := parseCertificateAuthority(config.CACertPEM, config.CAKeyPEM)
	return err
}

func parseCertificateAuthority(certPEM, keyPEM []byte) (*tls.Certificate, error) {
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing ClickHouse certificate authority: %w", err)
	}
	if len(pair.Certificate) > 1 {
		return nil, fmt.Errorf("ClickHouse certificate authority must be one certificate, not a chain")
	}
	if !pair.Leaf.IsCA || (pair.Leaf.KeyUsage != 0 && pair.Leaf.KeyUsage&x509.KeyUsageCertSign == 0) {
		return nil, fmt.Errorf("ClickHouse certificate authority must be a CA certificate with the certSign key usage")
	}
	return &pair, nil
}
