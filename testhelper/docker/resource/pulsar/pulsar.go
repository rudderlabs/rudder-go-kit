package pulsar

import (
	"bytes"
	"fmt"

	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

type Resource struct {
	URL      string
	AdminURL string
	// URLInNetwork is the URL accessible from the provided Docker network (if any).
	URLInNetwork string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	c := &config{
		tag: "3.3.6",
	}
	for _, opt := range opts {
		opt(c)
	}

	var networkID string
	if c.network != nil {
		networkID = c.network.ID
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "hub.dev-rudder.rudderlabs.com/apachepulsar/pulsar",
		Tag:          c.tag,
		Env:          []string{},
		ExposedPorts: []string{"6650/tcp", "8080/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{"6650", "8080"}),
		Cmd:          []string{"bin/pulsar", "standalone"},
		NetworkID:    networkID,
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot run pulsar container: %w", err)
	}

	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	url := fmt.Sprintf("pulsar://127.0.0.1:%s", container.GetPort("6650/tcp"))
	adminURL := fmt.Sprintf("http://127.0.0.1:%s", container.GetPort("8080/tcp"))

	if err := pool.Retry(func() (err error) {
		var w bytes.Buffer
		code, err := container.Exec(
			[]string{
				"sh", "-c", "curl -I http://localhost:8080/admin/v2/namespaces/public/default | grep '200' || exit 1",
			},
			dockertest.ExecOptions{StdOut: &w, StdErr: &w},
		)
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("pulsar healthcheck failed")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var urlInNetwork string
	if c.network != nil {
		urlInNetwork = "pulsar://" + container.GetIPInNetwork(&dockertest.Network{Network: c.network}) + ":6650"
	}

	return &Resource{
		URL:          url,
		AdminURL:     adminURL,
		URLInNetwork: urlInNetwork,
	}, nil
}
