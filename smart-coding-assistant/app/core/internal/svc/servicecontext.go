package svc

import (
	"context"
	"log"
	"os"
	"time"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/embedding"
	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/mcp"
	"smart-coding-assistant/app/core/internal/planner"
	"smart-coding-assistant/pkg/llm"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	MCPClient *mcp.ClientManager
	Embedding *embedding.EmbeddingClient
	MemoryRpc memorypb.MemoryServiceClient
	Planner   planner.Planner
	Executor  executor.Executor
}

func NewServiceContext(c config.Config) *ServiceContext {
	mcpClient := mcp.NewClientManager(context.Background(), c.MCP.Endpoints)

	// Embedding 客户端（语义记忆）
	apiKey := os.ExpandEnv(c.Embedding.ApiKey)
	var embClient *embedding.EmbeddingClient
	if apiKey != "" && c.Embedding.Endpoint != "" {
		embClient = embedding.NewEmbeddingClient(c.Embedding.Endpoint, apiKey, c.Embedding.Model)
		log.Printf("[Core] Embedding 客户端已初始化 (endpoint=%s, model=%s)", c.Embedding.Endpoint, c.Embedding.Model)
	} else {
		log.Printf("[Core] 警告: Embedding 未配置，语义记忆功能将不可用")
	}

	// Memory RPC 客户端
	memoryConn := zrpc.MustNewClient(c.MemoryRpc)
	memoryRpc := memorypb.NewMemoryServiceClient(memoryConn.Conn())
	log.Printf("[Core] Memory RPC 客户端已连接")

	// LLM 客户端（Planner + Executor）
	llmClient := llm.NewClient(llm.Config{
		Endpoint: c.LLM.Endpoint,
		APIKey:   c.LLM.APIKey,
		Model:    c.LLM.Model,
		Timeout:  60 * time.Second,
	})

	return &ServiceContext{
		Config:    c,
		MCPClient: mcpClient,
		Embedding: embClient,
		MemoryRpc: memoryRpc,
		Planner:   planner.NewLLMPlanner(llmClient),
		Executor:  executor.NewDefaultExecutor(mcpClient, llmClient),
	}
}
