package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smart-coding-assistant/pkg/llm"

	"github.com/google/uuid"
)

// Planner 计划生成器接口
type Planner interface {
	GeneratePlan(ctx context.Context, userMessage string, historyContext map[string]string, availableTools []string) (*Plan, error)
}

// LLMPlanner 基于大模型生成执行计划
type LLMPlanner struct {
	llmClient *llm.Client
}

func NewLLMPlanner(llmClient *llm.Client) *LLMPlanner {
	return &LLMPlanner{llmClient: llmClient}
}

func (p *LLMPlanner) GeneratePlan(ctx context.Context, userMessage string, historyContext map[string]string, availableTools []string) (*Plan, error) {
	messages := []llm.ChatMessage{
		{Role: "system", Content: buildSystemPrompt(availableTools)},
		{Role: "user", Content: buildUserPrompt(userMessage, historyContext)},
	}

	response, err := p.llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm plan generation: %w", err)
	}

	plan, err := parsePlanJSON(response)
	if err != nil {
		// 重试一次
		messages = append(messages, llm.ChatMessage{
			Role: "user", Content: "你的回复不是有效的 JSON。请严格按 JSON 格式重新输出。",
		})
		response, err = p.llmClient.Chat(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("llm retry: %w", err)
		}
		plan, err = parsePlanJSON(response)
		if err != nil {
			return nil, fmt.Errorf("parse plan after retry: %w", err)
		}
	}

	plan.ID = uuid.New().String()
	plan.CreatedAt = time.Now()
	for i := range plan.Steps {
		plan.Steps[i].Index = i + 1
		plan.Steps[i].Status = string(StepStatusPending)
	}
	return plan, nil
}

func buildSystemPrompt(tools []string) string {
	toolList := strings.Join(tools, ", ")
	if toolList == "" {
		toolList = "（无可用工具）"
	}
	return fmt.Sprintf(
		`你是计划制定器。根据用户需求生成有序执行步骤。

可用工具：%s

严格输出 JSON（不要 markdown 包裹）:
{"goal":"用户目标","steps":[{"description":"步骤描述","tool_name":"工具名或空","tool_args":{}}]}

规则：
- 有序、可独立执行
- tool_name 为空表示纯推理步骤
- tool_name 不为空必须从可用工具列表选`, toolList)
}

func buildUserPrompt(message string, context map[string]string) string {
	if len(context) > 0 {
		ctxJSON, err := json.Marshal(context)
		if err != nil {
			return fmt.Sprintf("用户需求：%s", message)
		}
		return fmt.Sprintf("用户需求：%s\n\n历史上下文：%s", message, string(ctxJSON))
	}
	return fmt.Sprintf("用户需求：%s", message)
}

func parsePlanJSON(raw string) (*Plan, error) {
	raw = strings.TrimSpace(raw)
	if after, ok := strings.CutPrefix(raw, "```json"); ok {
		raw = strings.TrimSuffix(after, "```")
	} else if after, ok := strings.CutPrefix(raw, "```"); ok {
		raw = strings.TrimSuffix(after, "```")
	}
	raw = strings.TrimSpace(raw)

	var plan Plan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w, raw=%s", err, truncate(raw, 200))
	}
	if plan.Goal == "" {
		return nil, fmt.Errorf("plan goal is empty")
	}
	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}
	return &plan, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var _ Planner = (*LLMPlanner)(nil)
