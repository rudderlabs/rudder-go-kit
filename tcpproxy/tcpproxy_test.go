package tcpproxy

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
)

func TestProxy(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	}))
	defer httpServer.Close()

	httpServerParsedURL, err := url.Parse(httpServer.URL)
	require.NoError(t, err)

	_, httpServerPort, err := net.SplitHostPort(httpServerParsedURL.Host)
	require.NoError(t, err)

	proxyPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	proxy := &Proxy{
		LocalAddr:  "localhost:" + strconv.Itoa(proxyPort),
		RemoteAddr: "localhost:" + httpServerPort,
	}
	go proxy.Start(t)
	t.Cleanup(proxy.Stop)

	var (
		body       []byte
		statusCode int
	)
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://" + proxy.LocalAddr)
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
