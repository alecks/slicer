package main

import (
	"context"

	pb "github.com/alecks/slicer/proto"
)

type metaService struct {
	pb.UnimplementedMetaServiceServer
}

func (s *metaService) Info(ctx context.Context, in *pb.InfoRequest) (*pb.InfoResponse, error) {
	return &pb.InfoResponse{
		Version: slicerVersion,
	}, nil
}
