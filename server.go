package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	pb "github.com/alecks/slicer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type metaService struct {
	pb.UnimplementedMetaServiceServer
}

func (s *metaService) Info(ctx context.Context, in *pb.InfoRequest) (*pb.InfoResponse, error) {
	return &pb.InfoResponse{
		Version: slicerVersion,
	}, nil
}

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
	s := grpc.NewServer()

	pb.RegisterMetaServiceServer(s, &metaService{})
	reflection.Register(s)

	return s
}
