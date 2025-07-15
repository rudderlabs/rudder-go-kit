package httputil_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

func TestGetRequestIP(t *testing.T) {
	testCases := []struct {
		name           string
		headerValue    string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Forwarded-For provided",
			headerValue:    "192.168.0.1, 192.168.0.2",
			remoteAddr:     "192.168.0.3:8080",
			expectedResult: "192.168.0.1",
		},
		{
			name:           "X-Forwarded-For empty, RemoteAddr provided",
			headerValue:    "",
			remoteAddr:     "192.168.0.4:8080",
			expectedResult: "192.168.0.4",
		},
		{
			name:           "X-Forwarded-For and RemoteAddr both empty",
			headerValue:    "",
			remoteAddr:     "",
			expectedResult: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := &http.Request{
				Header:     http.Header{"X-Forwarded-For": {testCase.headerValue}},
				RemoteAddr: testCase.remoteAddr,
			}
			require.Equal(t, testCase.expectedResult, httputil.GetRequestIP(req))
		})
	}
}

func TestDefaultHttpClient(t *testing.T) {
	client := httputil.DefaultHttpClient()
	require.NotNil(t, client)
	require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.False(t, transport.DisableKeepAlives)
	require.Equal(t, httputil.DefaultMaxConnsPerHost, transport.MaxConnsPerHost)
	require.Equal(t, httputil.DefaultMaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
}

func TestNewHttpClientWithOptions(t *testing.T) {
	customTransport := &http.Transport{
		DisableKeepAlives: false,
		Proxy:             http.ProxyURL(nil),
	}
	testCases := []struct {
		name          string
		options       []httputil.HttpClientOptions
		expectedCheck func(*testing.T, *http.Client)
	}{
		{
			name:    "No options",
			options: nil,
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.False(t, transport.DisableKeepAlives)
			},
		},
		{
			name:    "WithTimeout",
			options: []httputil.HttpClientOptions{httputil.WithTimeout(5 * time.Second)},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.Equal(t, 5*time.Second, client.Timeout)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.False(t, transport.DisableKeepAlives)
			},
		},
		{
			name:    "WithTransport",
			options: []httputil.HttpClientOptions{httputil.WithTransport(customTransport)},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.NotNil(t, client)
				require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)
				require.Same(t, customTransport, client.Transport)
			},
		},
		{
			name: "Multiple options",
			options: []httputil.HttpClientOptions{
				httputil.WithTimeout(15 * time.Second),
				httputil.WithTransport(customTransport),
			},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.NotNil(t, client)
				require.Equal(t, 15*time.Second, client.Timeout)
				require.Same(t, customTransport, client.Transport)
			},
		},
		{
			name: "WithConfig",
			options: []httputil.HttpClientOptions{
				httputil.WithConfig(config.New()),
			},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.NotNil(t, client)
				require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.False(t, transport.DisableKeepAlives)
				require.Equal(t, httputil.DefaultMaxConnsPerHost, transport.MaxConnsPerHost)
				require.Equal(t, httputil.DefaultMaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
				require.Equal(t, httputil.DefaultIdleConnTimeout, transport.IdleConnTimeout)
				require.Equal(t, httputil.DefaultTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
				require.Equal(t, httputil.DefaultExpectContinueTimeout, transport.ExpectContinueTimeout)
				require.Equal(t, httputil.DefaultForceHTTP2, transport.ForceAttemptHTTP2)
			},
		},
		{
			name: "WithConfig and without prefix",
			options: []httputil.HttpClientOptions{
				httputil.WithConfig(func() *config.Config {
					conf := config.New()
					conf.Set("DefaultHttpClient.ForceHTTP2", false)
					return conf
				}()),
			},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.NotNil(t, client)
				require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.False(t, transport.DisableKeepAlives)
				require.Equal(t, httputil.DefaultMaxConnsPerHost, transport.MaxConnsPerHost)
				require.Equal(t, httputil.DefaultMaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
				require.Equal(t, httputil.DefaultIdleConnTimeout, transport.IdleConnTimeout)
				require.Equal(t, httputil.DefaultTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
				require.Equal(t, httputil.DefaultExpectContinueTimeout, transport.ExpectContinueTimeout)
				require.False(t, transport.ForceAttemptHTTP2)
			},
		},
		{
			name: "WithConfig and with prefix",
			options: []httputil.HttpClientOptions{
				httputil.WithConfig(func() *config.Config {
					conf := config.New()
					conf.Set("router.httpclient.ForceHTTP2", false)
					conf.Set("DefaultHttpClient.ForceHTTP2", true)
					return conf
				}(), "router.httpclient"),
			},
			expectedCheck: func(t *testing.T, client *http.Client) {
				require.NotNil(t, client)
				require.Equal(t, httputil.DefaultRequestTimeout, client.Timeout)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.False(t, transport.DisableKeepAlives)
				require.Equal(t, httputil.DefaultMaxConnsPerHost, transport.MaxConnsPerHost)
				require.Equal(t, httputil.DefaultMaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
				require.Equal(t, httputil.DefaultIdleConnTimeout, transport.IdleConnTimeout)
				require.Equal(t, httputil.DefaultTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
				require.Equal(t, httputil.DefaultExpectContinueTimeout, transport.ExpectContinueTimeout)
				require.False(t, transport.ForceAttemptHTTP2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := httputil.NewHttpClient(tc.options...)
			tc.expectedCheck(t, client)
		})
	}
}

func ExampleNewHttpClient() {
	// no need to use .Clone() since a new transport is built each time
	transport := httputil.DefaultTransport()
	transport.ForceAttemptHTTP2 = false
	client := httputil.NewHttpClient(httputil.WithTransport(transport))
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	_ = resp.Body.Close()
}

func ExampleDefaultHttpClient() {
	client := httputil.DefaultHttpClient()
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	_ = resp.Body.Close()
}
