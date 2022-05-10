package server

import (
	"context"
	"fmt"
	"github.com/AlexeySadkovich/eldberg/rpc/pb"
	"github.com/AlexeySadkovich/eldberg/rpc/server/controller"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/AlexeySadkovich/eldberg/config"
)

func Register(s *Server, nsController *controller.NodeServiceController) {
	pb.RegisterNodeServiceServer(s.server, nsController)
}

type Server struct {
	server *grpc.Server
	port   int

	logger *zap.SugaredLogger
}

func New(config config.Config, logger *zap.SugaredLogger) *Server {
	rpcServer := grpc.NewServer()
	port := config.GetNodeConfig().Port

	server := &Server{
		server: rpcServer,
		port:   port,
	}

	return server
}

func (s *Server) Start(ctx context.Context) {
	addr := fmt.Sprintf(":%d", s.port)

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorf("listen failed: %w", err)
		return
	}

	if err := s.server.Serve(conn); err != nil {
		s.logger.Errorf("serving failed: %w", err)
		return
	}
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
