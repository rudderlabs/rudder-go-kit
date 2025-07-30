package sshserver

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

const exposedPort = "2222"

type Option func(*config)

type config struct {
	publicKeyPath      string
	username, password string
	network            *dc.Network
}

// WithCredentials sets the username and password to use for the SSH server.
func WithCredentials(username, password string) Option {
	return func(c *config) {
		c.username = username
		c.password = password
	}
}

// WithPublicKeyPath sets the public key path to use for the SSH server.
func WithPublicKeyPath(publicKeyPath string) Option {
	return func(c *config) {
		c.publicKeyPath = publicKeyPath
	}
}

// WithDockerNetwork sets the Docker network to use for the SSH server.
func WithDockerNetwork(network *dc.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

type Resource struct {
	Host string
	Port int

	container *dockertest.Resource
}

func Setup(pool *dockertest.Pool, cln resource.Cleaner, opts ...Option) (*Resource, error) {
	var c config
	for _, opt := range opts {
		opt(&c)
	}

	network := c.network
	if c.network == nil {
		var err error
		network, err = pool.Client.CreateNetwork(dc.CreateNetworkOptions{Name: "sshserver_network"})
		if err != nil {
			return nil, fmt.Errorf("could not create docker network: %w", err)
		}
		cln.Cleanup(func() {
			if err := pool.Client.RemoveNetwork(network.ID); err != nil {
				cln.Log(fmt.Sprintf("could not remove sshserver_network: %v", err))
			}
		})
	}

	var (
		mounts  []string
		envVars = []string{
			"SUDO_ACCESS=false",
			"DOCKER_MODS=linuxserver/mods:openssh-server-ssh-tunnel",
		}
	)
	if c.username != "" {
		envVars = append(envVars, "USER_NAME="+c.username)
		if c.password != "" {
			envVars = append(envVars, []string{
				"USER_PASSWORD=" + c.password,
				"PASSWORD_ACCESS=true",
			}...)
		}
	}
	if c.publicKeyPath != "" {
		envVars = append(envVars, "PUBLIC_KEY_FILE=/test_key.pub")
		mounts = []string{c.publicKeyPath + ":/test_key.pub"}
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "hub.dev-rudder.rudderlabs.com/lscr.io/linuxserver/openssh-server",
		Tag:        "9.3_p2-r1-ls145",
		NetworkID:  network.ID,
		Hostname:   "sshserver",
		ExposedPorts: []string{
			exposedPort + "/tcp",
		},
		PortBindings: map[dc.Port][]dc.PortBinding{
			exposedPort + "/tcp": {
				{HostIP: "127.0.0.1", HostPort: "0"},
			},
		},
		Env:    envVars,
		Mounts: mounts,
		Auth: dc.AuthConfiguration{
			Username: os.Getenv("HARBOR_USER_NAME"),
			Password: os.Getenv("HARBOR_PASSWORD"),
		},
	}, internal.DefaultHostConfig)
	cln.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			cln.Log("Could not purge resource", err)
		}
	})
	if err != nil {
		return nil, err
	}

	var (
		buf     *bytes.Buffer
		timeout = time.After(60 * time.Second)
		ticker  = time.NewTicker(200 * time.Millisecond)
	)
loop:
	for {
		select {
		case <-ticker.C:
			buf = bytes.NewBuffer(nil)
			exitCode, err := container.Exec([]string{"cat", "/config/logs/openssh/current"}, dockertest.ExecOptions{
				StdOut: buf,
			})
			if err != nil {
				cln.Log("could not exec into SSH server:", err)
				continue
			}
			if exitCode != 0 {
				cln.Log("invalid exit code while exec-ing into SSH server:", exitCode)
				continue
			}
			if buf.String() == "" {
				cln.Log("SSH server not ready yet")
				continue
			}
			if !strings.Contains(buf.String(), "Server listening on :: port "+exposedPort) {
				cln.Log("SSH server not listening on port yet")
				continue
			}
			cln.Log("SSH server is ready:", exposedPort, "=>", container.GetPort(exposedPort+"/tcp"))
			break loop
		case <-timeout:
			return nil, fmt.Errorf("ssh server not health within timeout")
		}
	}
	p, err := strconv.Atoi(container.GetPort(exposedPort + "/tcp"))
	if err != nil {
		return nil, fmt.Errorf("could not convert port %q to int: %w", container.GetPort(exposedPort+"/tcp"), err)
	}

	return &Resource{
		Host:      container.GetBoundIP(exposedPort + "/tcp"),
		Port:      p,
		container: container,
	}, nil
}
