package main

import (
	"log"
	"os"
	"time"

	pbrs "github.com/brotherlogic/rstore/proto"
	//"google.golang.org/protobuf/types/known/anypb"

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

	result, err := client.GetKeys(ctx, &pbrs.GetKeysRequest{
		Prefix: "",
	})
	log.Printf("Initial Read: %v, %v", result, err)

}
