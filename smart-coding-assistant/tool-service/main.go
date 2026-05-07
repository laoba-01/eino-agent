package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"eino/tool-service/proto"

	"google.golang.org/grpc"
)

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
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterToolServiceServer(s, &ToolServiceServer{})

	fmt.Println("Tool Service listening on port 50052...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}