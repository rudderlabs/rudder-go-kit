package kafka

import (
	"context"
	"testing"
	"time"

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
