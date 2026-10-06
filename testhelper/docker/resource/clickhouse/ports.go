package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ory/dockertest/v3/docker"
)

// portBindingsTimeout bounds the wait for Docker to report published ports.
const portBindingsTimeout = 15 * time.Second

func exposedPorts(ports []string) []string {
	exposed := make([]string, 0, len(ports))
	for _, port := range ports {
		exposed = append(exposed, port+"/tcp")
	}
	return exposed
}

// waitForPortBindings returns a container state that has a host binding for every port.
// Docker Desktop can return a started container before it reports the bindings, so the state is inspected
// again with a short backoff until the bindings appear or timeout passes.
func waitForPortBindings(ctx context.Context, inspect func(string) (*docker.Container, error), container *docker.Container, ports []string, timeout time.Duration) (*docker.Container, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	id := container.ID
	delay := 50 * time.Millisecond
	// lastInspectErr keeps a persistent Docker failure visible when the deadline passes.
	var lastInspectErr error
	for {
		if hasPortBindings(container, ports) {
			return container, nil
		}
		select {
		case <-ctx.Done():
			if lastInspectErr != nil {
				return nil, fmt.Errorf("reading ClickHouse node port bindings: %w (last inspect: %w)", ctx.Err(), lastInspectErr)
			}
			return nil, fmt.Errorf("reading ClickHouse node port bindings: %w", ctx.Err())
		case <-time.After(delay):
		}
		delay = min(2*delay, time.Second)
		latest, err := inspect(id)
		if err != nil {
			lastInspectErr = err
			continue
		}
		lastInspectErr = nil
		container = latest
	}
}

func hasPortBindings(container *docker.Container, ports []string) bool {
	if container.NetworkSettings == nil {
		return false
	}
	for _, port := range ports {
		bindings := container.NetworkSettings.Ports[docker.Port(port+"/tcp")]
		if len(bindings) == 0 || bindings[0].HostPort == "" {
			return false
		}
	}
	return true
}
