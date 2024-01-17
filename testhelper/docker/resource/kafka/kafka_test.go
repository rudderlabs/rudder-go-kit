package kafka

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	"github.com/linkedin/goavro/v2"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
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

func TestAvroSchemaRegistry(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	container, err := Setup(pool, t, WithBrokers(1), WithSchemaRegistry())
	require.NoError(t, err)

	c, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers":  fmt.Sprintf("localhost:%s", container.Ports[0]),
		"group.id":           "group-1",
		"session.timeout.ms": 6000,
		"auto.offset.reset":  "earliest",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	topic := "my-topic"
	err = c.SubscribeTopics([]string{topic}, nil)
	require.NoError(t, err)

	type User struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	consumeUserMsg := func(t *testing.T, deser *avro.GenericDeserializer) {
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-timeout:
				t.Fatal("Timed out waiting for expected message")
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *confluent.Message:
					value := User{}
					err = deser.DeserializeInto(*e.TopicPartition.Topic, e.Value, &value)
					require.NoErrorf(t, err, "Failed to deserialize payload: %s", err)
					require.Equal(t, User{FirstName: "John", LastName: "Doe"}, value)
					return
				case confluent.Error:
					t.Logf("Kafka Confluent Error: %v: %v", e.Code(), e)
				default:
					t.Logf("Ignoring consumer entry: %+v", e)
				}
			}
		}
	}

	// Registering schemas and setting up writer
	schemaRegistryClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(container.SchemaRegistryURL))
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	path := func(file string) string { return filepath.Join(cwd, "testdata", "avro", file) }
	_, schemaID1 := registerSchema(t, "user1", path("user1.avsc"), schemaRegistryClient)
	userSchema2, schemaID2 := registerSchema(t, "user2", path("user2.avsc"), schemaRegistryClient)
	t.Logf("Schemas IDs: %d, %d", schemaID1, schemaID2)

	rawMessage := json.RawMessage(`{
		"first_name": "John",
		"last_name": "Doe"
	}`)
	avroMessage := serializeAvroMessage(t, schemaID2, userSchema2, rawMessage)

	w := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:" + container.Ports[0]),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	t.Cleanup(func() { _ = w.Close() })

	t.Log("Writing message")
	require.Eventually(t, func() bool {
		err := w.WriteMessages(context.Background(),
			kafka.Message{Topic: topic, Key: []byte("123"), Value: avroMessage},
		)
		if err != nil {
			t.Logf("failed to write messages: %s", err)
		}
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)

	// Start consuming
	t.Log("Consuming message")
	deser, err := avro.NewGenericDeserializer(schemaRegistryClient, serde.ValueSerde, avro.NewDeserializerConfig())
	require.NoError(t, err)
	consumeUserMsg(t, deser)
}

func registerSchema(
	t *testing.T, schemaName, schemaPath string, c schemaregistry.Client,
) (schema string, schemaID int) {
	t.Helper()

	buf, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	si := schemaregistry.SchemaInfo{Schema: string(buf)}
	require.Eventuallyf(t, func() bool {
		schemaID, err = c.Register(schemaName, si, true)
		return err == nil
	}, 30*time.Second, 100*time.Millisecond, "failed to register schema %s: %v", schemaName, err)

	schema = string(buf)
	return
}

func serializeAvroMessage(t *testing.T, schemaID int, schema string, value []byte) []byte {
	t.Helper()

	codec, err := goavro.NewCodec(schema)
	require.NoError(t, err)

	native, _, err := codec.NativeFromTextual(value)
	require.NoError(t, err)

	bin, err := codec.BinaryFromNative(nil, native)
	require.NoError(t, err)

	return addAvroSchemaIDHeader(t, schemaID, bin)
}

func addAvroSchemaIDHeader(t *testing.T, schemaID int, msgBytes []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoErrorf(t, buf.WriteByte(byte(0x0)), "avro header: unable to write magic byte")

	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, uint32(schemaID))
	_, err := buf.Write(idBytes)
	require.NoError(t, err)

	_, err = buf.Write(msgBytes)
	require.NoError(t, err)

	return buf.Bytes()
}
