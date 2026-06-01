package planner

import "time"

// Plan LLM 生成的执行计划
type Plan struct {
	ID        string    `json:"id"`
	Goal      string    `json:"goal"`
	Steps     []Step    `json:"steps"`
	CreatedAt time.Time `json:"created_at"`
}

// Step 单个执行步骤
type Step struct {
	Index       int               `json:"index"`
	Description string            `json:"description"`
	ToolName    string            `json:"tool_name"` // 空 = 纯 LLM 推理
	ToolArgs    map[string]string `json:"tool_args"`
	Status      string            `json:"status"` // pending | running | completed | failed
	Result      string            `json:"result"`
	Error       string            `json:"error"`
}
