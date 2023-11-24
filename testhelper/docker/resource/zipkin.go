package resource

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper"
)

const zipkinPort = "9411"

type ZipkinResource struct {
	Port string
}

func SetupZipkin(pool *dockertest.Pool, d cleaner) (*ZipkinResource, error) {
	freePort, err := testhelper.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %w", err)
	}

	zipkin, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "openzipkin/zipkin",
		ExposedPorts: []string{zipkinPort},
		PortBindings: map[docker.Port][]docker.PortBinding{
			zipkinPort + "/tcp": {{HostPort: strconv.Itoa(freePort)}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start zipkin: %w", err)
	}
	d.Cleanup(func() {
		if err := pool.Purge(zipkin); err != nil {
			d.Log("Could not purge zipkin resource:", err)
		}
	})

	zipkinHealthURL := "http://localhost:" + strconv.Itoa(freePort) + "/health"
	healthReq, err := http.NewRequest(http.MethodGet, zipkinHealthURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create zipkin health request: %w", err)
	}

	err = pool.Retry(func() error {
		resp, err := http.DefaultClient.Do(healthReq)
		if err != nil {
			return fmt.Errorf("failed to get zipkin health: %w", err)
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("zipkin health returned status code %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wait for zipkin to be ready: %w", err)
	}

	return &ZipkinResource{
		Port: zipkin.GetPort(zipkinPort + "/tcp"),
	}, nil
}
