package main

import (
	"log/slog"
	"net"
	"os"

	pb "github.com/alecks/slicer/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func serve[T any](addr string, stop chan T) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to listen on tcp", "address", addr, "err", err)
		os.Exit(1)
	}
	defer listener.Close()

	s := newServer()

	go func() {
		slog.Info("starting Slicer", "address", addr)
		if err := s.Serve(listener); err != nil {
			slog.Error("failed to serve gRPC server", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("stopping gracefully")
	s.GracefulStop()
}

func newServer() *grpc.Server {
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authInterceptor), selector.MatchFunc(requireAuth)),
	))

	pb.RegisterMetaServiceServer(s, &metaService{})
	pb.RegisterAuthServiceServer(s, &authService{})
	reflection.Register(s)

	return s
}
