package mcp

import (
	"net/http"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	toolproto "eino/tool-service/proto"
)

// BuildHTTPHandler 创建 MCP 服务器并注册所有工具，返回 HTTP 处理器
func BuildHTTPHandler(grpcSrv toolproto.ToolServiceServer) (http.Handler, error) {
	mcpSrv := mcp.NewServer(&mcp.Implementation{
		Name:    "smart-coding-tool-service",
		Version: "1.0.0",
	}, nil)

	wrapper := NewServer(grpcSrv)

	mcp.AddTool(mcpSrv, &mcp.Tool{
		Name:        "analyze_code_error",
		Description: "分析代码错误并提供修复建议。传入源代码、错误信息和编程语言，返回详细的根因分析和可操作的修复方案。",
	}, wrapper.handleAnalyzeCodeError)

	mcp.AddTool(mcpSrv, &mcp.Tool{
		Name:        "query_syntax",
		Description: "查询编程语言语法。传入语言名称、语法概念查询词和可选的上下文，返回详细解释和代码示例。",
	}, wrapper.handleQuerySyntax)

	mcp.AddTool(mcpSrv, &mcp.Tool{
		Name:        "generate_problem_solution",
		Description: "为编程问题生成解题方案。传入问题描述、难度级别和目标语言，返回解题思路、实现代码和详细解释。",
	}, wrapper.handleGenerateProblemSolution)

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return mcpSrv },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
		},
	)

	return handler, nil
}
