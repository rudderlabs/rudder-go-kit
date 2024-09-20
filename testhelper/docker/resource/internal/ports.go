package internal

import (
	"runtime"

	"github.com/ory/dockertest/v3/docker"
)

const (
	BindHostIP       = "127.0.0.1"
	BindInternalHost = "host.docker.internal"
)

type additionalBindingOption func([]docker.PortBinding)

// IPv4PortBindings returns the port bindings for the given exposed ports forcing ipv4 address.
func IPv4PortBindings(exposedPorts []string, listenToAllInterfaces bool) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	bindings := []docker.PortBinding{
		{
			HostIP:   BindHostIP,
			HostPort: "0",
		},
	}

	if listenToAllInterfaces {
		bindings[0].HostIP = "0.0.0.0"
	}

	if runtime.GOOS == "linux" {
		bindings = append(bindings, docker.PortBinding{
			HostIP:   BindInternalHost,
			HostPort: "0",
		})
	}

	for _, exposedPort := range exposedPorts {
		portBindings[docker.Port(exposedPort)+"/tcp"] = bindings
	}

	return portBindings
}

func DefaultHostConfig(hc *docker.HostConfig) {
	hc.PublishAllPorts = false
}

// BindToAllInterfaces returns a function that appends a binding to all interfaces
func BindToAllInterfaces(port string) func([]docker.PortBinding) {
	return func(bindings []docker.PortBinding) {
		bindings = append(bindings, docker.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: port,
		})
	}
}
