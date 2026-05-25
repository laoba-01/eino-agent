package svc

import (
	"context"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/mcp"
)

type ServiceContext struct {
	Config    config.Config
	MCPClient *mcp.ClientManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		MCPClient: mcp.NewClientManager(context.Background(), c.MCP.Endpoints),
	}
}
