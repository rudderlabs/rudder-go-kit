package internal

import (
	"log"
	"net"
	"runtime"

	"github.com/ory/dockertest/v3/docker"
	"github.com/samber/lo"
)

const (
	BindHostIP       = "127.0.0.1"
	BindInternalHost = "host.docker.internal"
)

// IPv4PortBindings returns the port bindings for the given exposed ports forcing ipv4 address.
func IPv4PortBindings(exposedPorts []string, opts ...IPv4PortBindingsOpt) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	c := &ipv4PortBindingsConfig{
		ips: []string{BindHostIP},
	}
	if runtime.GOOS == "linux" && isHostDockerInternalAvailable() {
		c.ips = append(c.ips, BindInternalHost)
	}
	for _, opt := range opts {
		opt(c)
	}
	bindings := lo.Map(c.ips, func(ip string, _ int) docker.PortBinding {
		return docker.PortBinding{
			HostIP:   ip,
			HostPort: "0",
		}
	})
	for _, exposedPort := range exposedPorts {
		portBindings[docker.Port(exposedPort)+"/tcp"] = bindings
	}
	return portBindings
}

type IPv4PortBindingsOpt func(*ipv4PortBindingsConfig)

type ipv4PortBindingsConfig struct {
	ips []string
}

func WithBindIP(ip string) IPv4PortBindingsOpt {
	return func(c *ipv4PortBindingsConfig) {
		if ip != "" {
			c.ips = []string{ip}
		}
	}
}

func DefaultHostConfig(hc *docker.HostConfig) {
	hc.PublishAllPorts = false
}

func isHostDockerInternalAvailable() bool {
	ips, err := net.LookupHost("host.docker.internal")
	if err != nil {
		log.Printf("Error looking up host.docker.internal: %v", err)
	}
	return err == nil && len(ips) > 0
}
