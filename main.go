package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/brotherlogic/rstore/proto"

	"github.com/brotherlogic/goserver/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var (
	port         = flag.Int("port", 8080, "The server port.")
	metricsPort  = flag.Int("metrics_port", 8081, "Metrics port")
	redisAddress = flag.String("redis", "redis-server.redis-server:6379", "Where to find redis")
)

type Server struct {
	rdb *redis.Client
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func main() {
	flag.Parse()

	s := &Server{}
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     *redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := utils.ManualContext("redis", time.Minute)
	err := s.rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}
	cancel()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("rstore failed to listen on the serving port %v: %v", *port, err)
	}
	gs := grpc.NewServer()
	pb.RegisterRStoreServiceServer(gs, s)
	log.Printf("rstore is listening on %v", lis.Addr())

	// Setup prometheus export
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", *metricsPort), nil)
	}()

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("rstore failed to serve: %v", err)
	}
}
