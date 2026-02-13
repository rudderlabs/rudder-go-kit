package testhelper

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
	dt "github.com/rudderlabs/rudder-go-kit/testhelper/docker"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const healthPort = "13133"

type StartOTelCollectorOpt func(*startOTelCollectorConf)

// WithStartCollectorPort allows to specify the port on which the collector will be listening for gRPC requests.
func WithStartCollectorPort(port int) StartOTelCollectorOpt {
	return func(c *startOTelCollectorConf) {
		c.port = port
	}
}

func StartOTelCollector(t testing.TB, metricsPort, configPath string, opts ...StartOTelCollectorOpt) (
	container *docker.Container,
	grpcEndpoint string,
) {
	t.Helper()

	conf := &startOTelCollectorConf{}
	for _, opt := range opts {
		opt(conf)
	}

	if conf.port == 0 {
		var err error
		conf.port, err = testhelper.GetFreePort()
		require.NoError(t, err)
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	collector, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   registry.ImagePath("otel/opentelemetry-collector"),
		Tag:          "0.115.0",
		ExposedPorts: []string{healthPort + "/tcp", metricsPort + "/tcp", "4317/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"4317/tcp":                        {{HostPort: strconv.Itoa(conf.port)}},
			healthPort + "/tcp":               {{}},
			docker.Port(metricsPort + "/tcp"): {{}},
		},
		Mounts: []string{configPath + ":/etc/otelcol/config.yaml:z"},
		Auth:   registry.AuthConfiguration(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pool.Purge(collector); err != nil {
			t.Logf("Could not purge resource: %v", err)
		}
	})

	healthEndpoint := fmt.Sprintf("http://%s:%d", collector.GetBoundIP(healthPort), dt.GetHostPort(t, healthPort, collector.Container))
	healthy := false
	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(healthEndpoint) //nolint:gosec,bodyclose // this is a test helper, body closed below
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}
		statusOK := resp.StatusCode == http.StatusOK
		httputil.CloseResponse(resp)
		if statusOK {
			healthy = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !healthy {
		// Log container output to help debug why it's not healthy
		var stdout, stderr bytes.Buffer
		logsErr := pool.Client.Logs(docker.LogsOptions{
			Container:    collector.Container.ID,
			OutputStream: &stdout,
			ErrorStream:  &stderr,
			Stdout:       true,
			Stderr:       true,
		})
		if logsErr != nil {
			t.Logf("Failed to get container logs: %v", logsErr)
		} else {
			t.Logf("Container stdout:\n%s", stdout.String())
			t.Logf("Container stderr:\n%s", stderr.String())
		}
		require.Fail(t, "Collector was not ready on health port", "health endpoint: %s, last error: %v", healthEndpoint, lastErr)
	}

	t.Log("Container is healthy")

	return collector.Container, collector.GetBoundIP("4317/tcp") + ":" + strconv.Itoa(conf.port)
}

type startOTelCollectorConf struct {
	port int
}
