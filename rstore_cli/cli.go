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

<<<<<<< Updated upstream
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
=======
	res, err := client.Read(ctx, &pbrs.ReadRequest{Key: "testing"})
	log.Printf("First: %v and %v", res, err)

	data := &pbrs.ReadRequest{Key: "donkey"}
	bytes, _ := proto.Marshal(data)

	_, err = client.Write(ctx, &pbrs.WriteRequest{Key: "testing", Value: &anypb.Any{Value: bytes}})
	log.Printf("Write: %v", err)

	res, err = client.Read(ctx, &pbrs.ReadRequest{Key: "testing"})
	log.Printf("Second: %v and %v", res, err)
>>>>>>> Stashed changes
}
