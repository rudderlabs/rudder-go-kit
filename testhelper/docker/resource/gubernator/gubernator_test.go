package gubernator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

// TestGubernator verifies the helper boots a working gubernator instance by exercising the rate limit endpoint
// over HTTP (so the test stays free of the gubernator Go client dependency).
// It sends more hits than the configured limit and asserts the instance reports OVER_LIMIT,
// proving it is both healthy and actually enforcing limits.
func TestGubernator(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	res, err := Setup(pool, t)
	require.NoError(t, err)
	require.NotEmpty(t, res.GRPCURL)
	require.NotEmpty(t, res.HTTPURL)

	const limit = 2
	getRateLimit := func(t *testing.T) string {
		t.Helper()
		// "hits" is the cost of the request
		// "algorithm" is an enum (1=TOKEN_BUCKET, 2=LEAKY_BUCKET)
		// "behavior" is a bitmask (2=GLOBAL)
		body := `{
			"requests": [
				{
					"name":"test",
					"unique_key":"key-1",
					"hits":1,
					"limit":2,
					"duration":1000,
					"algorithm":1,
					"behavior":2
				}
			]
		}`
		resp, err := http.Post(res.HTTPURL+"/v1/GetRateLimits", "application/json", bytes.NewBufferString(body))
		require.NoError(t, err)
		defer func() { httputil.CloseResponse(resp) }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var decoded struct {
			Responses []struct {
				Status string `json:"status"`
			} `json:"responses"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
		require.Len(t, decoded.Responses, 1)
		return decoded.Responses[0].Status
	}

	// First `limit` requests are under the limit.
	for range limit {
		require.Equal(t, "UNDER_LIMIT", getRateLimit(t))
	}
	// The next request exceeds the limit.
	require.Equal(t, "OVER_LIMIT", getRateLimit(t))
}
