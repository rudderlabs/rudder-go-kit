package httputil

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	headerXForwardedFor = "X-Forwarded-For"

	DefaultMaxIdleConnsPerHost = 10
	DefaultMaxConnsPerHost     = 100
	DefaultDisableKeepAlives   = true
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
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = DefaultDisableKeepAlives
	tr.MaxConnsPerHost = DefaultMaxConnsPerHost
	tr.MaxIdleConnsPerHost = DefaultMaxIdleConnsPerHost
	return tr
}

type HttpClientOptions func(*http.Client)

// DefaultHttpClient returns a default HTTP client with a custom transport configuration.
// It disables keep-alives, sets max connections per host, and configures idle connection timeout.
// This is useful for clients that need to make many short-lived requests without reusing connections.
// It also sets a default timeout of 30 seconds.
// e.g.
//
//	func Example() {
//		// no need to use .Clone() since a new transport is built each time
//		transport := httputil.DefaultTransport()
//		transport.ForceAttemptHTTP2 = false
//		client := httputil.NewHttpClient(httputil.WithTransport(transport))
//		req, err := http.NewRequest("GET", "https://example.com", nil)
//		if err != nil {
//			panic(err)
//		}
//		resp, err := client.Do(req)
//		if err != nil {
//			panic(err)
//		}
//		_ = resp.Body.Close()
//	}
func DefaultHttpClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
		Timeout:   DefaultRequestTimeout,
	}
}

func NewHttpClient(options ...HttpClientOptions) *http.Client {
	client := DefaultHttpClient()
	for _, option := range options {
		option(client)
	}
	return client
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
