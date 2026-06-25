// Package gubernator provides a docker test helper that starts a single-node
// gubernator (https://github.com/gubernator-io/gubernator) instance suitable
// for integration tests.
//
// gubernator is distributed as a "FROM scratch" image (no shell), so readiness
// is verified with an HTTP GET against the /v1/HealthCheck endpoint from the
// host rather than by exec-ing into the container.
//
// The instance is configured for member-list peer discovery pointing at itself
// over the loopback interface. This way the single node owns every rate limit
// key and reports itself as a healthy peer, which is required for GLOBAL
// behavior requests to be served.
package gubernator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const (
	grpcPort       = "1051"
	httpPort       = "1050"
	defaultVersion = "v2.19.1"
)

type Resource struct {
	// GRPCURL is the host:port to use with gubernator.DialV1Server.
	GRPCURL string
	// HTTPURL is the base URL (with scheme) of the gubernator HTTP gateway.
	HTTPURL string
	// GRPCURLInNetwork is the gRPC address reachable from the provided docker
	// network (if any).
	GRPCURLInNetwork string

	purge func()
}

// Purge can be called to remove the container before the test finishes. Useful
// to test how the system behaves when gubernator becomes unreachable.
func (r *Resource) Purge() { r.purge() }

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	c := &config{tag: defaultVersion}
	for _, opt := range opts {
		opt(c)
	}

	var networkID string
	if c.network != nil {
		networkID = c.network.ID
	}

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   registry.ImagePath("ghcr.io/gubernator-io/gubernator"),
		Tag:          c.tag,
		ExposedPorts: []string{grpcPort + "/tcp", httpPort + "/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{grpcPort, httpPort}),
		Env: []string{
			"GUBER_GRPC_ADDRESS=0.0.0.0:" + grpcPort,
			"GUBER_HTTP_ADDRESS=0.0.0.0:" + httpPort,
			// Advertise and discover ourselves over loopback so the single node forms a one-member cluster and
			// owns all rate limit keys.
			"GUBER_ADVERTISE_ADDRESS=127.0.0.1:" + grpcPort,
			"GUBER_PEER_DISCOVERY_TYPE=member-list",
			"GUBER_MEMBERLIST_KNOWN_NODES=127.0.0.1:7946",
		},
		NetworkID: networkID,
		Auth:      registry.AuthConfiguration(),
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot run gubernator container: %w", err)
	}

	purgeOnce := sync.Once{}
	purge := func() {
		purgeOnce.Do(func() {
			if err := pool.Purge(container); err != nil {
				d.Log("Could not purge resource:", err)
			}
		})
	}
	d.Cleanup(purge)

	grpcURL := fmt.Sprintf("127.0.0.1:%s", container.GetPort(grpcPort+"/tcp"))
	httpURL := fmt.Sprintf("http://127.0.0.1:%s", container.GetPort(httpPort+"/tcp"))

	if err := pool.Retry(func() error {
		resp, err := http.Get(httpURL + "/v1/HealthCheck")
		if err != nil {
			return err
		}
		defer func() { httputil.CloseResponse(resp) }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gubernator healthcheck status: %s", resp.Status)
		}
		var hc struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&hc); err != nil {
			return fmt.Errorf("decoding gubernator healthcheck: %w", err)
		}
		if hc.Status != "healthy" {
			return fmt.Errorf("gubernator not healthy: %s (%s)", hc.Status, hc.Message)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var grpcURLInNetwork string
	if c.network != nil {
		grpcURLInNetwork = container.GetIPInNetwork(&dockertest.Network{Network: c.network}) + ":" + grpcPort
	}

	return &Resource{
		GRPCURL:          grpcURL,
		HTTPURL:          httpURL,
		GRPCURLInNetwork: grpcURLInNetwork,
		purge:            purge,
	}, nil
}
