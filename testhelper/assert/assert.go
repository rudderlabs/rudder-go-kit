package assert

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func RequireEventuallyResponse(
	t *testing.T, expectedStatusCode int, r *http.Request,
	waitFor, pollInterval time.Duration,
) string {
	t.Helper()

	var (
		body             []byte
		actualStatusCode int
	)
	require.Eventuallyf(t,
		func() bool {
			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				return false
			}

			defer func() { _ = resp.Body.Close() }()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return false
			}

			actualStatusCode = resp.StatusCode
			return expectedStatusCode == actualStatusCode
		},
		waitFor, pollInterval,
		"Expected status code %d, got %d. Body: %s",
		expectedStatusCode, actualStatusCode, string(body),
	)

	return string(body)
}
