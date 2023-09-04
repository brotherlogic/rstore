package rstore_client

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestTestClient(t *testing.T) {
	client := GetTestClient()

	_, err := client.Write(context.Background(), &pb.WriteRequest{
		Key:   "123",
		Value: &anypb.Any{Value: []byte{1, 2, 3}},
	})
	if err != nil {
		t.Errorf("Bad write: %v", err)
	}

	val, err := client.Read(context.Background(), &pb.ReadRequest{
		Key: "123",
	})
	if err != nil {
		t.Errorf("Bad first read: %v", err)
	}
	if val.GetValue().Value[0] != 1 {
		t.Errorf("Bad val: %v", val)
	}

	keys, err := client.GetKeys(context.Background(), &pb.GetKeysRequest{})
	if err != nil {
		t.Errorf("Bad get keys: %v", err)
	}
	if len(keys.GetKeys()) != 1 || keys.GetKeys()[0] != "123" {
		t.Errorf("Bad keys: %v", keys)
	}

	_, err = client.Delete(context.Background(), &pb.DeleteRequest{
		Key: "123",
	})
	if err != nil {
		t.Errorf("Bad delete: %v", err)
	}

	val, err = client.Read(context.Background(), &pb.ReadRequest{
		Key: "123",
	})
	if err == nil || status.Code(err) != codes.NotFound {
		t.Errorf("Bad final read: %v, %v", val, err)
	}
}

func TestGetKeysSuffix(t *testing.T) {
	client := GetTestClient()

	client.Write(context.Background(), &pb.WriteRequest{
		Key:   "magicpocket",
		Value: &anypb.Any{Value: []byte{1, 2, 3}},
	})
	client.Write(context.Background(), &pb.WriteRequest{
		Key:   "magic",
		Value: &anypb.Any{Value: []byte{1, 2, 3}},
	})

	keys, err := client.GetKeys(context.Background(), &pb.GetKeysRequest{
		Prefix:      "magic",
		AvoidSuffix: []string{"pocket"},
	})
	if err != nil {
		t.Errorf("Bad return: %v", err)
	}

	if len(keys.GetKeys()) == 2 {
		t.Errorf("Should only be one key: %v", keys)
	}
}
