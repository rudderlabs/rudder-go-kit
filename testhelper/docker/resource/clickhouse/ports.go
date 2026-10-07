package clickhouse

import (
	"context"
	"errors"
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
// dockertest re-inspects a container only while a port has an empty binding list, so a state with no port keys
// at all is inspected again here with a short backoff until the bindings appear or timeout passes.
// An exited or removed container ends the wait at once.
func waitForPortBindings(ctx context.Context, inspect func(context.Context, string) (*docker.Container, error), container *docker.Container, ports []string, timeout time.Duration) (*docker.Container, error) {
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
		// A status check keeps a zero State, which a fake inspect returns, waiting.
		if status := container.State.Status; status == "exited" || status == "dead" {
			return nil, fmt.Errorf("reading ClickHouse node port bindings: container %s is %s (exit code %d, OOM killed: %t)",
				id, status, container.State.ExitCode, container.State.OOMKilled)
		}
		select {
		case <-ctx.Done():
			err := fmt.Errorf("reading ClickHouse node port bindings: Docker still reports no host binding for every port of container %s, check the Docker port forwarding: %w", id, ctx.Err())
			if lastInspectErr != nil {
				return nil, fmt.Errorf("%w (last inspect: %w)", err, lastInspectErr)
			}
			return nil, err
		case <-time.After(delay):
		}
		delay = min(2*delay, time.Second)
		latest, err := inspect(ctx, id)
		if err != nil {
			var missing *docker.NoSuchContainer
			if errors.As(err, &missing) {
				return nil, fmt.Errorf("reading ClickHouse node port bindings: %w", err)
			}
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
