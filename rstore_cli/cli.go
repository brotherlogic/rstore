package main

import (
	"log"
	"os"
	"time"

	pbrs "github.com/brotherlogic/rstore/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/brotherlogic/goserver/utils"
)

func main() {
	ctx, cancel := utils.ManualContext("restore-cli", time.Minute)
	defer cancel()

	conn, err := utils.LFDial(os.Args[1])
	if err != nil {
		log.Fatalf("Bad dial: %v", err)
	}

	client := pbrs.NewRStoreServiceClient(conn)

	result, err := client.Read(ctx, &pbrs.ReadRequest{
		Key: "testkey",
	})
	log.Printf("Initial Read: %v, %v", result, err)

	data := "blah"
	res, err := client.Write(ctx, &pbrs.WriteRequest{
		Key:   "testkey",
		Value: &anypb.Any{Value: []byte(data)},
	})
	log.Printf("Initial Write: %v, %v", res, err)

	result, err = client.Read(ctx, &pbrs.ReadRequest{
		Key: "testkey",
	})
	log.Printf("Second Read: %v, %v", result, err)
}
