package rstore_client

import (
	"context"
	"strings"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type TestClient struct {
	mapper map[string][]byte
}

func (c *TestClient) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return &pb.ReadResponse{Value: &anypb.Any{Value: c.mapper[req.GetKey()]}}, nil
}

func (c *TestClient) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	c.mapper[req.Key] = req.GetValue().Value
	return &pb.WriteResponse{}, nil
}

func (c *TestClient) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	var keys []string
	for key := range c.mapper {
		if strings.HasPrefix(key, req.GetPrefix()) {
			keys = append(keys, key)
		}
	}
	return &pb.GetKeysResponse{Keys: keys}, nil
}

func (c *TestClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	delete(c.mapper, req.GetKey())
	return &pb.DeleteResponse{}, nil
}
