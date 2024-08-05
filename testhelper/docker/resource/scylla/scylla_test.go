package scylla

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestScylla(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	scyllaContainer, err := Setup(pool, t)
	require.NoError(t, err)
	require.NotNil(t, scyllaContainer)

	cluster := gocql.NewCluster(scyllaContainer.URL)
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	session.Close()
}

func TestScylla_WithKeySpace(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	scyllaContainer, err := Setup(pool, t, WithKeyspace("test"))
	require.NoError(t, err)
	require.NotNil(t, scyllaContainer)

	cluster := gocql.NewCluster(scyllaContainer.URL)
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	session.Close()
}
