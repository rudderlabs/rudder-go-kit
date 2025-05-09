package pulsar

import (
	"context"
	"net/http"
	"testing"
	"time"

	pulsarclient "github.com/apache/pulsar-client-go/pulsar"
	pulsarlog "github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

func TestPulsar(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	pulsarContainer, err := Setup(pool, t)
	require.NoError(t, err)

	res, err := http.Head(pulsarContainer.AdminURL + "/admin/v2/namespaces/public/default")
	defer func() { httputil.CloseResponse(res) }()
	require.NoError(t, err)

	client, err := pulsarclient.NewClient(pulsarclient.ClientOptions{
		URL:               pulsarContainer.URL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		Logger:            pulsarlog.DefaultNopLogger(),
	})
	require.NoError(t, err)
	defer client.Close()

	const topic = "my-test"

	consumer, err := client.Subscribe(pulsarclient.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: "my-sub",
		Type:             pulsarclient.Exclusive,
	})
	require.NoError(t, err)
	defer consumer.Close()

	producer, err := client.CreateProducer(pulsarclient.ProducerOptions{
		Topic: topic,
	})
	require.NoError(t, err)
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msgID, err := producer.Send(ctx, &pulsarclient.ProducerMessage{
		Key:     "foo",
		Payload: []byte("bar"),
	})
	require.NoError(t, err)
	require.NotNil(t, msgID)

	msg, err := consumer.Receive(ctx)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "foo", msg.Key())
	require.Equal(t, "bar", string(msg.Payload()))
}
