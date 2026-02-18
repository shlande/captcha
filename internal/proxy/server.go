package proxy

import (
	"captcha/api/captcha/v1"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	grpcServer *grpc.Server
	handler    *Handler
	config     ServerConfig
}

func NewServer(config ServerConfig, handler *Handler) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		handler:    handler,
		config:     config,
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.config.Addr())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	v1.RegisterCaptchaServiceServer(s.grpcServer, s.handler)

	fmt.Printf("gRPC server listening on %s\n", s.config.Addr())
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
