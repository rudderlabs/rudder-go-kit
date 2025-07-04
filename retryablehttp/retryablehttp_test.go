package retryablehttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryableHTTPClient_Do_SuccessNoRetry(t *testing.T) {
	// set up test server
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// verify request details
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// verify body
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, `{"test":"data"}`, string(bodyBytes))

		// return success
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"status":"ok"}`))
		require.NoError(t, err)
	}))
	defer ts.Close()

	// create client with default config
	client := NewRetryableHTTPClient(nil)

	// make request
	body := strings.NewReader(`{"test":"data"}`)
	headers := map[string]string{"Content-Type": "application/json"}
	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	require.NoError(t, err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)

	// assertions
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

func TestRetryableHTTPClient_Do_PostRetryOn5xx(t *testing.T) {
	// set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify request details
		require.Equal(t, http.MethodPost, r.Method)
		// require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// verify body
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, `{"test":"data"}`, string(bodyBytes))
		attempts++
		if attempts < 3 {
			// return server error for first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// return success on third attempt
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"status":"ok"}`))
		require.NoError(t, err)
	}))
	defer ts.Close()

	// create client with custom config (faster retries for testing)
	config := &Config{
		MaxRetry:        3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}
	client := NewRetryableHTTPClient(config)

	// make request
	body := strings.NewReader(`{"test":"data"}`)
	headers := map[string]string{"Content-Type": "application/json"}
	req, err := http.NewRequest(http.MethodPost, ts.URL, body)
	require.NoError(t, err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)

	// assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 3, attempts) // Should retry twice then succeed on third try
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_RetryOn5xx(t *testing.T) {
	// set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// return server error for first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// return success on third attempt
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		require.NoError(t, err)
	}))
	defer ts.Close()

	// create client with custom config (faster retries for testing)
	config := &Config{
		MaxRetry:        3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}
	client := NewRetryableHTTPClient(config)

	// make request
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)

	// assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 3, attempts) // Should retry twice then succeed on third try
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_RetryOn429(t *testing.T) {
	// set up test server
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

	// create client with custom config (faster retries for testing)
	config := &Config{
		MaxRetry:        3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}
	client := NewRetryableHTTPClient(config)

	// make request
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)

	// assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 3, attempts) // Should retry twice then succeed
	resp.Body.Close()
}

func TestRetryableHTTPClient_Do_MaxRetriesExceeded(t *testing.T) {
	// set up test server that always returns server error
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	// create client with limited retries
	config := &Config{
		MaxRetry:        2,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}
	client := NewRetryableHTTPClient(config)

	// make request
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)

	// assertions - should return the last failed response after max retries
	require.NoError(t, err) // Error is not returned, only the failed response
	require.NotNil(t, resp)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, 3, attempts) // Initial + 2 retries
	_ = resp.Body.Close()
}

func TestRetryableHTTPClient_WithCustomOptions(t *testing.T) {
	// set up a successful test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// create custom HTTP client
	customHTTPClient := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	// create retryable client with custom HTTP client
	client := NewRetryableHTTPClient(nil, WithHttpClient(customHTTPClient))

	// verify the client was set correctly (using type assertion)
	retryClient, ok := client.(*retryableHTTPClient)
	require.True(t, ok)
	require.Equal(t, customHTTPClient, retryClient.HttpClient)

	// make request to verify it works
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRetryableHTTPClient_WithOnFailure(t *testing.T) {
	// set up test server
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

	// variables to track callback invocation
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

	// create client with onFailure callback
	config := &Config{
		MaxRetry:        3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}
	client := NewRetryableHTTPClient(config, WithOnFailure(onFailure))

	// make request
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// verify callback was called correctly
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, failureCallCount)
	require.Contains(t, lastError.Error(), "unexpected HTTP status 500 Internal Server Error")
}

func TestRetryableHTTPClient_WithCustomRetryStrategy(t *testing.T) {
	// set up test server
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// return 404 for all requests (normally wouldn't trigger retry)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// create a custom retry strategy that retries on 404s (which normally wouldn't retry)
	customRetryCount := 0
	customRetryStrategy := func(resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		// specifically retry on 404 status, which default strategy wouldn't retry
		if resp.StatusCode == http.StatusNotFound {
			customRetryCount++
			// only retry twice to avoid infinite loop
			if customRetryCount < 3 {
				return true, fmt.Errorf("custom retry strategy: retrying on 404")
			}
		}
		return false, nil
	}

	// create client with custom retry strategy and fast retry intervals
	config := &Config{
		MaxRetry:        3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}

	// Track failures with a callback
	failureCalls := 0
	onFailure := func(err error, duration time.Duration) {
		failureCalls++
		require.Contains(t, err.Error(), "custom retry strategy")
	}

	client := NewRetryableHTTPClient(
		config,
		WithCustomRetryStrategy(customRetryStrategy),
		WithOnFailure(onFailure),
	)

	// Make request
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, 3, attempts)         // Initial + 2 retries
	require.Equal(t, 2, failureCalls)     // Should be called for each retry
	require.Equal(t, 3, customRetryCount) // Our strategy should have been called and incremented
	resp.Body.Close()
}

func TestRetryableHTTPClient_WithCustomRetryStrategyHttpClientReturnError(t *testing.T) {
	// create a custom retry strategy that retries on 404s (which normally wouldn't retry)
	customRetryCount := 0
	customRetryStrategy := func(resp *http.Response, err error) (bool, error) {
		customRetryCount++
		if customRetryCount < 3 {
			return true, fmt.Errorf("custom retry strategy: retrying")
		}
		return false, nil
	}

	// create client with custom retry strategy and fast retry intervals
	config := &Config{
		MaxRetry:        10,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		MaxElapsedTime:  time.Second,
		Multiplier:      1.5,
	}

	// Track failures with a callback
	failureCalls := 0
	onFailure := func(err error, duration time.Duration) {
		failureCalls++
		require.Contains(t, err.Error(), "custom retry strategy")
	}

	client := NewRetryableHTTPClient(
		config,
		WithCustomRetryStrategy(customRetryStrategy),
		WithOnFailure(onFailure),
	)

	// Make request
	req, err := http.NewRequest(http.MethodGet, "http://random-endpoint-akdhf", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, 2, failureCalls)     // Should be called for each retry
	require.Equal(t, 3, customRetryCount) // Our strategy should have been called and incremented

	if resp != nil {
		resp.Body.Close()
	}
}

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	require.Equal(t, 5, config.MaxRetry)
	require.Equal(t, 100*time.Millisecond, config.InitialInterval)
	require.Equal(t, 1000*time.Millisecond, config.MaxInterval)
	require.Equal(t, 10*time.Second, config.MaxElapsedTime)
	require.Equal(t, 1.5, config.Multiplier)
}
