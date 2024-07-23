package main

import (
	"context"
	"log"
	"time"

	pb "github.com/brotherlogic/rstore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("kclust1:31791",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1000*1024*1024)))
	client := pb.NewRStoreServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	log.Printf("Getting Keys")
	keys, err := client.GetKeys(ctx, &pb.GetKeysRequest{AllKeys: true})

	if err != nil {
		log.Fatalf("Bad keys: %v", err)
	}

	log.Printf("Found %v keys", len(keys.GetKeys()))

	count := 0
	for _, key := range keys.GetKeys() {
		log.Printf("%v", key)
		count++

		if count > 20 {
			return
		}
	}
}
