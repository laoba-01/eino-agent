# Plan-and-Execute 模式重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Core Service 从简单 tool-use 模式重构为 plan-and-execute 模式（LLM 生成计划 → 逐步执行 → 分步返回结果）

**Architecture:** Core Service 内部新增 planner/、executor/、llm/ 三个模块。Planner 调用 LLM 生成有序步骤列表，Executor 按序执行（调 MCP 工具或 LLM 推理），ChatLogic 负责编排并通过改造后的 ChatResponse 分步返回结果。不改动其他微服务。

**Tech Stack:** Go 1.25.6, go-zero v1.10.1, gRPC + Protobuf, MCP go-sdk v1.6.1

---

## 文件变更总览

| 操作 | 文件 | 职责 |
|------|------|------|
| Create | `app/core/internal/llm/client.go` | LLM API HTTP 客户端，供 Planner 和 Executor 共用 |
| Create | `app/core/internal/planner/types.go` | Plan, Step 数据结构 |
| Create | `app/core/internal/planner/planner.go` | Planner 接口 + LLMPlanner 实现 |
| Create | `app/core/internal/executor/reporter.go` | StepReporter 回调接口 |
| Create | `app/core/internal/executor/executor.go` | Executor 接口 + DefaultExecutor 实现 |
| Modify | `app/core/core.proto` | ChatResponse 增加 PlanInfo, StepResult 消息 |
| Modify | `app/core/pb/core.pb.go` | 手动添加 PlanInfo/StepResult 结构体和 ChatResponse 字段 |
| Modify | `app/core/internal/config/config.go` | 增加 LLM 配置 |
| Modify | `app/core/etc/core.yaml` | 增加 LLM endpoint/key/model 配置 |
| Modify | `app/core/internal/svc/servicecontext.go` | 注入 Planner, Executor |
| Modify | `app/core/internal/logic/chatlogic.go` | 编排 Planner → Executor 流程 |

---

### Task 1: 新增 LLM HTTP 客户端

**Files:**
- Create: `app/core/internal/llm/client.go`

- [ ] **Step 1: 创建 LLM 客户端文件**

```go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 大模型 API HTTP 客户端
type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

// Config LLM 客户端配置
type Config struct {
	Endpoint string
	APIKey   string
	Model    string
	Timeout  time.Duration
}

// ChatMessage LLM 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Client{
		endpoint: cfg.Endpoint,
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := chatRequest{Model: c.model, Messages: messages}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm api status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	var cr chatResponse
	if err := json.Unmarshal(respBytes, &cr); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if cr.Error != nil {
		return "", fmt.Errorf("llm api error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("llm api returned no choices")
	}
	return cr.Choices[0].Message.Content, nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/internal/llm/...
```

- [ ] **Step 3: Commit**

```bash
git add smart-coding-assistant/app/core/internal/llm/client.go
git commit -m "feat(core): 新增 LLM HTTP 客户端模块"
```

---

### Task 2: 新增 Planner 模块（数据结构 + 接口 + LLM 实现）

**Files:**
- Create: `app/core/internal/planner/types.go`
- Create: `app/core/internal/planner/planner.go`

- [ ] **Step 1: 创建 types.go**

```go
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
```

- [ ] **Step 2: 创建 planner.go**

```go
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smart-coding-assistant/app/core/internal/llm"

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
		plan.Steps[i].Status = "pending"
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
		ctxJSON, _ := json.Marshal(context)
		return fmt.Sprintf("用户需求：%s\n\n历史上下文：%s", message, string(ctxJSON))
	}
	return fmt.Sprintf("用户需求：%s", message)
}

func parsePlanJSON(raw string) (*Plan, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimSuffix(raw, "```")
	} else if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
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
```

- [ ] **Step 3: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/internal/planner/...
```

- [ ] **Step 4: Commit**

```bash
git add smart-coding-assistant/app/core/internal/planner/
git commit -m "feat(core): 新增 Planner 模块（LLM 生成执行计划）"
```

---

### Task 3: 新增 Executor 模块

**Files:**
- Create: `app/core/internal/executor/reporter.go`
- Create: `app/core/internal/executor/executor.go`

- [ ] **Step 1: 创建 reporter.go**

