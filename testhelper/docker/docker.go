package docker

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

// GetHostPort returns the desired port mapping
func GetHostPort(t testing.TB, port string, container *docker.Container) int {
	t.Helper()
	for p, bindings := range container.NetworkSettings.Ports {
		if p.Port() == port {
			pi, err := strconv.Atoi(bindings[0].HostPort)
			require.NoError(t, err)
			return pi
		}
	}
	return 0
}

// ToInternalDockerHost replaces localhost and 127.0.0.1 with host.docker.internal
func ToInternalDockerHost(url string) string {
	return regexp.MustCompile(`(localhost|127\.0\.0\.1)`).ReplaceAllString(url, "host.docker.internal")
}

func CreateNetwork(pool *dockertest.Pool, cln resource.Cleaner, prefix string) (*docker.Network, error) {
	network, err := pool.Client.CreateNetwork(docker.CreateNetworkOptions{Name: prefix + "_test_" + time.Now().Format("YY-MM-DD-") + rand.String(6)})
	if err != nil {
		return nil, err
	}

	cln.Cleanup(func() {
		if err := pool.Client.RemoveNetwork(network.ID); err != nil {
			cln.Logf("Error while removing Docker network: %v", err)
		}
	})

	return network, nil
}
