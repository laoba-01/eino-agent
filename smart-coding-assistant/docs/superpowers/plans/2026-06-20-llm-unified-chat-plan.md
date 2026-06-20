# LLM 统一接管对话 实现计划

> **For agentic workers:** 使用 checkboxes 追踪进度，每个 Task 独立可测。

**目标:** 移除所有规则匹配分支，所有用户消息统一经过 Eino ReAct Agent（DeepSeek），LLM 自主判断闲聊或调用工具。

**架构:** Agent 增加 System Prompt 定义角色行为；ChatLogic 删除 ~100 行规则匹配代码，闲聊/编程问题一视同仁走 Agent.Run()；网关层删除 `handleSystemQuestion` 拦截。

**技术栈:** Go 1.25.6, Eino ReAct Agent, DeepSeek v4, go-zero gRPC

## 全局约束

- 不改变 gRPC proto 接口（ChatRequest / ChatResponse 不变）
- 不改变前端
- 不改变 auth / tool / memory 服务
- DeepSeek API 已配置，超时 60s，错误时返回简洁降级提示

---

### Task 1: Agent System Prompt 支持

**文件:**
- 修改: `pkg/agent/agent.go`

**接口:**
- 产出: `Config.SystemPrompt string` 字段
- 产出: `Agent.Run()` 注入 system message 作为第一条消息

- [ ] **Step 1: Config 增加 SystemPrompt 字段，Run() 注入 system message**

`pkg/agent/agent.go` 改动如下：

```go
// Config Agent 配置
type Config struct {
	ChatModel    model.ToolCallingChatModel
	MaxSteps     int
	SystemPrompt string  // 新增：系统提示词，定义 Agent 角色和行为
}
```

`Run()` 方法中，在用户消息前插入 system message：

```go
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
```

同时需要在 Agent 结构体中存储 systemPrompt：

```go
type Agent struct {
	reactAgent   *react.Agent
	mcpClient    *mcpmgr.ClientManager
	systemPrompt string  // 新增
}
```

`New()` 函数中存储：

```go
func New(ctx context.Context, cfg Config, mcpClient *mcpmgr.ClientManager) (*Agent, error) {
	// ... 现有逻辑 ...

	return &Agent{
		reactAgent:   agent,
		mcpClient:    mcpClient,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd d:/agent/smart-coding-assistant && go build ./pkg/agent/...
```

- [ ] **Step 3: Commit**

```bash
git add pkg/agent/agent.go
git commit -m "feat(agent): 支持 SystemPrompt 配置，注入到 ReAct Agent 对话"
```

---

### Task 2: ServiceContext 传入 System Prompt

**文件:**
- 修改: `app/core/internal/svc/servicecontext.go`
- 修改: `app/core/internal/config/config.go`
- 修改: `app/core/etc/core.yaml`

**接口:**
- 消费: `agent.Config.SystemPrompt`（来自 Task 1）
- 产出: `config.Config.LLM.SystemPrompt` 配置项

- [ ] **Step 1: Config 增加 SystemPrompt 字段**

`app/core/internal/config/config.go` 中 LLM 结构体增加字段：

```go
LLM struct {
	Endpoint     string
	APIKey       string
	Model        string
	SystemPrompt string  // 新增：系统提示词
}
```

- [ ] **Step 2: YAML 配置增加 systemPrompt**

`app/core/etc/core.yaml` 增加：

```yaml
LLM:
  Endpoint: "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
  APIKey: ""
  Model: "deepseek-v4-pro"
  SystemPrompt: "你是「智能编程学习助手」，一名友好的 AI 编程导师。对于问候、闲聊、自我介绍等问题，自然地回复，不需要调用工具。对于代码报错分析、语法查询、算法解题等编程问题，调用对应工具后给出答案。回复风格：专业但不生硬，适当使用 emoji。"
```

- [ ] **Step 3: ServiceContext 传入 SystemPrompt**

`app/core/internal/svc/servicecontext.go` 中 `NewServiceContext()` 的 Agent 创建部分，增加 `SystemPrompt`：

```go
einoAgent, err := agent.New(context.Background(), agent.Config{
	ChatModel:    chatModel,
	MaxSteps:     12,
	SystemPrompt: c.LLM.SystemPrompt,  // 新增
}, mcpClient)
```

- [ ] **Step 4: 验证编译通过**

```bash
cd d:/agent/smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 5: Commit**

```bash
git add app/core/internal/config/config.go app/core/etc/core.yaml app/core/internal/svc/servicecontext.go
git commit -m "feat(core): ServiceContext 传入 SystemPrompt 到 Agent"
```

---

### Task 3: ChatLogic 去规则匹配

**文件:**
- 修改: `app/core/internal/logic/chatlogic.go`

**接口:**
- 消费: `svcCtx.Agent.Run()`（已存在，不变）
- 移除: `isSystemQuestion()`, `chatResponse()`, `defaultResponse()`, `formatToolResult()`
- 修改: `Chat()` 方法，移除闲聊快速路径

- [ ] **Step 1: 删除规则匹配函数和闲聊分支，简化 Chat()**

`app/core/internal/logic/chatlogic.go` 完整修改后的文件：

```go
package logic

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"time"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

const memoryCollection = "chat_history"

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

