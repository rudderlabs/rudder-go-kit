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
		tag: "3.2.4",
	}
	for _, opt := range opts {
		opt(c)
	}

	var networkID string
	if c.network != nil {
		networkID = c.network.ID
	}
	portBindings, err := internal.PortBindings([]string{"6650", "8080"})
	if err != nil {
		return nil, err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "apachepulsar/pulsar",
		Tag:          c.tag,
		Env:          []string{},
		PortBindings: portBindings,
		Cmd:          []string{"bin/pulsar", "standalone"},
		NetworkID:    networkID,
	})
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	url := fmt.Sprintf("pulsar://%s:%s", container.GetBoundIP("6650/tcp"), container.GetPort("6650/tcp"))
	adminURL := fmt.Sprintf("http://%s:%s", container.GetBoundIP("8080/tcp"), container.GetPort("8080/tcp"))

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
