package svc

import (
	"context"
	"log"
	"os"
	"time"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/pkg/agent"
	"smart-coding-assistant/pkg/mcp"

	openaiChat "github.com/cloudwego/eino-ext/components/model/openai"
	openaiEmbed "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	MCPClient    *mcp.ClientManager
	Embedder     embedding.Embedder
	MemoryRpc    memorypb.MemoryServiceClient
	Agent        *agent.Agent
	SystemPrompt map[string]string // 多语言 system prompt
}

func NewServiceContext(c config.Config) *ServiceContext {
	poolSize := c.MCP.PoolSize
	if poolSize <= 0 {
		poolSize = 4
	}
	mcpClient := mcp.NewClientManagerWithPoolSize(context.Background(), c.MCP.Endpoints, poolSize)

	// ===== Eino Embedder（替代手写 embedding client） =====
	var embedder embedding.Embedder
	if apiKey := os.ExpandEnv(c.Embedding.ApiKey); apiKey != "" && c.Embedding.Endpoint != "" {
		emb, err := openaiEmbed.NewEmbedder(context.Background(), &openaiEmbed.EmbeddingConfig{
			APIKey:  apiKey,
			Model:   c.Embedding.Model,
			BaseURL: c.Embedding.Endpoint,
			Timeout: 30 * time.Second,
		})
		if err != nil {
			log.Printf("[Core] 警告: 创建 Embedder 失败: %v", err)
		} else {
			embedder = emb
			log.Printf("[Core] Eino Embedder 已初始化 (endpoint=%s, model=%s)", c.Embedding.Endpoint, c.Embedding.Model)
		}
	} else {
		log.Printf("[Core] 警告: Embedding 未配置，语义记忆功能将不可用")
	}

	// ===== Memory RPC =====
	memoryConn := zrpc.MustNewClient(c.MemoryRpc)
	memoryRpc := memorypb.NewMemoryServiceClient(memoryConn.Conn())
	log.Printf("[Core] Memory RPC 客户端已连接")

	// ===== Eino ChatModel（替代 pkg/llm） =====
	chatModel, err := openaiChat.NewChatModel(context.Background(), &openaiChat.ChatModelConfig{
		APIKey:  c.LLM.APIKey,
		BaseURL: c.LLM.Endpoint,
		Model:   c.LLM.Model,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		log.Printf("[Core] 严重: 创建 ChatModel 失败: %v", err)
		panic(err)
	}
	log.Printf("[Core] Eino ChatModel 已初始化 (endpoint=%s, model=%s)", c.LLM.Endpoint, c.LLM.Model)

	// ===== Eino ReAct Agent（替代 Planner + Executor） =====
	einoAgent, err := agent.New(context.Background(), agent.Config{
		ChatModel: chatModel,
		MaxSteps:  12,
		// SystemPrompt 不再此处固定，由 ChatLogic 运行时按语言选择
	}, mcpClient)
	if err != nil {
		log.Printf("[Core] 严重: 创建 Agent 失败: %v", err)
		panic(err)
	}
	log.Printf("[Core] Eino ReAct Agent 已就绪")

	return &ServiceContext{
		Config:       c,
		MCPClient:    mcpClient,
		Embedder:     embedder,
		MemoryRpc:    memoryRpc,
		Agent:        einoAgent,
		SystemPrompt: c.LLM.SystemPrompt,
	}
}
