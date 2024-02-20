package transformer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	kithttptest "github.com/rudderlabs/rudder-go-kit/testhelper/httptest"

	dockerTestHelper "github.com/rudderlabs/rudder-go-kit/testhelper/docker"

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
	repository   string
	tag          string
	exposedPorts []string
	envs         []string
	extraHosts   []string
}

func (c *config) updateBackendConfigURL(url string) {
	found := false
	for i, env := range c.envs {
		if strings.HasPrefix(env, "CONFIG_BACKEND_URL=") {
			found = true
			c.envs[i] = fmt.Sprintf("CONFIG_BACKEND_URL=%s", url)
		}
	}
	if !found {
		c.envs = append(c.envs, fmt.Sprintf("CONFIG_BACKEND_URL=%s", url))
	}
}

// WithTransformations will mock BE config to set transformation for given transformation versionID to transformation function map
//
// - events with transformationVersionID not present in map will not be transformed and transformer will return 404 for those requests
//
// - WithTransformations should not be used with WithConfigBackendURL option
//
// - only javascript transformation functions are supported
//
// e.g.
//
//	WithTransformations(map[string]string{
//				"transform-version-id-1": `export function transformEvent(event, metadata) {
//											event.transformed=true
//											return event;
//										}`,
//			})
func WithTransformations(transformations map[string]string, cleaner resource.Cleaner) func(*config) {
	return func(conf *config) {
		mux := http.NewServeMux()
		mockBackendConfigServer := &mockHttpServer{Transformations: transformations}
		mux.HandleFunc(getByVersionIdEndPoint, mockBackendConfigServer.handleGetByVersionId)
		backendConfigSvc := kithttptest.NewServer(mux)

		conf.updateBackendConfigURL(dockerTestHelper.ToInternalDockerHost(backendConfigSvc.URL))
		if conf.extraHosts == nil {
			conf.extraHosts = make([]string, 0)
		}
		conf.extraHosts = append(conf.extraHosts, "host.docker.internal:host-gateway")
		cleaner.Cleanup(func() {
			backendConfigSvc.Close()
		})
	}
}

// WithConfigBackendURL lets transformer use custom backend config server for transformations
// WithConfigBackendURL should not be used with WithTransformations option
func WithConfigBackendURL(url string) func(*config) {
	return func(conf *config) {
		conf.updateBackendConfigURL(url)
	}
}

func WithDockerImageTag(tag string) func(*config) {
	return func(conf *config) {
		conf.tag = tag
	}
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...func(conf *config)) (*Resource, error) {
	// Set Rudder Transformer
	// pulls an image first to make sure we don't have an old cached version locally,
	// then it creates a container based on it and runs it
	conf := &config{
		repository:   "rudderstack/rudder-transformer",
		tag:          "latest",
		exposedPorts: []string{"9090"},
		envs: []string{
			"CONFIG_BACKEND_URL=https://api.rudderstack.com",
		},
	}

	for _, opt := range opts {
		opt(conf)
	}

	err := pool.Client.PullImage(docker.PullImageOptions{
		Repository: conf.repository,
		Tag:        conf.tag,
	}, docker.AuthConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	transformerContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   conf.repository,
		Tag:          conf.tag,
		ExposedPorts: conf.exposedPorts,
		Env:          conf.envs,
		ExtraHosts:   conf.extraHosts,
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
		if resp.StatusCode != http.StatusOK {
			return errors.New(resp.Status)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transformerResource, nil
}
