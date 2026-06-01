package executor

import (
	"context"
	"fmt"
	"strings"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"smart-coding-assistant/app/core/internal/llm"
	mcpmgr "smart-coding-assistant/app/core/internal/mcp"
	"smart-coding-assistant/app/core/internal/planner"
)

// Executor 计划执行器接口
type Executor interface {
	Execute(ctx context.Context, plan *planner.Plan, reporter StepReporter) error
}

// DefaultExecutor 按序执行步骤
type DefaultExecutor struct {
	mcpClient *mcpmgr.ClientManager
	llmClient *llm.Client
}

func NewDefaultExecutor(mcpClient *mcpmgr.ClientManager, llmClient *llm.Client) *DefaultExecutor {
	return &DefaultExecutor{mcpClient: mcpClient, llmClient: llmClient}
}

func (e *DefaultExecutor) Execute(ctx context.Context, plan *planner.Plan, reporter StepReporter) error {
	if reporter == nil {
		reporter = &noopReporter{}
	}
	for i := range plan.Steps {
		step := &plan.Steps[i]
		step.Status = string(planner.StepStatusRunning)
		reporter.OnStepStart(*step)

		if err := e.executeStep(ctx, plan, step); err != nil {
			step.Status = string(planner.StepStatusFailed)
			step.Error = err.Error()
			reporter.OnStepDone(*step)
			return fmt.Errorf("步骤%d失败: %w", step.Index, err)
		}

		step.Status = string(planner.StepStatusCompleted)
		reporter.OnStepDone(*step)
	}
	reporter.OnAllDone(*plan)
	return nil
}

func (e *DefaultExecutor) executeStep(ctx context.Context, plan *planner.Plan, step *planner.Step) error {
	if step.ToolName != "" {
		return e.executeToolStep(ctx, step)
	}
	return e.executeReasoningStep(ctx, plan, step)
}

func (e *DefaultExecutor) executeToolStep(ctx context.Context, step *planner.Step) error {
	allTools := e.mcpClient.ListAllTools(ctx)
	for serverName, tools := range allTools {
		for _, t := range tools {
			if t.Name == step.ToolName {
				args := make(map[string]any)
				for k, v := range step.ToolArgs {
					args[k] = v
				}
				result, err := e.mcpClient.CallTool(ctx, serverName, step.ToolName, args)
				if err != nil {
					return fmt.Errorf("mcp调用 %s/%s: %w", serverName, step.ToolName, err)
				}
				if result.IsError {
					return fmt.Errorf("mcp工具 %s/%s 执行失败: %s", serverName, step.ToolName, extractText(result))
				}
				step.Result = extractText(result)
				return nil
			}
		}
	}
	return fmt.Errorf("工具 %q 未在任何 MCP 服务器中找到", step.ToolName)
}

func (e *DefaultExecutor) executeReasoningStep(ctx context.Context, plan *planner.Plan, step *planner.Step) error {
	var prevResults strings.Builder
	for i := 0; i < step.Index-1 && i < len(plan.Steps); i++ {
		s := plan.Steps[i]
		if s.Status == string(planner.StepStatusCompleted) && s.Result != "" {
			fmt.Fprintf(&prevResults, "\n步骤%d(%s)结果: %s\n", s.Index, s.Description, s.Result)
		}
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: "你是执行助手。根据计划和前面步骤的结果，完成当前步骤。给出简洁准确的回答。"},
		{Role: "user", Content: fmt.Sprintf("计划目标: %s\n\n当前步骤: %s\n\n前面步骤结果:%s\n\n请完成当前步骤。", plan.Goal, step.Description, prevResults.String())},
	}

	result, err := e.llmClient.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("llm推理: %w", err)
	}
	step.Result = result
	return nil
}

// extractText 从 CallToolResult 提取文本内容
func extractText(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return fmt.Sprintf("%v", result)
}

// noopReporter 空实现
type noopReporter struct{}

func (n *noopReporter) OnStepStart(step planner.Step) {}
func (n *noopReporter) OnStepDone(step planner.Step)  {}
func (n *noopReporter) OnAllDone(plan planner.Plan)   {}

var _ Executor = (*DefaultExecutor)(nil)
