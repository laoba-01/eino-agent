package svc

import (
	"context"
	"time"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/pkg/llm"
	"smart-coding-assistant/app/core/internal/mcp"
	"smart-coding-assistant/app/core/internal/planner"
)

type ServiceContext struct {
	Config    config.Config
	MCPClient *mcp.ClientManager
	Planner   planner.Planner
	Executor  executor.Executor
}

func NewServiceContext(c config.Config) *ServiceContext {
	mcpClient := mcp.NewClientManager(context.Background(), c.MCP.Endpoints)

	llmClient := llm.NewClient(llm.Config{
		Endpoint: c.LLM.Endpoint,
		APIKey:   c.LLM.APIKey,
		Model:    c.LLM.Model,
		Timeout:  60 * time.Second,
	})

	return &ServiceContext{
		Config:    c,
		MCPClient: mcpClient,
		Planner:   planner.NewLLMPlanner(llmClient),
		Executor:  executor.NewDefaultExecutor(mcpClient, llmClient),
	}
}
