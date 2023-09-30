package main

import (
	"log"
	"os"
	"time"

	pbrs "github.com/brotherlogic/rstore/proto"

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

	result, err := client.GetKeys(ctx, &pbrs.GetKeysRequest{})
	if err != nil {
		log.Fatalf("Unable to get keys: %v", err)
	}

	log.Printf("Found %v keys", len(result.GetKeys()))
	for _, key := range result.GetKeys() {
		log.Printf("Key: %v", key)
	}

}
