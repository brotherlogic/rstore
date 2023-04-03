package rstore_client

import (
	"context"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/grpc"
)

type RStoreClient struct {
	test    bool
	gClient pb.RStoreServiceClient
}

func GetClient() (*RStoreClient, error) {
	conn, err := grpc.Dial("rstore.rstore:80", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &RStoreClient{gClient: pb.NewRStoreServiceClient(conn)}, nil
}

func GetTestClient() (*RStoreClient, error) {
	return &RStoreClient{test: true}, nil
}

func (c *RStoreClient) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	if c.test {
		return &pb.ReadResponse{}, nil
	}

	return c.gClient.Read(ctx, req)
}

func (c *RStoreClient) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	if c.test {
		return &pb.WriteResponse{}, nil
	}

	return c.gClient.Write(ctx, req)
}

func (c *RStoreClient) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	if c.test {
		return &pb.GetKeysResponse{}, nil
	}

	return c.gClient.GetKeys(ctx, req)
}

func (c *RStoreClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if c.test {
		return &pb.DeleteResponse{}, nil
	}

	return c.gClient.Delete(ctx, req)
}