```go
package executor

import "smart-coding-assistant/app/core/internal/planner"

// StepReporter 步骤执行回调接口
type StepReporter interface {
	OnStepStart(step planner.Step)
	OnStepDone(step planner.Step)
	OnAllDone(plan planner.Plan)
}
```

- [ ] **Step 2: 创建 executor.go**

```go
package executor

import (
	"context"
	"fmt"

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
		step.Status = "running"
		reporter.OnStepStart(*step)

		if err := e.executeStep(ctx, plan, step); err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			reporter.OnStepDone(*step)
			return fmt.Errorf("步骤%d失败: %w", step.Index, err)
		}

		step.Status = "completed"
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
				args := make(map[string]interface{})
				for k, v := range step.ToolArgs {
					args[k] = v
				}
				result, err := e.mcpClient.CallTool(ctx, serverName, step.ToolName, args)
				if err != nil {
					return fmt.Errorf("mcp调用 %s/%s: %w", serverName, step.ToolName, err)
				}
				step.Result = extractText(result)
				return nil
			}
		}
	}
	return fmt.Errorf("工具 %q 未在任何 MCP 服务器中找到", step.ToolName)
}

func (e *DefaultExecutor) executeReasoningStep(ctx context.Context, plan *planner.Plan, step *planner.Step) error {
	var prevResults string
	for i := 0; i < step.Index-1 && i < len(plan.Steps); i++ {
		s := plan.Steps[i]
		if s.Status == "completed" && s.Result != "" {
			prevResults += fmt.Sprintf("\n步骤%d(%s)结果: %s\n", s.Index, s.Description, s.Result)
		}
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: "你是执行助手。根据计划和前面步骤的结果，完成当前步骤。给出简洁准确的回答。"},
		{Role: "user", Content: fmt.Sprintf("计划目标: %s\n\n当前步骤: %s\n\n前面步骤结果:%s\n\n请完成当前步骤。", plan.Goal, step.Description, prevResults)},
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
```

- [ ] **Step 3: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/internal/executor/...
```

- [ ] **Step 4: Commit**

```bash
git add smart-coding-assistant/app/core/internal/executor/
git commit -m "feat(core): 新增 Executor 模块（逐步执行计划）"
```

---

### Task 4: 修改 Proto 定义和生成代码

**Files:**
- Modify: `app/core/core.proto`
- Modify: `app/core/pb/core.pb.go`

- [ ] **Step 1: 修改 core.proto — 增加 PlanInfo 和 StepResult**

找到 `ChatResponse` message，在其后添加新消息定义：

```protobuf
message ChatResponse {
  string response = 1;
  bool is_finished = 2;
  map<string, string> context = 3;
  PlanInfo plan = 4;
  repeated StepResult steps = 5;
}

message PlanInfo {
  string plan_id = 1;
  string goal = 2;
  int32 total_steps = 3;
  int32 completed_steps = 4;
}

message StepResult {
  int32 index = 1;
  string description = 2;
  string status = 3;
  string result = 4;
  string error = 5;
}
```

操作：在 `app/core/core.proto` 中，将现有的 `ChatResponse` 替换为上面的版本，并在文件末尾追加 `PlanInfo` 和 `StepResult` 消息。

- [ ] **Step 2: 手动修改 core.pb.go — 添加新结构体**

在 `app/core/pb/core.pb.go` 中追加以下内容（放在文件末尾 `func init()` 之前）：

```go
type PlanInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlanId        string `protobuf:"bytes,1,opt,name=plan_id,json=planId,proto3" json:"plan_id,omitempty"`
	Goal          string `protobuf:"bytes,2,opt,name=goal,proto3" json:"goal,omitempty"`
	TotalSteps    int32  `protobuf:"varint,3,opt,name=total_steps,json=totalSteps,proto3" json:"total_steps,omitempty"`
	CompletedSteps int32 `protobuf:"varint,4,opt,name=completed_steps,json=completedSteps,proto3" json:"completed_steps,omitempty"`
}

func (x *PlanInfo) Reset()         { *x = PlanInfo{} }
func (x *PlanInfo) String() string { return protoimpl.X.MessageStringOf(x) }
func (*PlanInfo) ProtoMessage()    {}

