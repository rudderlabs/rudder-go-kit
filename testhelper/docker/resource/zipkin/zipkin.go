package zipkin

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const zipkinPort = "9411"

type Resource struct {
	URL string

	pool     *dockertest.Pool
	resource *dockertest.Resource
	purged   bool
	purgedMu sync.Mutex
}

func (z *Resource) Purge() error {
	z.purgedMu.Lock()
	defer z.purgedMu.Unlock()

	if z.purged {
		return nil
	}

	if err := z.pool.Purge(z.resource); err != nil {
		return err
	}

	z.purged = true

	return nil
}

func Setup(pool *dockertest.Pool, d resource.Cleaner) (*Resource, error) {
	zipkin, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   registry.ImagePath("openzipkin/zipkin"),
		ExposedPorts: []string{zipkinPort + "/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{zipkinPort}),
		Auth:         registry.AuthConfiguration(),
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start zipkin: %w", err)
	}

	res := &Resource{
		pool:     pool,
		resource: zipkin,
		URL:      fmt.Sprintf("http://%s:%s", zipkin.GetBoundIP(zipkinPort+"/tcp"), zipkin.GetPort(zipkinPort+"/tcp")),
	}

	if zipkin.GetBoundIP(zipkinPort+"/tcp") == "" {
		return nil, fmt.Errorf("failed to get zipkin bound ip")
	}

	d.Cleanup(func() {
		if err := res.Purge(); err != nil {
			d.Log("Could not purge zipkin resource:", err)
		}
	})

	healthReq, err := http.NewRequest(http.MethodGet, res.URL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create zipkin health request: %w", err)
	}

	err = pool.Retry(func() error {
		resp, err := http.DefaultClient.Do(healthReq)
		if err != nil {
			return fmt.Errorf("failed to get zipkin health: %w", err)
		}

		defer func() { httputil.CloseResponse(resp) }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("zipkin health returned status code %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wait for zipkin to be ready: %w", err)
	}

	return res, nil
}
