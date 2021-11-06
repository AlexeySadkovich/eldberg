package control

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
)

type WSServer struct {
	srv *http.Server
}

func newWSServer(port int) *WSServer {
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
	srv := &http.Server{
		Addr: addr,
	}
	s := &WSServer{
		srv: srv,
	}

	return s
}

func (s *WSServer) run() error {
	return s.srv.ListenAndServe()
}

func (s *WSServer) shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return s.srv.Shutdown(ctx)
}