func (x *PlanInfo) GetPlanId() string {
	if x != nil { return x.PlanId }
	return ""
}
func (x *PlanInfo) GetGoal() string {
	if x != nil { return x.Goal }
	return ""
}
func (x *PlanInfo) GetTotalSteps() int32 {
	if x != nil { return x.TotalSteps }
	return 0
}
func (x *PlanInfo) GetCompletedSteps() int32 {
	if x != nil { return x.CompletedSteps }
	return 0
}

type StepResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index       int32  `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Result      string `protobuf:"bytes,4,opt,name=result,proto3" json:"result,omitempty"`
	Error       string `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StepResult) Reset()         { *x = StepResult{} }
func (x *StepResult) String() string { return protoimpl.X.MessageStringOf(x) }
func (*StepResult) ProtoMessage()    {}

func (x *StepResult) GetIndex() int32 {
	if x != nil { return x.Index }
	return 0
}
func (x *StepResult) GetDescription() string {
	if x != nil { return x.Description }
	return ""
}
func (x *StepResult) GetStatus() string {
	if x != nil { return x.Status }
	return ""
}
func (x *StepResult) GetResult() string {
	if x != nil { return x.Result }
	return ""
}
func (x *StepResult) GetError() string {
	if x != nil { return x.Error }
	return ""
}
```

- [ ] **Step 3: 修改 core.pb.go — ChatResponse 增加 Plan 和 Steps 字段**

找到 `ChatResponse` 结构体（约在第 24 行附近），在 `Context` 字段后添加：

```go
Plan  *PlanInfo      `protobuf:"bytes,4,opt,name=plan,proto3" json:"plan,omitempty"`
Steps []*StepResult  `protobuf:"bytes,5,rep,name=steps,proto3" json:"steps,omitempty"`
```

同时为 ChatResponse 添加对应的 Getter 方法：

```go
func (x *ChatResponse) GetPlan() *PlanInfo {
	if x != nil { return x.Plan }
	return nil
}
func (x *ChatResponse) GetSteps() []*StepResult {
	if x != nil { return x.Steps }
	return nil
}
```

- [ ] **Step 4: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 5: Commit**

```bash
git add smart-coding-assistant/app/core/core.proto smart-coding-assistant/app/core/pb/core.pb.go
git commit -m "feat(core): ChatResponse 增加 PlanInfo 和 StepResult 支持"
```

---

### Task 5: 修改配置文件和 Config 结构体

**Files:**
- Modify: `app/core/internal/config/config.go`
- Modify: `app/core/etc/core.yaml`

- [ ] **Step 1: 修改 config.go — 增加 LLM 配置**

```go
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MCP struct {
		Endpoints string
	}
	LLM struct {
		Endpoint string
		APIKey   string
		Model    string
	}
}
```

- [ ] **Step 2: 修改 core.yaml — 增加 LLM 配置段**

```yaml
Name: core.rpc
ListenOn: 0.0.0.0:50051

MCP:
  Endpoints: "tool-service=http://localhost:8081/mcp"

LLM:
  Endpoint: "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
  APIKey: ""
  Model: "deepseek-v4-pro"
```

- [ ] **Step 3: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 4: Commit**

```bash
git add smart-coding-assistant/app/core/internal/config/config.go smart-coding-assistant/app/core/etc/core.yaml
git commit -m "feat(core): 配置增加 LLM endpoint/key/model"
```

---

### Task 6: 修改 ServiceContext 注入新依赖

**Files:**
- Modify: `app/core/internal/svc/servicecontext.go`

- [ ] **Step 1: 修改 servicecontext.go**

```go
package svc

