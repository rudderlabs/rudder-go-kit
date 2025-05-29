package retryablehttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	conf "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryableHTTPClient_Do_SuccessNoRetry(t *testing.T) {
	// Set up test server
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Verify request details
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify body
		bodyBytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"test":"data"}`, string(bodyBytes))

		// Return success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Create client with default config
	client := NewRetryableHTTPClient(nil)

	// Make request
	body := strings.NewReader(`{"test":"data"}`)
	headers := map[string]string{"Content-Type": "application/json"}
	resp, err := client.Do(http.MethodPost, ts.URL, body, headers)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 1, attempts) // Should not retry on success

	// Check response body
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{"status":"ok"}`, string(respBody))
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_RetryOn5xx(t *testing.T) {
	// Set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Return server error for first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Return success on third attempt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Create client with custom config (faster retries for testing)
	config := &Config{
		MaxRetry:        conf.GetReloadableIntVar(3, 1, "maxRetry"),
		InitialInterval: conf.GetReloadableDurationVar(10, time.Millisecond, "initialInterval"),
		MaxInterval:     conf.GetReloadableDurationVar(50, time.Millisecond, "maxInterval"),
		MaxElapsedTime:  conf.GetReloadableDurationVar(1, time.Second, "maxElapsedTime"),
		Multiplier:      conf.GetReloadableFloat64Var(1.5, "multiplier"),
	}
	client := NewRetryableHTTPClient(config)

	// Make request
	resp, err := client.Do(http.MethodGet, ts.URL, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts) // Should retry twice then succeed on third try
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_RetryOn429(t *testing.T) {
	// Set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Return rate limit error for first two attempts
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Return success on third attempt
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create client with custom config (faster retries for testing)
	config := &Config{
		MaxRetry:        conf.GetReloadableIntVar(3, 1, "maxRetry"),
		InitialInterval: conf.GetReloadableDurationVar(10, time.Millisecond, "initialInterval"),
		MaxInterval:     conf.GetReloadableDurationVar(50, time.Millisecond, "maxInterval"),
		MaxElapsedTime:  conf.GetReloadableDurationVar(1, time.Second, "maxElapsedTime"),
		Multiplier:      conf.GetReloadableFloat64Var(1.5, "multiplier"),
	}
	client := NewRetryableHTTPClient(config)

	// Make request
	resp, err := client.Do(http.MethodGet, ts.URL, nil, nil)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts) // Should retry twice then succeed
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_MaxRetriesExceeded(t *testing.T) {
	// Set up test server that always returns server error
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	// Create client with limited retries
	config := &Config{
		MaxRetry:        conf.GetReloadableIntVar(2, 1, "maxRetry2"),
		InitialInterval: conf.GetReloadableDurationVar(10, time.Millisecond, "initialInterval"),
		MaxInterval:     conf.GetReloadableDurationVar(50, time.Millisecond, "maxInterval"),
		MaxElapsedTime:  conf.GetReloadableDurationVar(1, time.Second, "maxElapsedTime"),
		Multiplier:      conf.GetReloadableFloat64Var(1.5, "multiplier"),
	}
	client := NewRetryableHTTPClient(config)

	// Make request
	resp, err := client.Do(http.MethodGet, ts.URL, nil, nil)

	// Assertions - should return the last failed response after max retries
	require.NoError(t, err) // Error is not returned, only the failed response
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, 3, attempts) // Initial + 2 retries
	resp.Body.Close()
}

func TestRetryableHTTPClient_WithCustomOptions(t *testing.T) {
	// Set up a successful test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create custom HTTP client
	customHTTPClient := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	// Create retryable client with custom HTTP client
	client := NewRetryableHTTPClient(nil, WithHttpClient(customHTTPClient))

	// Verify the client was set correctly (using type assertion)
	retryClient, ok := client.(*retryableHTTPClient)
	require.True(t, ok)
	assert.Equal(t, customHTTPClient, retryClient.Client)

	// Make request to verify it works
	resp, err := client.Do(http.MethodGet, ts.URL, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRetryableHTTPClient_WithOnFailure(t *testing.T) {
	// Set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	// Variables to track callback invocation
	var (
		failureCallCount int
		lastError        error
		mu               sync.Mutex
	)

	onFailure := func(err error, duration time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		failureCallCount++
		lastError = err
	}

	// Create client with onFailure callback
	config := &Config{
		MaxRetry:        conf.GetReloadableIntVar(3, 1, "maxRetry"),
		InitialInterval: conf.GetReloadableDurationVar(10, time.Millisecond, "initialInterval"),
		MaxInterval:     conf.GetReloadableDurationVar(50, time.Millisecond, "maxInterval"),
		MaxElapsedTime:  conf.GetReloadableDurationVar(1, time.Second, "maxElapsedTime"),
		Multiplier:      conf.GetReloadableFloat64Var(1.5, "multiplier"),
	}
	client := NewRetryableHTTPClient(config, WithOnFailure(onFailure))

	// Make request
	resp, err := client.Do(http.MethodGet, ts.URL, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify callback was called correctly
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, failureCallCount)
	assert.Contains(t, lastError.Error(), "non-success status code: 500")
}

func TestRetryableHTTPClient_RequestCreationError(t *testing.T) {
	client := NewRetryableHTTPClient(nil)

	// Use invalid URL to trigger request creation error
	resp, err := client.Do(http.MethodGet, "://invalid-url", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	assert.Equal(t, 5, config.MaxRetry.Load())
	assert.Equal(t, 100*time.Millisecond, config.InitialInterval.Load())
	assert.Equal(t, 1000*time.Millisecond, config.MaxInterval.Load())
	assert.Equal(t, 10*time.Second, config.MaxElapsedTime.Load())
	assert.Equal(t, 1.5, config.Multiplier.Load())
}
