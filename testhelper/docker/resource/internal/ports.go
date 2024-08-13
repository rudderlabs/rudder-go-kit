package internal

import (
	"github.com/ory/dockertest/v3/docker"
)

// IPv4PortBindings returns the port bindings for the given exposed ports forcing ipv4 address.
func IPv4PortBindings(exposedPorts []string) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	// p, err := testhelper.GetFreePort()
	// if err != nil {
	// 	panic(err)
	// }

	for _, exposedPort := range exposedPorts {
		portBindings[docker.Port(exposedPort)+"/tcp"] = []docker.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "0",
			},
		}
	}

	return portBindings
}
