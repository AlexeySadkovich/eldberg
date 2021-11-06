package control

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
)

type HTTPServer struct {
	srv *http.Server
}

func newHTTPServer(port int) *HTTPServer {
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
	srv := &http.Server{
		Addr: addr,
	}
	s := &HTTPServer{
		srv: srv,
	}

	return s
}

func (s *HTTPServer) run() error {
	return s.srv.ListenAndServe()
}

func (s *HTTPServer) shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return s.srv.Shutdown(ctx)
}
