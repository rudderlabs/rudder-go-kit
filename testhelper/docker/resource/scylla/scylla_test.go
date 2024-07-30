package scylla

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestScylla(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	scyllaContainer, err := Setup(pool, t)
	require.NoError(t, err)
	require.NotNil(t, scyllaContainer)
}
