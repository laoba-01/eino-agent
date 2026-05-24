package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"eino/core-service/proto"
	coremcp "eino/core-service/mcp"

	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
	mcpClient *coremcp.ClientManager
}

func (s *CoreServiceServer) Chat(ctx context.Context, req *proto.ChatRequest) (*proto.ChatResponse, error) {
	// 收集所有可用 MCP 工具信息
	var toolInfo string
	if s.mcpClient != nil {
		allTools := s.mcpClient.ListAllTools(ctx)
		if len(allTools) > 0 {
			toolInfo = "\n可用 MCP 工具:\n"
			for serverName, tools := range allTools {
				for _, t := range tools {
					toolInfo += fmt.Sprintf("- [%s] %s: %s\n", serverName, t.Name, t.Description)
				}
			}
		}
	}

	return &proto.ChatResponse{
		Response:   "Hello from Core Service!" + toolInfo,
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
	// 初始化 MCP 客户端
	mcpEndpoints := getEnv("MCP_SERVER_ENDPOINTS", "")
	mcpClient := coremcp.NewClientManager(context.Background(), mcpEndpoints)
	defer mcpClient.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterCoreServiceServer(s, &CoreServiceServer{mcpClient: mcpClient})

	fmt.Println("Core Service 监听端口 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}