package sshserver

import (
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/melbahja/goph"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestResourceCredentials(t *testing.T) {
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