// Chat 是 Agent 的主入口：
//   1. RAG 语义召回历史
//   2. Eino ReAct Agent 自主推理 + MCP 工具调用（统一处理闲聊和编程问题）
//   3. 异步回存对话到向量记忆
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()

	// ========== RAG 语义召回 ==========
	recalledHistory := l.recallSimilarHistory(in.GetUserId(), message)

	// ========== Eino Agent 统一处理 ==========
	agentInput := message
	if recalledHistory != "" {
		agentInput = fmt.Sprintf("【历史上下文】\n%s\n\n【当前问题】\n%s", recalledHistory, message)
	}

	response, err := l.svcCtx.Agent.Run(l.ctx, agentInput)
	if err != nil {
		response = fmt.Sprintf("抱歉，执行过程中出现错误: %v\n\n请稍后重试或更具体地描述你的问题。", err)
	}

	// ========== 异步存储 ==========
	go l.rememberMessage(in.GetUserId(), message, response)

	return &pb.ChatResponse{
		Response:   response,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}

// ==============================
// RAG: Eino Embedder + MemoryRpc（不变）
// ==============================

func (l *ChatLogic) recallSimilarHistory(userId, message string) string {
	if l.svcCtx.Embedder == nil || l.svcCtx.MemoryRpc == nil {
		return ""
	}

	vecs, err := l.svcCtx.Embedder.EmbedStrings(l.ctx, []string{message})
	if err != nil || len(vecs) == 0 {
		if err != nil {
			log.Printf("[Memory] 向量化失败(召回): %v", err)
		}
		return ""
	}

	queryVec := float64sToFloat32s(vecs[0])

	resp, err := l.svcCtx.MemoryRpc.SearchSimilar(l.ctx, &memorypb.SearchSimilarRequest{
		Collection:  memoryCollection,
		QueryVector: queryVec,
		TopK:        3,
	})
	if err != nil || !resp.Success || len(resp.Results) == 0 {
		return ""
	}

	var sb = &strings.Builder{}
	// ... 下面和原代码完全一样 ...
}
```

（注：实际编辑时用 Edit 工具精确删除，这里仅展示目标状态。）

具体需要删除的代码块：
- 第 37-43 行: `if isSystemQuestion(msgLower)` 分支 + `go l.rememberMessage(...)` + `return ...`
- 第 46-47 行: `enrichedCtx := mergeContext(...)` 和 `_ = enrichedCtx`
- 第 61-63 行: 合并 `agentInput` 赋值从两处变为一处
- 第 75-128 行: 删除 `isSystemQuestion`, `chatResponse`, `defaultResponse` 三个函数
- 第 242-275 行: 删除未使用的 `formatToolResult` 函数
- 第 226-236 行: 删除未使用的 `mergeContext` 函数

- [ ] **Step 2: 验证编译通过**

```bash
cd d:/agent/smart-coding-assistant && go build ./app/core/...
```

- [ ] **Step 3: Commit**

```bash
git add app/core/internal/logic/chatlogic.go
git commit -m "refactor(core): 移除规则匹配分支，统一由 Eino Agent 处理所有对话"
```

---

### Task 4: 网关层去拦截

**文件:**
- 修改: `app/gateway/internal/logic/chatlogic.go`

**接口:**
- 移除: `handleSystemQuestion()`
- 修改: `Handle()` 方法，删除网关层拦截调用

- [ ] **Step 1: 删除 handleSystemQuestion 函数及其调用**

`app/gateway/internal/logic/chatlogic.go` 需要删除：

1. 第 59-68 行的拦截调用块：
```go
	// 在网关层拦截系统/闲聊问题，直接返回（避免后端中文编码问题）
	if sysResp := handleSystemQuestion(req.Message); sysResp != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(corepb.ChatResponse{
			Response:   sysResp,
			IsFinished: true,
			Context:    req.Context,
		})
		return
	}
```

2. 第 126-175 行：整个 `handleSystemQuestion()` 函数

删除后 `Handle()` 方法直接调用 CoreRpc：

```go
func (l *ChatLogic) Handle(w http.ResponseWriter, r *http.Request) {
	// ... 方法检查、body 读取、GBK 修复、JSON 解析 不变 ...

	// 所有请求透传到核心服务（不再在网关层拦截）
	resp, err := l.svcCtx.CoreRpc.Chat(r.Context(), &corepb.ChatRequest{
		UserId:  userID,
		Message: req.Message,
		Context: req.Context,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd d:/agent/smart-coding-assistant && go build ./app/gateway/...
```

- [ ] **Step 3: Commit**

```bash
git add app/gateway/internal/logic/chatlogic.go
git commit -m "refactor(gateway): 移除网关层闲聊拦截，所有请求透传到核心服务"
```

---

### Task 5: 全量编译验证

**文件:**
- 无新建或修改

- [ ] **Step 1: 编译全量项目**

```bash
cd d:/agent/smart-coding-assistant && go build ./...
```

预期: 全部编译通过，无错误

- [ ] **Step 2: 确认删除的符号无残留引用**

```bash
cd d:/agent/smart-coding-assistant && grep -r "isSystemQuestion\|handleSystemQuestion\|formatToolResult\|mergeContext" --include="*.go" .
```

预期: 无输出（所有已删除符号无残留引用）

- [ ] **Step 3: 运行项目确认启动正常**

```bash
cd d:/agent/smart-coding-assistant && go run app/core/core.go &
sleep 3 && kill %1
```

预期: 日志显示 "Eino ReAct Agent 已就绪"，无 panic

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: 最终清理，确认 LLM 统一接管编译通过"
```
