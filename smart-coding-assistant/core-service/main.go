package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"eino/core-service/proto"

	"google.golang.org/grpc"
)

type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
}

func (s *CoreServiceServer) Chat(ctx context.Context, req *proto.ChatRequest) (*proto.ChatResponse, error) {
	return &proto.ChatResponse{
		Response:   "Hello from Core Service!",
		IsFinished: true,
		Context:    req.Context,
	}, nil
}

func (s *CoreServiceServer) GetHistory(ctx context.Context, req *proto.GetHistoryRequest) (*proto.GetHistoryResponse, error) {
	return &proto.GetHistoryResponse{
		Messages: []*proto.ChatMessage{},
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterCoreServiceServer(s, &CoreServiceServer{})

	fmt.Println("Core Service listening on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}