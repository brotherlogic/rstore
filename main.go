package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	ghbpb "github.com/brotherlogic/githubridge/proto"
	pb "github.com/brotherlogic/rstore/proto"

	ghbclient "github.com/brotherlogic/githubridge/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	t1 := time.Now()
	defer func() {
		log.Printf("Read %v in %v", req.GetKey(), time.Since(t1))
	}()
	return s.redisClient.Read(ctx, req)
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

	s.redisClient = &redisClient{rdb: s.rdb}

	mclient, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoAddress))
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err = mclient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	s.mongoClient = &mongoClient{client: mclient}

	err = mclient.Ping(ctx, readpref.Primary())
	if err != nil {
		_, err = s.gclient.CreateIssue(ctx, &ghbpb.CreateIssueRequest{
			User:  "brotherlogic",
			Repo:  "rstore",
			Title: "Mongo Ping Failure",
			Body:  fmt.Sprintf("Error: %v", err),
		})
		if err != nil {
			panic(err)
		}
	}

	cancel()

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

	// Setup prometheus export
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", *metricsPort), nil)
	}()

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("rstore failed to serve: %v", err)
	}
}
