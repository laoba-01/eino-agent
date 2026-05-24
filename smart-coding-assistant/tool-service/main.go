package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"eino/tool-service/proto"
	toolmcp "eino/tool-service/mcp"

	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type ToolServiceServer struct {
	proto.UnimplementedToolServiceServer
}

func (s *ToolServiceServer) AnalyzeCodeError(ctx context.Context, req *proto.AnalyzeCodeErrorRequest) (*proto.AnalyzeCodeErrorResponse, error) {
	return &proto.AnalyzeCodeErrorResponse{
		Analysis:     "Error analysis placeholder",
		SuggestedFix: "Suggested fix placeholder",
		Success:      true,
	}, nil
}

func (s *ToolServiceServer) QuerySyntax(ctx context.Context, req *proto.QuerySyntaxRequest) (*proto.QuerySyntaxResponse, error) {
	return &proto.QuerySyntaxResponse{
		Explanation: "Syntax explanation placeholder",
		Example:     "Syntax example placeholder",
		Success:     true,
	}, nil
}

func (s *ToolServiceServer) GenerateProblemSolution(ctx context.Context, req *proto.GenerateProblemSolutionRequest) (*proto.GenerateProblemSolutionResponse, error) {
	return &proto.GenerateProblemSolutionResponse{
		Approach:    "Problem approach placeholder",
		Code:        "Solution code placeholder",
		Explanation: "Solution explanation placeholder",
		Success:     true,
	}, nil
}

func main() {
	toolSrv := &ToolServiceServer{}

	// 启动 gRPC 服务
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterToolServiceServer(s, &ToolServiceServer{})

	go func() {
		fmt.Println("Tool Service (gRPC) 监听端口 50052...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC 服务启动失败: %v", err)
		}
	}()

	// 启动 MCP HTTP 服务
	mcpHandler, err := toolmcp.BuildHTTPHandler(toolSrv)
	if err != nil {
		log.Fatalf("构建 MCP 处理器失败: %v", err)
	}
	mcpPort := getEnv("MCP_SERVER_PORT", "8081")
	fmt.Printf("MCP Server 监听端口 %s...\n", mcpPort)
	if err := http.ListenAndServe(":"+mcpPort, mcpHandler); err != nil {
		log.Fatalf("MCP 服务启动失败: %v", err)
	}
}