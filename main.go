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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	port         = flag.Int("port", 8080, "The server port.")
	metricsPort  = flag.Int("metrics_port", 8081, "Metrics port")
	redisAddress = flag.String("redis", "redis-server.redis-server:6379", "Redis")
)

type Server struct {
	rdb *redis.Client
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	cmd := s.rdb.Get(ctx, req.GetKey())
	result, err := cmd.Bytes()

	if err == redis.Nil {
		return nil, status.Errorf(codes.NotFound, "key %v was not found", req.GetKey())
	}

	if err != nil {
		log.Printf("remote err on read: %v", err)
	}
	return &pb.ReadResponse{Value: &anypb.Any{Value: result}}, err
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	err := s.rdb.Set(ctx, req.GetKey(), req.GetValue().GetValue(), 0).Err()
	if err != nil {
		log.Printf("remote err on write: %v", err)
	}
	return &pb.WriteResponse{}, err
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	keys, err := s.rdb.Keys(ctx, req.GetSuffix()).Result()
	if err != nil {
		return nil, err
	}

	return &pb.GetKeysResponse{Keys: keys}, nil
}

func main() {
	flag.Parse()

	s := &Server{}
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     *redisAddress,
		Password: "",
		DB:       0,
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
