package rstore_client

import (
	"context"
	"log"
	"strings"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type TestClient struct {
	mapper map[string][]byte
}

func GetTestClient() RStoreClient {
	return &TestClient{mapper: make(map[string][]byte)}
}

func (c *TestClient) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	if val, ok := c.mapper[req.GetKey()]; ok {
		return &pb.ReadResponse{Value: &anypb.Any{Value: val}}, nil
	}

	return nil, status.Errorf(codes.NotFound, "Unable to locate %v", req.GetKey())
}

func (c *TestClient) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	c.mapper[req.Key] = req.GetValue().Value
	return &pb.WriteResponse{}, nil
}

func (c *TestClient) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	var keys []string
	log.Printf("Reading %v", c.mapper)
	for key := range c.mapper {
		if strings.HasPrefix(key, req.GetPrefix()) {
			valid := true
			for _, suffix := range req.GetAvoidSuffix() {
				if strings.HasSuffix(key, suffix) {
					valid = false
				}
			}
			if valid {
				keys = append(keys, key)
			}
		}
	}
	return &pb.GetKeysResponse{Keys: keys}, nil
}

func (c *TestClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	delete(c.mapper, req.GetKey())
	return &pb.DeleteResponse{}, nil
}
