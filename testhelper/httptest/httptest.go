package httptest

import (
	"fmt"
	"net"
	"net/http"
	nethttptest "net/http/httptest"
)

// NewServer starts a new httptest server that listens on all interfaces, contrary to the standard net/httptest.Server that listens only on localhost.
// This is useful when you want to access the test http server from within a docker container.
func NewServer(handler http.Handler) *Server {
	ts := newUnStartedServer(handler)
	ts.start()
	return ts
}

// Server wraps net/httptest.Server to listen on all network interfaces
type Server struct {
	*nethttptest.Server
}

func (s *Server) start() {
	s.Server.Start()
	_, port, err := net.SplitHostPort(s.Listener.Addr().String())
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to parse listener address: %v", err))
	}
	s.URL = fmt.Sprintf("http://%s:%s", "localhost", port)
}

func newUnStartedServer(handler http.Handler) *Server {
	return &Server{&nethttptest.Server{
		Listener: newListener(),
		Config:   &http.Server{Handler: handler},
	}}
}

func newListener() net.Listener {
	listener, tcpError := net.Listen("tcp", "0.0.0.0:0")
	if tcpError == nil {
		return listener
	}
	listener, tcp6Error := net.Listen("tcp6", "[::]:0")
	if tcp6Error == nil {
		return listener
	}
	panic(fmt.Sprintf("httptest: failed to start listener on a port for tcp (%v) and tcp6 (%v)", tcpError, tcp6Error))
}
