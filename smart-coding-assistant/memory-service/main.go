package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

// MemoryServiceServer 实现MemoryService接口
type MemoryServiceServer struct {
	rdb *redis.Client
}

// SaveContext 保存上下文到Redis
func (s *MemoryServiceServer) SaveContext(ctx context.Context, req *SaveContextRequest) (*SaveContextResponse, error) {
	// 这里将实现保存上下文到Redis的逻辑
	return &SaveContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

// GetContext 从Redis获取上下文
func (s *MemoryServiceServer) GetContext(ctx context.Context, req *GetContextRequest) (*GetContextResponse, error) {
	// 这里将实现从Redis获取上下文的逻辑
	return &GetContextResponse{
		Context: make(map[string]string),
		Success: true,
		Error:   "",
	}, nil
}

// DeleteContext 从Redis删除上下文
func (s *MemoryServiceServer) DeleteContext(ctx context.Context, req *DeleteContextRequest) (*DeleteContextResponse, error) {
	// 这里将实现从Redis删除上下文的逻辑
	return &DeleteContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

// UpdateContext 更新Redis中的上下文
func (s *MemoryServiceServer) UpdateContext(ctx context.Context, req *UpdateContextRequest) (*UpdateContextResponse, error) {
	// 这里将实现更新Redis中上下文的逻辑
	return &UpdateContextResponse{
		Success: true,
		Error:   "",
	}, nil
}

func main() {
	// 初始化Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 无密码
		DB:       0,  // 默认DB
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis connection...")
	}

	// 启动gRPC服务器
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterMemoryServiceServer(s, &MemoryServiceServer{rdb: rdb})

	fmt.Println("Memory Service listening on port 50053...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
