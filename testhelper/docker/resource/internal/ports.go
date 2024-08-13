package internal

import (
	"strconv"

	"github.com/ory/dockertest/v3/docker"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
)

func PortBindings(exposedPorts []string) (map[docker.Port][]docker.PortBinding, error) {
	portBindings := make(map[docker.Port][]docker.PortBinding)
	for _, exposedPort := range exposedPorts {
		localPort, err := testhelper.GetFreePort()
		if err != nil {
			return nil, err
		}
		portBindings[docker.Port(exposedPort)+"/tcp"] = []docker.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: strconv.Itoa(localPort),
			},
		}
	}

	return portBindings, nil
}
