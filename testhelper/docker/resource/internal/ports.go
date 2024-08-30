package internal

import (
	"github.com/ory/dockertest/v3/docker"
)

const BindHostIP = "0.0.0.0"

// IPv4PortBindings returns the port bindings for the given exposed ports forcing ipv4 address.
func IPv4PortBindings(exposedPorts []string) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for _, exposedPort := range exposedPorts {
		portBindings[docker.Port(exposedPort)+"/tcp"] = []docker.PortBinding{
			{
				HostIP:   BindHostIP,
				HostPort: "0",
			},
		}
	}

	return portBindings
}

func DefaultHostConfig(hc *docker.HostConfig) {
	hc.PublishAllPorts = false
}
