package client

import (
	"testing"

	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/stretchr/testify/require"
)

func TestScramHashGeneratorString(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tests := []struct {
			generator ScramHashGenerator
			expected  string
		}{
			{ScramPlainText, "plain"},
			{ScramSHA256, "sha256"},
			{ScramSHA512, "sha512"},
		}

		for _, test := range tests {
			require.Equal(t, test.expected, test.generator.String())
		}
	})
	t.Run("panic", func(t *testing.T) {
		require.Panics(t, func() {
			_ = ScramHashGenerator(123).String()
		})
	})
}

func TestScramHashGeneratorFromString(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tests := []struct {
			generator string
			expected  ScramHashGenerator
		}{
			{"plain", ScramPlainText},
			{"sha256", ScramSHA256},
			{"sha512", ScramSHA512},
		}

		for _, test := range tests {
			generator, err := ScramHashGeneratorFromString(test.generator)
			require.NoError(t, err)
			require.Equal(t, test.expected, generator)
		}
	})
	t.Run("error", func(t *testing.T) {
		_, err := ScramHashGeneratorFromString("foo")
		require.Error(t, err)
	})
}

func TestTLS(t *testing.T) {
	t.Run("empty configuration", func(t *testing.T) {
		conf, err := (&TLS{}).build()
		require.Nil(t, conf)
		require.ErrorContains(t, err, "invalid TLS configuration, either provide certificates or skip validation")
	})
	t.Run("can run with system cert pool", func(t *testing.T) {
		conf, err := (&TLS{WithSystemCertPool: true}).build()
		require.NoError(t, err)
		require.NotNil(t, conf)
	})
	t.Run("it fails with invalid ca cert", func(t *testing.T) {
		conf, err := (&TLS{CACertificate: []byte("foo")}).build()
		require.Nil(t, conf)
		require.ErrorContains(t, err, "could not append certs from PEM")
	})
	t.Run("it fails with invalid cert and key", func(t *testing.T) {
		conf, err := (&TLS{Cert: []byte("foo"), Key: []byte("bar"), WithSystemCertPool: true}).build()
		require.Nil(t, conf)
		require.ErrorContains(t, err, "could not get TLS certificate")
	})
}

func TestSASL(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		mechanism, err := (&SASL{
			ScramHashGen: ScramPlainText,
			Username:     "foo",
			Password:     "bar",
		}).build()
		require.NoError(t, err)
		require.NotNil(t, mechanism)
		require.IsType(t, plain.Mechanism{}, mechanism)
	})
	t.Run("sha256", func(t *testing.T) {
		mechanism, err := (&SASL{
			ScramHashGen: ScramSHA256,
			Username:     "foo",
			Password:     "bar",
		}).build()
		require.NoError(t, err)
		require.NotNil(t, mechanism)
		require.Equal(t, "SCRAM-SHA-256", mechanism.Name())
	})
	t.Run("sha512", func(t *testing.T) {
		mechanism, err := (&SASL{
			ScramHashGen: ScramSHA512,
			Username:     "foo",
			Password:     "bar",
		}).build()
		require.NoError(t, err)
		require.NotNil(t, mechanism)
		require.Equal(t, "SCRAM-SHA-512", mechanism.Name())
	})
	t.Run("error", func(t *testing.T) {
		mechanism, err := (&SASL{
			ScramHashGen: ScramHashGenerator(123),
			Username:     "foo",
			Password:     "bar",
		}).build()
		require.Nil(t, mechanism)
		require.Error(t, err)
	})
}
