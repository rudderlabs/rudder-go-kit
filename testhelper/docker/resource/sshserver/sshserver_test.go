package sshserver

import (
	"path/filepath"
	"testing"

	"github.com/melbahja/goph"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestCredentials(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	res, err := Setup(pool, t,
		WithCredentials("qux", "foobar"),
	)
	require.NoError(t, err)

	c, err := goph.NewConn(&goph.Config{
		Addr:     "localhost",
		Port:     uint(res.Port),
		User:     "qux",
		Auth:     goph.Password("foobar"),
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Close())
	})

	out, err := c.Run("ls /")
	require.NoError(t, err)
	require.Contains(t, string(out), `home`)
}

func TestKeys(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	network, err := pool.Client.CreateNetwork(dc.CreateNetworkOptions{Name: "test_network"})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pool.Client.RemoveNetwork(network.ID); err != nil {
			t.Logf("Error while removing Docker network: %v", err)
		}
	})

	publicKeyPath, err := filepath.Abs("./testdata/test_key.pub")
	require.NoError(t, err)
	res, err := Setup(pool, t,
		WithPublicKeyPath(publicKeyPath),
		WithCredentials("linuxserver.io", ""),
		WithDockerNetwork(network),
	)
	require.NoError(t, err)

	privateKeyPath, err := filepath.Abs("./testdata/test_key")
	require.NoError(t, err)
	auth, err := goph.Key(privateKeyPath, "")
	require.NoError(t, err)

	c, err := goph.NewConn(&goph.Config{
		Addr:     res.Host,
		Port:     uint(res.Port),
		User:     "linuxserver.io",
		Auth:     auth,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Close())
	})

	out, err := c.Run("ls /")
	require.NoError(t, err)
	require.Contains(t, string(out), `home`)
}
