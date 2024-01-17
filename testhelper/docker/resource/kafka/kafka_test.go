package kafka

import (
	"context"
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
	"time"

	dc "github.com/ory/dockertest/v3/docker"

	"github.com/segmentio/kafka-go/sasl/plain"

	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"

	"github.com/ory/dockertest/v3"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestResource(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	res, err := Setup(pool, t,
		WithBrokers(3),
	)
	require.NoError(t, err)

	var (
		ctx     = context.Background()
		topic   = "my-topic"
		brokers = []string{
			"localhost:" + res.Ports[0],
			"localhost:" + res.Ports[1],
			"localhost:" + res.Ports[2],
		}
	)

	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	t.Cleanup(func() { _ = w.Close() })

	require.Eventually(t, func() bool {
		err := w.WriteMessages(ctx,
			kafka.Message{Topic: topic, Key: []byte("one"), Value: []byte("one!")},
			kafka.Message{Topic: topic, Key: []byte("two"), Value: []byte("two!")},
			kafka.Message{Topic: topic, Key: []byte("three"), Value: []byte("three!")},
		)
		if err != nil {
			t.Logf("failed to write messages: %s", err)
		}
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)
}

func TestWithSASL(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	network, err := pool.Client.CreateNetwork(dc.CreateNetworkOptions{Name: "test_network"})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pool.Client.RemoveNetwork(network.ID); err != nil {
			t.Logf("Error while removing Docker network: %v", err)
		}
	})

	path, err := os.Getwd()
	require.NoError(t, err)

	saslConfiguration := SASLConfig{
		BrokerUser: User{Username: "kafka1", Password: "password"},
		Users: []User{
			{Username: "client1", Password: "password"},
		},
		CertificatePassword: "password",
		KeyStorePath:        filepath.Join(path, "testdata", "keystore", "kafka.keystore.jks"),
		TrustStorePath:      filepath.Join(path, "testdata", "truststore", "kafka.truststore.jks"),
	}

	hashTypes := []string{"scramPlainText", "scramSHA256", "scramSHA512"}
	for _, hashType := range hashTypes {
		t.Run(hashType, func(t *testing.T) {
			var mechanism sasl.Mechanism
			containerOptions := []Option{
				WithBrokers(1),
				WithNetwork(network),
			}

			switch hashType {
			case "scramPlainText":
				mechanism = plain.Mechanism{
					Username: saslConfiguration.Users[0].Username,
					Password: saslConfiguration.Users[0].Password,
				}
				containerOptions = append(containerOptions, WithSASLPlain(&saslConfiguration))
			case "scramSHA256":
				mechanism, err = scram.Mechanism(
					scram.SHA256, saslConfiguration.Users[0].Username, saslConfiguration.Users[0].Password,
				)
				require.NoError(t, err)
				containerOptions = append(containerOptions, WithSASLScramSHA256(&saslConfiguration))
			case "scramSHA512":
				mechanism, err = scram.Mechanism(
					scram.SHA512, saslConfiguration.Users[0].Username, saslConfiguration.Users[0].Password,
				)
				require.NoError(t, err)
				containerOptions = append(containerOptions, WithSASLScramSHA512(&saslConfiguration))
			}
			container, err := Setup(pool, t, containerOptions...)
			require.NoError(t, err)

			w := kafka.Writer{
				Addr:     kafka.TCP("localhost:" + container.Ports[0]),
				Balancer: &kafka.Hash{},
				Transport: &kafka.Transport{
					SASL: mechanism,
					TLS: &tls.Config{ // skipcq: GSC-G402
						MinVersion:         tls.VersionTLS11,
						MaxVersion:         tls.VersionTLS12,
						InsecureSkipVerify: true,
					},
				},
				AllowAutoTopicCreation: true,
			}
			t.Cleanup(func() { _ = w.Close() })

			require.Eventually(t, func() bool {
				err := w.WriteMessages(context.Background(),
					kafka.Message{Topic: "my-topic", Key: []byte("one"), Value: []byte("one!")},
					kafka.Message{Topic: "my-topic", Key: []byte("two"), Value: []byte("two!")},
					kafka.Message{Topic: "my-topic", Key: []byte("three"), Value: []byte("three!")},
				)
				if err != nil {
					t.Logf("failed to write messages: %s", err)
				}
				return err == nil
			}, 30*time.Second, 500*time.Millisecond)
		})
	}
}
