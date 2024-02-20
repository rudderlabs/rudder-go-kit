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
	ts.Start()
	return ts
}

// Simple net/httptest.Server wrapper
type Server struct {
	*nethttptest.Server
}

func (s *Server) Start() {
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
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}
