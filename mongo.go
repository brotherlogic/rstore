package main

import (
	"context"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Mongo struct {
}

func (m *Mongo) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return &pb.ReadResponse{Value: &anypb.Any{}}, nil
}
