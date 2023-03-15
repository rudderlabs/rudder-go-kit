package resource

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/pulsar"
)

type PulsarResource struct {
	URL      string
	AdminURL string
}

func SetupPulsar(pool *dockertest.Pool, d cleaner, opts ...pulsar.Opt) (*PulsarResource, error) {
	c := &pulsar.Config{
		Tag: "2.11.0",
	}
	for _, opt := range opts {
		opt(c)
	}
	cmd := []string{"bin/pulsar", "standalone"}

	repository := "apachepulsar/pulsar"
	tag := c.Tag
	if runtime.GOARCH == "arm64" { // TODO: use original image when multi-arch images are supported by pulsar
		repository = "atzoum/pulsar"
		tag = "latest"
	}
	pulsarContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   repository,
		Tag:          tag,
		Env:          []string{},
		ExposedPorts: []string{"6650", "8080"},
		Cmd:          cmd,
	})
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(pulsarContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	url := fmt.Sprintf("pulsar://localhost:%s", pulsarContainer.GetPort("6650/tcp"))
	adminURL := fmt.Sprintf("http://localhost:%s", pulsarContainer.GetPort("8080/tcp"))

	if err := pool.Retry(func() (err error) {
		var w bytes.Buffer
		code, err := pulsarContainer.Exec([]string{"bin/pulsar-admin", "brokers", "healthcheck"}, dockertest.ExecOptions{StdOut: &w, StdErr: &w})
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("pulsar healthcheck failed")
		}
		out := strings.ReplaceAll(w.String(), "\n", "")
		if !strings.Contains(out, "ok") {
			return fmt.Errorf("pulsar healthcheck failed")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &PulsarResource{
		URL:      url,
		AdminURL: adminURL,
	}, nil
}
