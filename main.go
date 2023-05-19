package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
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
	rdb   *redis.Client
	cache map[string][]byte
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	if val, ok := s.cache[req.GetKey()]; ok {
		return &pb.ReadResponse{Value: &anypb.Any{Value: val}}, nil
	}
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
	} else {
		s.cache[req.GetKey()] = req.GetValue().GetValue()
	}
	return &pb.WriteResponse{}, err
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	keys, err := s.rdb.Keys(ctx, fmt.Sprintf("%v*", req.GetPrefix())).Result()
	if err != nil {
		return nil, err
	}

	var akeys []string
	for _, key := range keys {
		if strings.Count(key, "/") == strings.Count(req.GetPrefix(), "/") {
			akeys = append(akeys, key)
		} else {
			log.Printf("Dropping %v -> %v vs %v (%v)", key, strings.Count(key, "/"), strings.Count(req.GetPrefix(), "/"), req.GetPrefix())
		}
	}

	return &pb.GetKeysResponse{Keys: akeys}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return &pb.DeleteResponse{}, s.rdb.Del(ctx, req.GetKey()).Err()
}

func main() {
	flag.Parse()

	s := &Server{cache: make(map[string][]byte)}
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
