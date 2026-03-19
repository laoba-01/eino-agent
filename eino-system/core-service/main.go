package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// CoreServiceServer 实现CoreService接口
type CoreServiceServer struct{}

// Chat 处理聊天请求
func (s *CoreServiceServer) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 这里将实现与大模型API的交互
	// 以及与工具调用服务和记忆管理服务的交互
	return &ChatResponse{
		Response:   "Hello from Core Service!",
		IsFinished: true,
		Context:    req.Context,
	}, nil
}

// GetHistory 获取聊天历史
func (s *CoreServiceServer) GetHistory(ctx context.Context, req *GetHistoryRequest) (*GetHistoryResponse, error) {
	// 这里将从记忆管理服务获取历史记录
	return &GetHistoryResponse{
		Messages: []*ChatMessage{},
	}, nil
}

func main() {
	// 启动gRPC服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterCoreServiceServer(s, &CoreServiceServer{})

	fmt.Println("Core Service listening on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}