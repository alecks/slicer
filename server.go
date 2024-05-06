package main

import (
	"log/slog"
	"net"
	"os"

	pb "github.com/alecks/slicer/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func serve[T any](conf *slicerConfig, db *bun.DB, stop chan T) {
	listener, err := net.Listen("tcp", conf.Server.Address)
	if err != nil {
		slog.Error("failed to listen on tcp", "err", err)
		os.Exit(1)
	}
	defer listener.Close()

	s := newServer(conf, db)

	go func() {
		slog.Info("starting Slicer", "address", conf.Server.Address)
		if err := s.Serve(listener); err != nil {
			slog.Error("failed to serve gRPC server", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("stopping gracefully")
	s.GracefulStop()
}

func newServer(conf *slicerConfig, db *bun.DB) *grpc.Server {
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authInterceptor), selector.MatchFunc(requireAuth)),
	))

	c := &serviceContext{db: db}

	var err error
	jwtSecret, err = readJwtSecret(conf.Auth.SecretLocation)
	if err != nil {
		slog.Error("failed to read jwt secret; check config/auth.secret_location", "err", err)
	}

	pb.RegisterMetaServiceServer(s, &metaService{})
	pb.RegisterAuthServiceServer(s, &authService{c: c})
	reflection.Register(s)

	return s
}
