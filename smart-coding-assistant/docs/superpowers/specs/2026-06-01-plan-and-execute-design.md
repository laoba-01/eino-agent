# Plan-and-Execute 模式重构设计文档

**日期**: 2026-06-01  
**状态**: 已确认  
**关联文档**: [[../../需求文档]], [[../../技术文档]]

---

## 1. 概述

将智能编程学习助手 Agent 从当前的简单 tool-use 模式重构为 **plan-and-execute（计划-执行）** 范式。

### 1.1 核心变化

- **之前**：用户消息 → LLM 边推理边调工具 → 返回结果
- **之后**：用户消息 → Planner（LLM 生成计划）→ Executor（逐步执行 + 调工具）→ 流式返回结果

### 1.2 关键决策

| 决策点 | 选择 |
|--------|------|
| 计划生成方式 | LLM 生成 |
| 返回方式 | 分步累积返回（后续可升级为 gRPC streaming + SSE） |
| 失败处理 | 终止并报告 |
| 计划结构 | 简单有序步骤列表 |
| 架构范围 | Core Service 内部模块化，不增加新微服务 |

---

## 2. 数据结构

### 2.1 Plan 和 Step（新增 `internal/planner/types.go`）

```go
// Plan LLM 生成的执行计划
type Plan struct {
    ID        string    `json:"id"`
    Goal      string    `json:"goal"`       // 用户原始目标
    Steps     []Step    `json:"steps"`       // 有序步骤列表
    CreatedAt time.Time `json:"created_at"`
}

// Step 单个执行步骤
type Step struct {
    Index       int               `json:"index"`        // 序号 1-based
    Description string            `json:"description"`   // 给人看的步骤描述
    ToolName    string            `json:"tool_name"`     // 要调用的工具名，空=纯LLM推理
    ToolArgs    map[string]string `json:"tool_args"`     // 工具参数
    Status      string            `json:"status"`        // pending | running | completed | failed
    Result      string            `json:"result"`        // 步骤执行结果
    Error       string            `json:"error"`         // 失败时的错误信息
}
```

**状态流转**：`pending → running → completed` 或 `pending → running → failed`

### 2.2 Proto 变更（修改 `core.proto`）

在 `ChatResponse` 中新增计划相关字段：

```protobuf
message ChatResponse {
  string response = 1;
  bool is_finished = 2;
  map<string, string> context = 3;
  
  // 新增 plan-and-execute 字段
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
  string status = 3;    // completed | failed
  string result = 4;
  string error = 5;     // 失败时填充
}
```

---

## 3. 模块设计

### 3.1 Planner 模块（新增 `internal/planner/`）

```
planner/
├── types.go          # Plan, Step 结构体
└── planner.go        # Planner 接口 + LLMPlanner 实现
```

**Planner 接口**：

```go
type Planner interface {
    GeneratePlan(ctx context.Context, message string, context map[string]string) (*Plan, error)
}
```

**LLMPlanner 工作流程**：

1. 从 Memory Service 读取历史上下文，拼接到 prompt
2. 列出当前可用工具（MCP + gRPC Tool Service），注入 system prompt
3. 构造 prompt，引导 LLM 输出结构化 JSON
4. 调用大模型 API（字节云）
5. 解析 JSON → `Plan` 结构体
6. 将 Plan 保存到 Memory Service（Redis key = `plan:{plan_id}`）

**System Prompt 模板**：

```
你是一个计划制定器。根据用户需求，将任务分解为有序的执行步骤。
可用工具列表：
- AnalyzeCodeError: 分析代码错误，参数 {code, error_message, language}
- QuerySyntax: 语法查询，参数 {language, query, context}
- GenerateProblemSolution: 生成解题方案，参数 {problem, difficulty, language}

输出严格的 JSON 格式，不要包含其他内容：
{"goal": "用户目标简述", "steps": [
  {"description": "步骤描述", "tool_name": "工具名或空", "tool_args": {}},
  ...
]}
规则：
- 步骤应有序、可独立执行
- tool_name 为空表示纯推理步骤（不需要调用工具，LLM 直接回答）
- 每个步骤的描述应清晰具体
```

### 3.2 Executor 模块（新增 `internal/executor/`）

