package assert

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type (
	ResponseBody            = string
	RequireStatusCodeOption func(*requireStatusCodeConfig)
)

type requireStatusCodeConfig struct {
	waitFor      time.Duration
	pollInterval time.Duration
	httpClient   *http.Client
}

func (c *requireStatusCodeConfig) reset() {
	c.waitFor = 10 * time.Second
	c.pollInterval = 100 * time.Millisecond
	c.httpClient = http.DefaultClient
}

func WithRequireStatusCodeWaitFor(waitFor time.Duration) RequireStatusCodeOption {
	return func(c *requireStatusCodeConfig) {
		c.waitFor = waitFor
	}
}

func WithRequireStatusCodePollInterval(pollInterval time.Duration) RequireStatusCodeOption {
	return func(c *requireStatusCodeConfig) {
		c.pollInterval = pollInterval
	}
}

func WithRequireStatusCodeHTTPClient(httpClient *http.Client) RequireStatusCodeOption {
	return func(c *requireStatusCodeConfig) {
		c.httpClient = httpClient
	}
}

// RequireEventuallyStatusCode is a helper function that retries a request until the expected status code is returned.
func RequireEventuallyStatusCode(
	t *testing.T, expectedStatusCode int, r *http.Request, opts ...RequireStatusCodeOption,
) ResponseBody {
	t.Helper()

	var config requireStatusCodeConfig
	config.reset()
	for _, opt := range opts {
		opt(&config)
	}

	var (
		body             []byte
		actualStatusCode int
	)
	require.Eventuallyf(t,
		func() bool {
			resp, err := config.httpClient.Do(r)
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
		config.waitFor, config.pollInterval,
		"Expected status code %d, got %d. Body: %s",
		expectedStatusCode, actualStatusCode, string(body),
	)

	return string(body)
}
