package main

import (
	"log"
	"time"

	pbrs "github.com/brotherlogic/rstore/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/brotherlogic/goserver/utils"
)

func main() {
	ctx, cancel := utils.ManualContext("restore-cli", time.Minute)
	defer cancel()

	conn, err := utils.LFDial("kclust1:32218")
	if err != nil {
		log.Fatalf("Bad dial: %v", err)
	}

	client := pbrs.NewRStoreServiceClient(conn)

	res, err := client.Read(ctx, &pbrs.ReadRequest{Key: "testing"})
	log.Printf("First: %v and %v", res, err)

	data := &pbrs.ReadRequest{Key: "donkey"}
	bytes, _ := proto.Marshal(data)

	client.Write(ctx, &pbrs.WriteRequest{Key: "testing", Value: &anypb.Any{Value: bytes}})

	res, err = client.Read(ctx, &pbrs.ReadRequest{Key: "testing"})
	log.Printf("Second: %v and %v", res, err)
}
