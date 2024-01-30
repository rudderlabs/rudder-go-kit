package pulsar

import (
	"net/http"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

func TestPulsar(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	pulsarContainer, err := Setup(pool, t)
	require.NoError(t, err)

	res, err := http.Head(pulsarContainer.AdminURL + "/admin/v2/namespaces/public/default")
	defer func() { httputil.CloseResponse(res) }()
	require.NoError(t, err)
}
