package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"eino/memory-service/proto"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

type MemoryServiceServer struct {
	proto.UnimplementedMemoryServiceServer
	rdb *redis.Client
}

func (s *MemoryServiceServer) SaveContext(ctx context.Context, req *proto.SaveContextRequest) (*proto.SaveContextResponse, error) {
	return &proto.SaveContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *MemoryServiceServer) GetContext(ctx context.Context, req *proto.GetContextRequest) (*proto.GetContextResponse, error) {
	return &proto.GetContextResponse{
		Context: make(map[string]string),
		Success: true,
		Error:   "",
	}, nil
}

func (s *MemoryServiceServer) DeleteContext(ctx context.Context, req *proto.DeleteContextRequest) (*proto.DeleteContextResponse, error) {
	return &proto.DeleteContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (s *MemoryServiceServer) UpdateContext(ctx context.Context, req *proto.UpdateContextRequest) (*proto.UpdateContextResponse, error) {
	return &proto.UpdateContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis connection...")
	}

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterMemoryServiceServer(s, &MemoryServiceServer{rdb: rdb})

	fmt.Println("Memory Service listening on port 50053...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}