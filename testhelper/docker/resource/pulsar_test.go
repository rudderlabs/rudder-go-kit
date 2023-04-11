package resource_test

import (
	"net/http"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/stretchr/testify/require"
)

func TestPulsar(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	pulsarContainer, err := resource.SetupPulsar(pool, t)
	require.NoError(t, err)
	res, err := http.Head(pulsarContainer.AdminURL + "/admin/v2/namespaces/public/default")
	defer func() { httputil.CloseResponse(res) }()
	require.NoError(t, err)
}