```
executor/
├── executor.go       # Executor 接口 + DefaultExecutor 实现
└── reporter.go       # StepReporter 回调接口
```

**Executor 接口**：

```go
type Executor interface {
    Execute(ctx context.Context, plan *Plan, reporter StepReporter) error
}
```

**StepReporter 回调接口**：

```go
type StepReporter interface {
    OnStepStart(step Step)
    OnStepDone(step Step)
    OnAllDone(plan Plan)
}
```

**DefaultExecutor 执行逻辑**：

对每个 Step（按 Index 顺序）：

1. `step.Status = "running"`，回调 `reporter.OnStepStart(step)`
2. 判断步骤类型：
   - **有 ToolName** → 优先查 MCP Client → 未找到则走 gRPC Tool Service → 均未找到则标记 failed
   - **无 ToolName**（纯推理）→ 构造 prompt（Plan Goal + 当前 Step Description + 前面步骤 Result）→ 调用 LLM
3. 保存结果 → `step.Status = "completed"` 或 `"failed"`
4. 回调 `reporter.OnStepDone(step)`
5. 若 `failed` → **立即终止**，不再执行后续步骤
6. 全 `completed` → 回调 `reporter.OnAllDone(plan)`

### 3.3 ServiceContext 变更

```go
type ServiceContext struct {
    Config    config.Config
    MCPClient *mcp.ClientManager
    Planner   planner.Planner     // 新增
    Executor  executor.Executor   // 新增
}
```

### 3.4 ChatLogic 编排（改造 `internal/logic/chatlogic.go`）

```
Chat() 执行流程:
┌─────────────────────────────────────────┐
│ 1. Planner.GeneratePlan(message, ctx)   │
│    → 返回 PlanInfo（goal, total_steps）  │
├─────────────────────────────────────────┤
│ 2. Executor.Execute(plan, reporter)     │
│    reporter.OnStepDone() 每步触发:       │
│      → 累积 steps 到 ChatResponse       │
│      → 更新 completed_steps 计数         │
├─────────────────────────────────────────┤
│ 3. 任一 Step failed:                    │
│      → is_finished = true               │
│      → 报告失败步骤和已完成步骤           │
│      → 终止                              │
├─────────────────────────────────────────┤
│ 4. 全部完成:                             │
│      → is_finished = true                │
│      → 返回完整 plan + steps             │
│      → 通过 Memory Service 保存上下文     │
└─────────────────────────────────────────┘
```

---

## 4. 文件变更清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 新增 | `app/core/internal/planner/types.go` | Plan, Step 结构体 |
| 新增 | `app/core/internal/planner/planner.go` | Planner 接口 + LLMPlanner |
| 新增 | `app/core/internal/executor/executor.go` | Executor 接口 + DefaultExecutor |
| 新增 | `app/core/internal/executor/reporter.go` | StepReporter 接口 |
| 修改 | `app/core/core.proto` | ChatResponse 增加 PlanInfo, StepResult |
| 修改 | `app/core/pb/core.pb.go` | proto 生成代码 |
| 修改 | `app/core/pb/core_grpc.pb.go` | proto 生成代码 |
| 修改 | `app/core/internal/svc/servicecontext.go` | 注入 Planner, Executor |
| 修改 | `app/core/internal/logic/chatlogic.go` | 编排 Planner + Executor |
| 修改 | `app/core/internal/config/config.go` | 增加 LLM API 配置 |
| 新增 | `app/core/etc/core.yaml` | 配置文件（LLM endpoint/key 等）|

---

## 5. 错误处理

| 场景 | 处理方式 |
|------|----------|
| LLM 生成计划失败 | 返回错误，提示用户重试 |
| LLM 返回非 JSON | 重试 1 次（更严格 prompt），仍失败则返回错误 |
| 工具调用超时 | 标记该步骤 failed，终止后续步骤 |
| 某步骤 LLM 推理失败 | 标记 failed，终止 |
| Memory Service 不可用 | 降级：跳过上下文存取，仅用当前消息 |

---

## 6. 后续迭代（不在本次范围）

- gRPC server-side streaming + SSE 实现真正的流式推送
- 步骤间并行执行（DAG 计划）
- 失败后自动重试 / Replanning
- 计划模板缓存（相似问题复用计划）
