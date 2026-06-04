package svc

import (
	"context"
	"log"
	"os"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/embedding"
	"smart-coding-assistant/app/core/internal/mcp"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	MCPClient *mcp.ClientManager
	Embedding *embedding.EmbeddingClient
	MemoryRpc memorypb.MemoryServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 解析 Embedding API Key（支持环境变量如 ${DEEPSEEK_API_KEY}）
	apiKey := os.ExpandEnv(c.Embedding.ApiKey)

	var embClient *embedding.EmbeddingClient
	if apiKey != "" && c.Embedding.Endpoint != "" {
		embClient = embedding.NewEmbeddingClient(c.Embedding.Endpoint, apiKey, c.Embedding.Model)
		log.Printf("[Core] Embedding 客户端已初始化 (endpoint=%s, model=%s)", c.Embedding.Endpoint, c.Embedding.Model)
	} else {
		log.Printf("[Core] 警告: Embedding 未配置，语义记忆功能将不可用")
	}

	// 连接 Memory RPC
	memoryConn := zrpc.MustNewClient(c.MemoryRpc)
	memoryRpc := memorypb.NewMemoryServiceClient(memoryConn.Conn())
	log.Printf("[Core] Memory RPC 客户端已连接")

	return &ServiceContext{
		Config:    c,
		MCPClient: mcp.NewClientManager(context.Background(), c.MCP.Endpoints),
		Embedding: embClient,
		MemoryRpc: memoryRpc,
	}
}
