package mcp

import (
	"context"
	"encoding/json"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	toolproto "eino/tool-service/proto"
)

// 工具输入结构体

type AnalyzeCodeErrorArgs struct {
	Code         string `json:"code" jsonschema:"产生错误的源代码"`
	ErrorMessage string `json:"error_message" jsonschema:"错误信息或堆栈跟踪"`
	Language     string `json:"language" jsonschema:"编程语言（如 python, go, javascript）"`
}

// 工具输出结构体

type AnalyzeCodeErrorOutput struct {
	Analysis     string `json:"analysis"`
	SuggestedFix string `json:"suggested_fix"`
	Success      bool   `json:"success"`
}

type QuerySyntaxArgs struct {
	Language string `json:"language" jsonschema:"要查询语法的编程语言"`
	Query    string `json:"query" jsonschema:"要解释的语法概念或关键字（如 async/await, goroutines, decorators）"`
	Context  string `json:"context,omitempty" jsonschema:"可选的附加上下文，说明用户正在做什么"`
}

type QuerySyntaxOutput struct {
	Explanation string `json:"explanation"`
	Example     string `json:"example"`
	Success     bool   `json:"success"`
}

type GenerateProblemSolutionArgs struct {
	Problem    string `json:"problem" jsonschema:"要解决的编程问题描述"`
	Difficulty string `json:"difficulty" jsonschema:"难度级别（easy, medium, hard）"`
	Language   string `json:"language" jsonschema:"解决方案的目标编程语言"`
}

type GenerateProblemSolutionOutput struct {
	Approach    string `json:"approach"`
	Code        string `json:"code"`
	Explanation string `json:"explanation"`
	Success     bool   `json:"success"`
}

// Server 包装 gRPC ToolServiceServer，提供 MCP 工具处理

type Server struct {
	grpcServer toolproto.ToolServiceServer
}

func NewServer(grpcSrv toolproto.ToolServiceServer) *Server {
	return &Server{grpcServer: grpcSrv}
}

func (s *Server) handleAnalyzeCodeError(ctx context.Context, request *mcp.CallToolRequest, args *AnalyzeCodeErrorArgs) (*mcp.CallToolResult, *AnalyzeCodeErrorOutput, error) {
	req := &toolproto.AnalyzeCodeErrorRequest{
		Code:         args.Code,
		ErrorMessage: args.ErrorMessage,
		Language:     args.Language,
	}
	resp, err := s.grpcServer.AnalyzeCodeError(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	output := &AnalyzeCodeErrorOutput{
		Analysis:     resp.Analysis,
		SuggestedFix: resp.SuggestedFix,
		Success:      resp.Success,
	}
	text, _ := json.Marshal(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, output, nil
}

func (s *Server) handleQuerySyntax(ctx context.Context, request *mcp.CallToolRequest, args *QuerySyntaxArgs) (*mcp.CallToolResult, *QuerySyntaxOutput, error) {
	req := &toolproto.QuerySyntaxRequest{
		Language: args.Language,
		Query:    args.Query,
		Context:  args.Context,
	}
	resp, err := s.grpcServer.QuerySyntax(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	output := &QuerySyntaxOutput{
		Explanation: resp.Explanation,
		Example:     resp.Example,
		Success:     resp.Success,
	}
	text, _ := json.Marshal(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, output, nil
}

func (s *Server) handleGenerateProblemSolution(ctx context.Context, request *mcp.CallToolRequest, args *GenerateProblemSolutionArgs) (*mcp.CallToolResult, *GenerateProblemSolutionOutput, error) {
	req := &toolproto.GenerateProblemSolutionRequest{
		Problem:    args.Problem,
		Difficulty: args.Difficulty,
		Language:   args.Language,
	}
	resp, err := s.grpcServer.GenerateProblemSolution(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	output := &GenerateProblemSolutionOutput{
		Approach:    resp.Approach,
		Code:        resp.Code,
		Explanation: resp.Explanation,
		Success:     resp.Success,
	}
	text, _ := json.Marshal(output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}, output, nil
}
