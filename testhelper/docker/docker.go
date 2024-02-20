package docker

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
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
	url = strings.ReplaceAll(url, "localhost", "host.docker.internal")
	url = strings.ReplaceAll(url, "127.0.0.1", "host.docker.internal")
	return url
}
