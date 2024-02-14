package transformer

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

type Resource struct {
	TransformerURL string
	Port           string
}

type config struct {
	Repository   string
	Tag          string
	ExposedPorts []string
	Env          []string
}

func WithConfigBackendURL(url string) func(*config) {
	return func(conf *config) {
		conf.Env = []string{fmt.Sprintf("CONFIG_BACKEND_URL=%s", url)}
	}
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...func(conf *config)) (*Resource, error) {
	// Set Rudder Transformer
	// pulls an image first to make sure we don't have an old cached version locally,
	// then it creates a container based on it and runs it
	conf := &config{
		Repository:   "rudderstack/rudder-transformer",
		Tag:          "latest",
		ExposedPorts: []string{"9090"},
		Env: []string{
			"CONFIG_BACKEND_URL=https://api.rudderstack.com",
		},
	}

	for _, opt := range opts {
		opt(conf)
	}

	err := pool.Client.PullImage(docker.PullImageOptions{
		Repository: conf.Repository,
		Tag:        conf.Tag,
	}, docker.AuthConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	transformerContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   conf.Repository,
		Tag:          conf.Tag,
		ExposedPorts: conf.ExposedPorts,
		Env:          conf.Env,
	})
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(transformerContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	transformerResource := &Resource{
		TransformerURL: fmt.Sprintf("http://localhost:%s", transformerContainer.GetPort("9090/tcp")),
		Port:           transformerContainer.GetPort("9090/tcp"),
	}

	err = pool.Retry(func() (err error) {
		resp, err := http.Get(transformerResource.TransformerURL + "/health")
		if err != nil {
			return err
		}
		defer func() { httputil.CloseResponse(resp) }()
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transformerResource, nil
}