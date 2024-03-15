package httptest_test

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	kithttptest "github.com/rudderlabs/rudder-go-kit/testhelper/httptest"
)

func TestServer(t *testing.T) {
	httpServer := kithttptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	}))
	defer httpServer.Close()

	httpServerParsedURL, err := url.Parse(httpServer.URL)
	require.NoError(t, err)

	_, httpServerPort, err := net.SplitHostPort(httpServerParsedURL.Host)
	require.NoError(t, err)

	var (
		body       []byte
		statusCode int
	)
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://0.0.0.0:" + httpServerPort)
		defer func() { httputil.CloseResponse(resp) }()
		if err == nil {
			statusCode = resp.StatusCode
			body, err = io.ReadAll(resp.Body)
		}
		return err == nil
	}, 5*time.Second, 10*time.Millisecond, "failed to connect to proxy")

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "Hello, world!", string(body))
}

func TestUnStartedServer(t *testing.T) {
	// create a server which is not started
	httpUnStartedServer := kithttptest.NewUnStartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	}))

	var (
		body       []byte
		statusCode int
	)
	require.Never(t, func() bool {
		resp, err := http.Get("http://" + httpUnStartedServer.Listener.Addr().String())
		defer func() { httputil.CloseResponse(resp) }()
		return err == nil
	}, 5*time.Second, 10*time.Millisecond, "connected to an un-started server")

	// start the server now
	httpUnStartedServer.Start()
	defer httpUnStartedServer.Close()

	require.Eventually(t, func() bool {
		resp, err := http.Get(httpUnStartedServer.URL)
		defer func() { httputil.CloseResponse(resp) }()
		if err == nil {
			statusCode = resp.StatusCode
			body, err = io.ReadAll(resp.Body)
		}
		return err == nil
	}, 2*time.Second, 100*time.Millisecond, "failed to connect to server")

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "Hello, world!", string(body))
}
