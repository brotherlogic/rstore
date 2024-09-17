package main

import (
	"log"
	"os"
	"time"

	pbrs "github.com/brotherlogic/rstore/proto"

	"github.com/brotherlogic/goserver/utils"
)

func main() {
	ctx, cancel := utils.ManualContext("restore-cli", time.Hour)
	defer cancel()

	conn, err := utils.LFDial(os.Args[1])
	if err != nil {
		log.Fatalf("Bad dial: %v", err)
	}

	client := pbrs.NewRStoreServiceClient(conn)

result, err := client.GetKeys(ctx, &pbrs.GetKeysRequest{Prefix: "gramophile/taskqueue/"})
if err != nil {
log.Printf("Error: %v", err)
}
log.Printf("%v", result)
}
