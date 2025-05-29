package retryablehttp

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"

	conf "github.com/rudderlabs/rudder-go-kit/config"
)

type HttpClient interface {
	Do(method, url string, body io.Reader, headers map[string]string) (*http.Response, error)
}

type Config struct {
	// MaxRetry is the maximum number of retries.
	MaxRetry conf.ValueLoader[int]
	//  InitialInterval is the initial interval between retries.
	InitialInterval conf.ValueLoader[time.Duration]
	// MaxInterval is the maximum interval between retries.
	MaxInterval conf.ValueLoader[time.Duration]
	// Multiplier is the multiplier used to calculate the next interval.
	MaxElapsedTime conf.ValueLoader[time.Duration]
	// Multiplier is the multiplier used to increase the interval between retries.
	Multiplier conf.ValueLoader[float64]
}

// NewDefaultConfig creates a new Config with default retry settings.
//
//	MaxRetry: Maximum number of retries (default: 5)
//	InitialInterval: Initial retry interval in milliseconds (default: 100ms)
//	MaxInterval: Maximum retry interval in milliseconds (default: 1000ms)
//	MaxElapsedTime: Maximum total elapsed time for retries in seconds (default: 10s)
//	Multiplier: Backoff multiplier for retry intervals (default: 1.5)
func NewDefaultConfig() *Config {
	return &Config{
		MaxRetry:        conf.GetReloadableIntVar(5, 1, "retryablehttp.maxRetry"),
		InitialInterval: conf.GetReloadableDurationVar(100, time.Millisecond, "retryablehttp.initialInterval"),
		MaxInterval:     conf.GetReloadableDurationVar(1000, time.Millisecond, "retryablehttp.maxInterval"),
		MaxElapsedTime:  conf.GetReloadableDurationVar(10, time.Second, "retryablehttp.maxElapsedTime"),
		Multiplier:      conf.GetReloadableFloat64Var(1.5, "retryablehttp.multiplier"),
	}
}

type retryableHTTPClient struct {
	*http.Client
	config *Config
	// onFailure is called when a retryable error occurs
	onFailure func(err error, duration time.Duration)
}

type Option func(*retryableHTTPClient)

func WithHttpClient(client *http.Client) Option {
	return func(retryableHTTPClient *retryableHTTPClient) {
		retryableHTTPClient.Client = client
	}
}

func WithOnFailure(onFailure func(err error, duration time.Duration)) Option {
	return func(retryableHTTPClient *retryableHTTPClient) {
		retryableHTTPClient.onFailure = onFailure
	}
}

// NewRetryableHTTPClient creates a new retryable HTTP client with the specified configuration and options.
// It uses the `backoff` package to implement retry logic for HTTP requests.
// All the 5xx and 429 errors will be retried
// The default retry strategy is exponential backoff
// Parameters:
// - config: Configuration for the exponential backoff strategy
// - options: Optional functional options to further configure the client
//
// Returns:
// - HttpClient: A configured retryable HTTP client
//
// The client is initialized with a transport that has:
// - Keep-alives disabled
// - Maximum 100 connections per host
// - Maximum 10 idle connections per host
// - Idle connection timeout of 30 seconds
func NewRetryableHTTPClient(config *Config, options ...Option) HttpClient {
	if config == nil {
		config = NewDefaultConfig()
	}
	httpClient := &retryableHTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		config: config,
	}
	for _, option := range options {
		option(httpClient)
	}
	return httpClient
}

// Do executes an HTTP request with retry logic.
func (c *retryableHTTPClient) Do(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	_ = backoff.RetryNotify(
		func() error {
			var req *http.Request
			req, err = http.NewRequest(method, url, body)
			if err == nil {
				for key, value := range headers {
					req.Header.Set(key, value)
				}
				resp, err = c.Client.Do(req) // nolint: bodyclose
				// retry 5xx errors
				if err == nil && (resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusTooManyRequests) {
					return fmt.Errorf("non-success status code: %d", resp.StatusCode)
				}
			}
			return err
		},
		backoff.WithMaxRetries(
			backoff.NewExponentialBackOff(
				backoff.WithInitialInterval(c.config.InitialInterval.Load()),
				backoff.WithMaxInterval(c.config.MaxInterval.Load()),
				backoff.WithMaxElapsedTime(c.config.MaxElapsedTime.Load()),
				backoff.WithMultiplier(c.config.Multiplier.Load()),
			),
			uint64(c.config.MaxRetry.Load()),
		),
		func(err error, duration time.Duration) {
			if c.onFailure != nil {
				c.onFailure(err, duration)
			}
		},
	)
	return resp, err
}
