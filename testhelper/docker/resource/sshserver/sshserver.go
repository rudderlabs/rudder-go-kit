package sshserver

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	kithelper "github.com/rudderlabs/rudder-go-kit/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
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

	port, err := kithelper.GetFreePort()
	if err != nil {
		return nil, err
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
		Repository: "lscr.io/linuxserver/openssh-server",
		Tag:        "9.3_p2-r1-ls145",
		NetworkID:  network.ID,
		Hostname:   "sshserver",
		PortBindings: map[dc.Port][]dc.PortBinding{
			exposedPort + "/tcp": {
				{HostIP: "sshserver", HostPort: fmt.Sprintf("%d", port)},
			},
		},
		Env:    envVars,
		Mounts: mounts,
	})
	if err != nil {
		return nil, err
	}
	cln.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			cln.Log("Could not purge resource", err)
		}
	})

	var (
		buf     *bytes.Buffer
		timeout = time.After(30 * time.Second)
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

	return &Resource{
		Port:      port,
		container: container,
	}, nil
}
