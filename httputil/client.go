package httputil

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
)

const (
	headerXForwardedFor = "X-Forwarded-For"

	DefaultMaxIdleConnsPerHost   = 10
	DefaultMaxConnsPerHost       = 100
	DefaultDisableKeepAlives     = false
	DefaultForceHTTP2            = true
	DefaultMaxIdleConns          = 100
	DefaultIdleConnTimeout       = 90 * time.Second
	DefaultTLSHandshakeTimeout   = 10 * time.Second
	DefaultExpectContinueTimeout = 1 * time.Second
	// DefaultRequestTimeout is the default timeout for HTTP requests for default HttpClient.
	DefaultRequestTimeout = 30 * time.Second
)

// CloseResponse closes the response's body. But reads at least some of the body so if it's
// small the underlying TCP connection will be re-used. No need to check for errors: if it
// fails, the Transport won't reuse it anyway.
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		const maxBodySlurpSize = 2 << 10 // 2KB
		_, _ = io.CopyN(io.Discard, resp.Body, maxBodySlurpSize)
		resp.Body.Close()
	}
}

func GetRequestIP(req *http.Request) string {
	addresses := strings.Split(req.Header.Get(headerXForwardedFor), ",")
	if addresses[0] == "" {
		splits := strings.Split(req.RemoteAddr, ":")
		return strings.Join(splits[:len(splits)-1], ":") // When there is no load-balancer
	}

	return strings.ReplaceAll(addresses[0], " ", "")
}

func DefaultTransport() *http.Transport {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     DefaultForceHTTP2,
		MaxIdleConns:          DefaultMaxIdleConns,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		ExpectContinueTimeout: DefaultExpectContinueTimeout,
		DisableKeepAlives:     DefaultDisableKeepAlives,
		MaxConnsPerHost:       DefaultMaxConnsPerHost,
		MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
	}
	return tr
}

type HttpClientOptions func(*http.Client)

// DefaultHttpClient returns a default HTTP client with a custom transport configuration.
// It disables keep-alives, sets max connections per host, and configures idle connection timeout.
// This is useful for clients that need to make many short-lived requests without reusing connections.
// It also sets a default timeout of 30 seconds.
func DefaultHttpClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
		Timeout:   DefaultRequestTimeout,
	}
}

func WithTimeout(timeout time.Duration) HttpClientOptions {
	return func(client *http.Client) {
		if client == nil {
			client = DefaultHttpClient()
		}
		client.Timeout = timeout
	}
}

// WithTransport returns a HttpClientOptions that sets a custom transport for the HTTP client.
func WithTransport(transport *http.Transport) HttpClientOptions {
	return func(client *http.Client) {
		if client == nil {
			client = DefaultHttpClient()
		}
		client.Transport = transport
	}
}

// WithConfig returns a HttpClientOptions that sets a custom configuration for the HTTP client.
func WithConfig(conf *config.Config, prefix ...string) HttpClientOptions {
	return func(client *http.Client) {
		if client == nil {
			client = DefaultHttpClient()
		}
		tr := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     conf.GetBoolVar(DefaultForceHTTP2, getConfigKeys("ForceHTTP2", prefix...)...),
			MaxIdleConns:          conf.GetIntVar(DefaultMaxIdleConns, 1, getConfigKeys("MaxIdleConns", prefix...)...),
			IdleConnTimeout:       conf.GetDurationVar(int64(DefaultIdleConnTimeout/time.Second), time.Second, getConfigKeys("IdleConnTimeout", prefix...)...),
			TLSHandshakeTimeout:   conf.GetDurationVar(int64(DefaultTLSHandshakeTimeout/time.Second), time.Second, getConfigKeys("TLSHandshakeTimeout", prefix...)...),
			ExpectContinueTimeout: conf.GetDurationVar(int64(DefaultExpectContinueTimeout/time.Second), time.Second, getConfigKeys("ExpectContinueTimeout", prefix...)...),
			DisableKeepAlives:     conf.GetBoolVar(DefaultDisableKeepAlives, getConfigKeys("DisableKeepAlives", prefix...)...),
			MaxConnsPerHost:       conf.GetIntVar(DefaultMaxConnsPerHost, 1, getConfigKeys("MaxConnsPerHost", prefix...)...),
			MaxIdleConnsPerHost:   conf.GetIntVar(DefaultMaxIdleConnsPerHost, 1, getConfigKeys("MaxIdleConnsPerHost", prefix...)...),
		}
		client.Transport = tr
	}
}

func getConfigKeys(key string, prefix ...string) []string {
	res := make([]string, 0, len(prefix)+1)
	for _, p := range prefix {
		res = append(res, p+"."+key)
	}
	res = append(res, "DefaultHttpClient."+key)
	return res
}

// NewHttpClient creates a configured HTTP client with customizable options.
// It initializes a client with DefaultHttpClient settings and applies functional options.
// Parameters:
//   - options: Variadic list of HttpClientOptions functions to customize client behavior
//
// Retur
// Returns:
//   - *http.Client configured with specified options
func NewHttpClient(options ...HttpClientOptions) *http.Client {
	client := DefaultHttpClient()
	for _, option := range options {
		option(client)
	}
	return client
}
