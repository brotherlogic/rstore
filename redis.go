package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/brotherlogic/rstore/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type redisClient struct {
	rdb *redis.Client
}

func (r *redisClient) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	/*if val, ok := s.cache[req.GetKey()]; ok {
		return &pb.ReadResponse{Value: &anypb.Any{Value: val}}, nil
	}*/
	cmd := r.rdb.Get(ctx, req.GetKey())
	result, err := cmd.Bytes()

	if err == redis.Nil {
		return nil, status.Errorf(codes.NotFound, "key %v was not found", req.GetKey())
	}

	if err != nil {
		log.Printf("remote err on read: %v", err)
	}
	return &pb.ReadResponse{Value: &anypb.Any{Value: result}}, err
}

func (r *redisClient) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	err := r.rdb.Set(ctx, req.GetKey(), req.GetValue().GetValue(), 0).Err()
	if err != nil {
		log.Printf("remote err on write: %v", err)
	} else {
		//s.cache[req.GetKey()] = req.GetValue().GetValue()
	}
	return &pb.WriteResponse{}, err
}

func (r *redisClient) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	var akeys []string
	t := time.Now()
	defer func(t time.Time) {
		log.Printf("Completed %v in %v", req.GetPrefix(), time.Since(t))
	}(t)
	iter := r.rdb.Scan(ctx, 0, fmt.Sprintf("%v*", req.GetPrefix()), 10000).Iterator()

	count := 0
	for iter.Next(ctx) {
		count++
		key := iter.Val()
		if req.GetAllKeys() || strings.Count(key, "/") == strings.Count(req.GetPrefix(), "/") {
			valid := true
			for _, suffix := range req.GetAvoidSuffix() {
				if strings.HasSuffix(key, suffix) {
					valid = false
				}
			}
			if valid {
				akeys = append(akeys, key)
			}
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("Failed to read keys (%v) in %v", req.GetPrefix(), time.Since(t))
		return nil, fmt.Errorf("database error reading keys %w", err)
	}

	log.Printf("returning %v items (%v) filtered from %v (took %v)", len(akeys), req.GetPrefix(), count, time.Since(t))
	return &pb.GetKeysResponse{Keys: akeys}, nil
}

func (r *redisClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return &pb.DeleteResponse{}, r.rdb.Del(ctx, req.GetKey()).Err()
}

func (r *redisClient) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	val, err := r.rdb.Incr(ctx, req.GetCounter()).Result()
	if err != nil {
		return nil, err
	}
	return &pb.CountResponse{Count: val}, nil
}
