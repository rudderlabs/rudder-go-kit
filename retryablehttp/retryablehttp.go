package retryablehttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"

	conf "github.com/rudderlabs/rudder-go-kit/config"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// retryStrategy is a function that determines whether to retry based on the response and error.
// It should return true if the request should be retried, along with an error for the reason of retry.
// If it returns false, it means no retry is needed and the error(if any while making http req) along with a response can be returned directly.
// if retry is true, it means the request should be retried and error(if any from retryStrategy) will be sent to onFailure.
type retryStrategy func(resp *http.Response, err error) (bool, error)

type Config struct {
	// MaxRetry is the maximum number of retries.
	MaxRetry int
	//  InitialInterval is the initial interval between retries.
	InitialInterval time.Duration
	// MaxInterval is the maximum interval between retries.
	MaxInterval time.Duration
	// MaxElapsedTime is the maximum total elapsed time for retries.
	MaxElapsedTime time.Duration
	// Multiplier is the multiplier used to increase the interval between retries.
	Multiplier float64
}

type retryableHTTPClient struct {
	HttpClient
	config *Config
	// shouldRetry is a function that determines whether to retry based on the response and error.
	shouldRetry retryStrategy
	// onFailure is called when a retryable error occurs
	onFailure func(err error, duration time.Duration)
}

type Option func(*retryableHTTPClient)

func WithHttpClient(client HttpClient) Option {
	return func(retryableHTTPClient *retryableHTTPClient) {
		retryableHTTPClient.HttpClient = client
	}
}

func WithOnFailure(onFailure func(err error, duration time.Duration)) Option {
	return func(retryableHTTPClient *retryableHTTPClient) {
		retryableHTTPClient.onFailure = onFailure
	}
}

func WithCustomRetryStrategy(retryStrategy retryStrategy) Option {
	return func(retryableHTTPClient *retryableHTTPClient) {
		retryableHTTPClient.shouldRetry = retryStrategy
	}
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
		MaxRetry:        conf.GetInt("retryablehttp.maxRetry", 5),
		InitialInterval: conf.GetDuration("retryablehttp.initialInterval", 100, time.Millisecond),
		MaxInterval:     conf.GetDuration("retryablehttp.maxInterval", 1000, time.Millisecond),
		MaxElapsedTime:  conf.GetDuration("retryablehttp.maxElapsedTime", 10, time.Second),
		Multiplier:      conf.GetFloat64("retryablehttp.multiplier", 1.5),
	}
}

// NewRetryableHTTPClient creates a new retryable HTTP client with the specified configuration and options.
// It uses the `backoff` package to implement retry logic for HTTP requests.
// by default all the 5xx and 429 errors will be retried
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
	client := &retryableHTTPClient{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		config:      config,
		shouldRetry: BaseRetryStrategy,
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// Do executes an HTTP request with retry logic.
func (c *retryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	var bodyBytes []byte
	// if the body is not nil, read it and store it in bodyBytes
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	_ = backoff.RetryNotify(
		func() error {
			// if the body was read, we need to reset it
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			resp, err = c.HttpClient.Do(req) // nolint: bodyclose
			if retry, retryErr := c.shouldRetry(resp, err); retry {
				return fmt.Errorf("retryable error: %w", retryErr)
			}
			return nil
		},
		backoff.WithMaxRetries(
			backoff.NewExponentialBackOff(
				backoff.WithInitialInterval(c.config.InitialInterval),
				backoff.WithMaxInterval(c.config.MaxInterval),
				backoff.WithMaxElapsedTime(c.config.MaxElapsedTime),
				backoff.WithMultiplier(c.config.Multiplier),
			),
			uint64(c.config.MaxRetry),
		),
		func(err error, duration time.Duration) {
			if c.onFailure != nil {
				c.onFailure(err, duration)
			}
		},
	)
	return resp, err
}

func BaseRetryStrategy(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	// 429 Too Many Requests is recoverable.
	// It indicates that the client has sent too many requests in a given amount of time.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, fmt.Errorf("too many requests")
	}

	//  We retry on 5xx responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}
