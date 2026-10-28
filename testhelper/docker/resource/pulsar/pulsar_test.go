package pulsar

import (
	"bytes"
	"context"
	"net/http"
	"strings"
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

func TestTestTuning(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	t.Run("tuned by default", func(t *testing.T) {
		pulsarContainer, err := Setup(pool, t)
		require.NoError(t, err)

		args := jvmArgs(t, pulsarContainer)
		require.Contains(t, args, "-Xmx512m")
		require.Contains(t, args, "-XX:+UseG1GC")
		require.NotContains(t, args, "-XX:+AlwaysPreTouch")
	})

	t.Run("caller variables replace the tuning", func(t *testing.T) {
		pulsarContainer, err := Setup(pool, t, WithEnv("PULSAR_MEM=-Xms256m -Xmx768m"))
		require.NoError(t, err)

		args := jvmArgs(t, pulsarContainer)
		require.Contains(t, args, "-Xmx768m")
		require.NotContains(t, args, "-Xmx512m")
		require.Contains(t, args, "-XX:+UseG1GC")
	})

	t.Run("stock settings without test tuning", func(t *testing.T) {
		pulsarContainer, err := Setup(pool, t, WithoutTestTuning())
		require.NoError(t, err)

		args := jvmArgs(t, pulsarContainer)
		require.Contains(t, args, "-Xmx2g")
		require.Contains(t, args, "-XX:+AlwaysPreTouch")
	})
}

// jvmArgs returns the command line of the Pulsar JVM, which bin/pulsar execs as the container's PID 1.
func jvmArgs(t *testing.T, r *Resource) []string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code, err := r.container.Exec([]string{"cat", "/proc/1/cmdline"}, dockertest.ExecOptions{
		StdOut: &stdout, StdErr: &stderr,
	})
	require.NoError(t, err)
	require.Zerof(t, code, "reading the JVM command line: %s", stderr.String())
	return strings.Split(strings.TrimRight(stdout.String(), "\x00"), "\x00")
}