import (
	"time"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/llm"
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
```

- [ ] **Step 2: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 3: Commit**

```bash
git add smart-coding-assistant/app/core/internal/svc/servicecontext.go
git commit -m "feat(core): ServiceContext 注入 Planner 和 Executor"
```

---

### Task 7: 重写 ChatLogic 编排 Planner → Executor

**Files:**
- Modify: `app/core/internal/logic/chatlogic.go`

- [ ] **Step 1: 重写 chatlogic.go**

```go
package logic

import (
	"context"
	"time"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/llm"
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
```

- [ ] **Step 2: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 3: Commit**

```bash
git add smart-coding-assistant/app/core/internal/svc/servicecontext.go
git commit -m "feat(core): ServiceContext 注入 Planner 和 Executor"
```

---

### Task 7: 重写 ChatLogic 编排 Planner → Executor

**Files:**
- Modify: `app/core/internal/logic/chatlogic.go`

- [ ] **Step 1: 重写 chatlogic.go**

```go
package logic

import (
	"context"
	"fmt"
	"strings"

	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/planner"
	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	// 1. 收集可用工具列表
	availableTools := l.collectTools()

	// 2. Planner 生成执行计划
	plan, err := l.svcCtx.Planner.GeneratePlan(l.ctx, in.Message, in.Context, availableTools)
	if err != nil {
		return &pb.ChatResponse{
			Response:   fmt.Sprintf("计划生成失败: %v", err),
			IsFinished: true,
		}, nil
	}

	// 3. 构建初始响应（计划概览）
	resp := &pb.ChatResponse{
		IsFinished: false,
		Plan: &pb.PlanInfo{
			PlanId:     plan.ID,
			Goal:       plan.Goal,
			TotalSteps: int32(len(plan.Steps)),
		},
		Context: in.Context,
	}

	// 4. Executor 逐步执行，通过 reporter 累积结果
	reporter := &chatReporter{resp: resp, totalSteps: len(plan.Steps)}
	if err := l.svcCtx.Executor.Execute(l.ctx, plan, reporter); err != nil {
		// 某步骤失败 — 报告已完成步骤 + 失败信息
		resp.Response = fmt.Sprintf("执行终止: %v", err)
		resp.IsFinished = true
		return resp, nil
	}

	// 5. 全部完成 — 构建最终响应
	resp.Response = l.buildFinalResponse(plan)
	resp.IsFinished = true
	return resp, nil
}

// collectTools 收集可用的 MCP 工具名称列表
func (l *ChatLogic) collectTools() []string {
	var tools []string
	if l.svcCtx.MCPClient == nil {
		return tools
	}
	allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
	for _, serverTools := range allTools {
		for _, t := range serverTools {
			tools = append(tools, t.Name)
		}
	}
	return tools
}

// buildFinalResponse 构建最终响应文本
func (l *ChatLogic) buildFinalResponse(plan *planner.Plan) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s\n\n", plan.Goal))
	for _, step := range plan.Steps {
		icon := "✅"
		if step.Status == "failed" {
			icon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s **步骤%d**: %s\n", icon, step.Index, step.Description))
		if step.Result != "" {
			sb.WriteString(fmt.Sprintf("   %s\n\n", step.Result))
		}
		if step.Error != "" {
			sb.WriteString(fmt.Sprintf("   错误: %s\n\n", step.Error))
		}
	}
	return sb.String()
}

// chatReporter 实现 executor.StepReporter，将步骤进度写入 ChatResponse
type chatReporter struct {
	resp        *pb.ChatResponse
	totalSteps  int
}

func (r *chatReporter) OnStepStart(step planner.Step) {
	r.resp.Response = fmt.Sprintf("正在执行步骤 %d/%d: %s", step.Index, r.totalSteps, step.Description)
}

func (r *chatReporter) OnStepDone(step planner.Step) {
	r.resp.Steps = append(r.resp.Steps, &pb.StepResult{
		Index:       int32(step.Index),
		Description: step.Description,
		Status:      step.Status,
		Result:      step.Result,
		Error:       step.Error,
	})
	r.resp.Plan.CompletedSteps = int32(len(r.resp.Steps))
}

func (r *chatReporter) OnAllDone(plan planner.Plan) {
	r.resp.Plan.CompletedSteps = int32(r.totalSteps)
}

var _ executor.StepReporter = (*chatReporter)(nil)
```

- [ ] **Step 2: 验证编译**

```bash
cd smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 3: 确认整个项目编译通过**

```bash
cd smart-coding-assistant && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add smart-coding-assistant/app/core/internal/logic/chatlogic.go
git commit -m "feat(core): ChatLogic 重构为 Planner→Executor 编排模式"
```

---

### Task 8: 最终验证和清理

- [ ] **Step 1: 完整构建**

```bash
cd smart-coding-assistant && go build ./...
```

- [ ] **Step 2: 运行 go vet 检查**

```bash
cd smart-coding-assistant && go vet ./app/core/...
```

- [ ] **Step 3: Commit 最终调整**

```bash
git add -A
git diff --cached --stat
git commit -m "chore: plan-and-execute 重构收尾"
```

---
