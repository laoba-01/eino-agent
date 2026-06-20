package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	mcpmgr "smart-coding-assistant/pkg/mcp"
)

// Config Agent 配置
type Config struct {
	ChatModel    model.ToolCallingChatModel // 支持 ToolCall 的 ChatModel（DeepSeek 等）
	MaxSteps     int                        // 最大推理步数，默认 12
	SystemPrompt string                     // 系统提示词，定义 Agent 角色和行为
}

// Agent 封装 Eino ReAct Agent
type Agent struct {
	reactAgent   *react.Agent
	mcpClient    *mcpmgr.ClientManager
	systemPrompt string // 系统提示词
}

// New 创建 Agent
func New(ctx context.Context, cfg Config, mcpClient *mcpmgr.ClientManager) (*Agent, error) {
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 12
	}

	einoTools := BuildEinoTools(ctx, mcpClient)

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cfg.ChatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: einoTools,
		},
		MaxStep: cfg.MaxSteps,
	})
	if err != nil {
		return nil, err
	}

	return &Agent{
		reactAgent:   agent,
		mcpClient:    mcpClient,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}

// SetSystemPrompt 设置系统提示词（在 Run 之前调用，按请求语言选择）
func (a *Agent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// Run 执行 Agent：接收用户消息，返回最终响应
func (a *Agent) Run(ctx context.Context, userMessage string) (string, error) {
	msgs := []*schema.Message{
		{Role: schema.User, Content: userMessage},
	}
	// 注入 system prompt（如果配置了）
	if a.systemPrompt != "" {
		msgs = append([]*schema.Message{
			{Role: schema.System, Content: a.systemPrompt},
		}, msgs...)
	}

	result, err := a.reactAgent.Generate(ctx, msgs)
	if err != nil {
		return "", err
	}

	return result.Content, nil
}
