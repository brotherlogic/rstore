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

	ghbclient "github.com/brotherlogic/githubridge/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	port         = flag.Int("port", 8080, "The server port.")
	metricsPort  = flag.Int("metrics_port", 8081, "Metrics port")
	redisAddress = flag.String("redis", "redis-server.redis-server:6379", "Redis")

	mongoAddress = flag.String("mongo", "mongodb://localhost:27017", "Connection String")
)

var (
	wCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rstore_wcount",
	}, []string{"client", "code"})

	rCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rstore_rcount",
	}, []string{"code"})
)

type Server struct {
	rdb     *redis.Client
	gclient ghbclient.GithubridgeClient

	redisClient *redisClient
	mongoClient *mongoClient

	cache map[string][]byte
}

type rstore interface {
	Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error)
	Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error)
	GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error)
	Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error)
	Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error)
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	t1 := time.Now()
	defer func() {
		log.Printf("Read %v in %v", req.GetKey(), time.Since(t1))
	}()
	r, err := s.redisClient.Read(ctx, req)
	rCount.With(prometheus.Labels{"code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	return r, err
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	t1 := time.Now()
	defer func() {
		log.Printf("Write %v in %v", req.GetKey(), time.Since(t1))
	}()
	// On the write path, do a fire or forget write into Mongo
	_, merr := s.mongoClient.Write(ctx, req)
	wCount.With(prometheus.Labels{"client": "mongo", "code": fmt.Sprintf("%v", status.Code(merr))}).Inc()
	return s.redisClient.Write(ctx, req)
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	t1 := time.Now()
	defer func() {
		log.Printf("Got Keys %v in %v", req.GetPrefix(), time.Since(t1))
	}()
	return s.redisClient.GetKeys(ctx, req)
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	t1 := time.Now()
	defer func() {
		log.Printf("Delete %v in %v", req.GetKey(), time.Since(t1))
	}()
	return s.redisClient.Delete(ctx, req)
}

func (s *Server) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	return s.redisClient.Count(ctx, req)
}

func main() {
	flag.Parse()

	s := &Server{cache: make(map[string][]byte)}
	client, err := ghbclient.GetClientInternal()
	if err != nil {
		log.Fatalf("Unable to reach GHB")
	}
	s.gclient = client

	s.rdb = redis.NewClient(&redis.Options{
		Addr:     *redisAddress,
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	err = s.rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	// print the db size
	val, err := s.rdb.DBSize(context.Background()).Result()
	log.Printf("Found %v, %v", val, err)

	s.redisClient = &redisClient{rdb: s.rdb}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("rstore failed to listen on the serving port %v: %v", *port, err)
	}
	size := 1024 * 1024 * 1000
	gs := grpc.NewServer(
		grpc.MaxSendMsgSize(size),
		grpc.MaxRecvMsgSize(size),
	)
	pb.RegisterRStoreServiceServer(gs, s)
	log.Printf("rstore is listening on %v", lis.Addr())

	cancel()

	// Setup prometheus export
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", *metricsPort), nil)
	}()

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("rstore failed to serve: %v", err)
	}
}
