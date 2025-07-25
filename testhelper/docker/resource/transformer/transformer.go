package transformer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	dockertesthelper "github.com/rudderlabs/rudder-go-kit/testhelper/docker"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

const transformerPort = "9090/tcp"

type Resource struct {
	TransformerURL string
}

type Option func(*config)

type config struct {
	repository   string
	tag          string
	exposedPorts []string
	envs         []string
	extraHosts   []string
	network      *docker.Network
	authConfig   docker.AuthConfiguration
	bindIP       string
}

func (c *config) setBackendConfigURL(url string) {
	c.envs = append(
		lo.Filter(c.envs, func(s string, _ int) bool {
			return !strings.HasPrefix(s, "CONFIG_BACKEND_URL=")
		}),
		"CONFIG_BACKEND_URL="+url)
}

// WithUserTransformations will mock BE config to set transformation for given transformation versionID to transformation function map
//
// - events with transformationVersionID not present in map will not be transformed and transformer will return 404 for those requests
//
// - WithUserTransformations should not be used with WithConfigBackendURL option
//
// - only javascript transformation functions are supported
//
// e.g.
//
//	WithUserTransformations(map[string]string{
//				"transform-version-id-1": `export function transformEvent(event, metadata) {
//											event.transformed=true
//											return event;
//										}`,
//			})
func WithUserTransformations(transformations map[string]string, cleaner resource.Cleaner) Option {
	return func(conf *config) {
		backendConfigSvc := newTestBackendConfigServer(transformations)

		conf.setBackendConfigURL(dockertesthelper.ToInternalDockerHost(backendConfigSvc.URL))
		conf.extraHosts = append(conf.extraHosts, "host.docker.internal:host-gateway")
		cleaner.Cleanup(func() {
			backendConfigSvc.Close()
		})
	}
}

// WithConnectionToHostEnabled lets transformer container connect with the host machine
// i.e. transformer container will be able to access localhost of the host machine
func WithConnectionToHostEnabled() Option {
	return func(conf *config) {
		conf.extraHosts = append(conf.extraHosts, "host.docker.internal:host-gateway")
	}
}

// WithConfigBackendURL lets transformer use custom backend config server for transformations
// WithConfigBackendURL should not be used with WithUserTransformations option
func WithConfigBackendURL(url string) Option {
	return func(conf *config) {
		conf.setBackendConfigURL(dockertesthelper.ToInternalDockerHost(url))
	}
}

func WithDockerImageTag(tag string) Option {
	return func(conf *config) {
		conf.tag = tag
	}
}

func WithDockerNetwork(network *docker.Network) Option {
	return func(conf *config) {
		conf.network = network
	}
}

func WithEnv(env string) Option {
	return func(conf *config) {
		conf.envs = append(conf.envs, env)
	}
}

func WithRepository(repository string) Option {
	return func(conf *config) {
		conf.repository = repository
	}
}

func WithDockerAuth(authConfig docker.AuthConfiguration) Option {
	return func(conf *config) {
		conf.authConfig = authConfig
	}
}

func WithBindIP(bindIP string) Option {
	return func(conf *config) {
		conf.bindIP = bindIP
	}
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	// Set Rudder Transformer
	// pulls an image first to make sure we don't have an old cached version locally,
	// then it creates a container based on it and runs it
	conf := &config{
		repository:   "rudderstack/rudder-transformer",
		tag:          "latest",
		exposedPorts: []string{"9090"},
		envs: []string{
			"CONFIG_BACKEND_URL=https://api.rudderstack.com",
			"NODE_OPTIONS=--no-node-snapshot",
		},
		authConfig: docker.AuthConfiguration{},
	}

	for _, opt := range opts {
		opt(conf)
	}

	if err := pool.Client.PullImage(docker.PullImageOptions{
		Repository: conf.repository,
		Tag:        conf.tag,
	}, conf.authConfig); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	var networkID string
	if conf.network != nil {
		networkID = conf.network.ID
	}
	transformerContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   conf.repository,
		Tag:          conf.tag,
		PortBindings: internal.IPv4PortBindings(conf.exposedPorts, internal.WithBindIP(conf.bindIP)),
		Env:          conf.envs,
		ExtraHosts:   conf.extraHosts,
		NetworkID:    networkID,
		Auth:         conf.authConfig,
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(transformerContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	transformerResource := &Resource{
		TransformerURL: fmt.Sprintf("http://%s:%s", transformerContainer.GetBoundIP(transformerPort), transformerContainer.GetPort(transformerPort)),
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
