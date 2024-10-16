package server

import (
	"fmt"
	"net"

	"github.com/DaanV2/mechanus/server/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type APIServer struct {
	server *grpc.Server
}

func NewApiServer() (*APIServer, error) {
	grpc_port := config.APIServer.GRPC.Port.Value()
	grpc_host := config.APIServer.GRPC.Host.Value()
	grpc_addr := fmt.Sprintf("%s:%d", grpc_host, grpc_port)

	lis, err := net.Listen("tcp", grpc_addr)
	if err != nil {
		return nil, err
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	err = grpcServer.Serve(lis)
	if err != nil {
		return nil, err
	}

	if config.APIServer.GRPC.Reflection.Value() {
		reflection.Register(grpcServer)
	}

	return &APIServer{
		server: grpcServer,
	}, nil
}

func (s *APIServer) GracefulStop() {
	s.server.GracefulStop()
}

func (s *APIServer) Close() error {
	s.server.Stop()
	return nil
}
