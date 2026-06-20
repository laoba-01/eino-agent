# LLM 统一接管对话设计文档

**日期:** 2026-06-20  
**状态:** 设计中

## 1. 背景

当前系统存在两套处理逻辑：

- **闲聊/系统问题**（"你是谁"、"你好"、"谢谢"等）→ `isSystemQuestion()` + `chatResponse()` 纯规则匹配，返回硬编码文本
- **编程学习问题**（代码报错、语法查询、算法解题等）→ 走 Eino ReAct Agent + DeepSeek，LLM 自主推理 + MCP 工具调用

这种分裂架构导致：
- 闲聊回复生硬、千人一面
- 无法处理规则未覆盖的边界输入
- 网关和核心服务各有一套重复的规则匹配代码

## 2. 目标

移除所有规则匹配分支，所有用户消息统一经过 Eino ReAct Agent（DeepSeek），由 LLM 自主判断是闲聊还是调用工具，生成自然回复。

## 3. 设计方案

### 3.1 Agent System Prompt（`pkg/agent/agent.go`）

Agent 增加 System Prompt 配置，定义角色和行为边界：

```
你是「智能编程学习助手」，一名友好的 AI 编程导师。
- 对于问候、闲聊、自我介绍等问题，自然地回复，不需要调用工具
- 对于代码报错分析、语法查询、算法解题等编程问题，调用对应工具后给出答案
- 回复风格：专业但不生硬，适当使用 emoji
```

- Config 结构体增加 `SystemPrompt string` 字段
- `Agent.Run()` 中，将 system prompt 作为第一条消息注入

### 3.2 ChatLogic 去分支化（`app/core/internal/logic/chatlogic.go`）

**移除：**
- `isSystemQuestion(msgLower string) bool` — 153 行规则匹配函数
- `chatResponse(message string) string` — 硬编码回复函数
- `defaultResponse() string` — 硬编码兜底函数
- `formatToolResult()` — 未使用的工具结果格式化函数
- `Chat()` 方法中开头的 `if isSystemQuestion()` 闲聊快速路径

**简化后 Chat() 流程：**

```
用户消息
  → RAG 语义召回历史（不变）
  → Agent.Run() 统一处理（LLM 自主决策）
  → 异步回存向量记忆（不变）
```

所有消息一视同仁，不再有特殊路径。

### 3.3 网关层去拦截（`app/gateway/internal/logic/chatlogic.go`）

**移除：**
- `handleSystemQuestion(msg string) string` — 网关层规则拦截函数
- `ChatLogic.Handle()` 中对 `handleSystemQuestion()` 的调用及快捷返回逻辑

网关层不再做任何意图判断，所有请求透传到核心服务。

## 4. 受影响文件

| 文件 | 改动类型 | 改动描述 |
|------|----------|----------|
| `pkg/agent/agent.go` | 修改 | Config 增加 `SystemPrompt` 字段，`Run()` 注入 system message |
| `app/core/internal/logic/chatlogic.go` | 修改 | 删除规则匹配函数和闲聊分支，简化 Chat() |
| `app/core/internal/svc/servicecontext.go` | 修改 | Agent 创建时传入 system prompt |
| `app/core/internal/config/config.go` | 修改 | 可选：增加 SystemPrompt 配置项 |
| `app/core/etc/core.yaml` | 修改 | 可选：增加 SystemPrompt 配置 |
| `app/gateway/internal/logic/chatlogic.go` | 修改 | 删除 `handleSystemQuestion()` 及其调用 |

## 5. 不变部分

- Eino ReAct Agent 推理流程不变
- MCP 工具注册/调用机制不变
- RAG 语义记忆召回/存储不变
- gRPC 接口协议不变（`ChatRequest` / `ChatResponse`）
- 前端不变
- 其他服务（auth、tool、memory）完全不受影响

## 6. 风险与缓解

| 风险 | 缓解 |
|------|------|
| 闲聊回复质量依赖 LLM 可用性 | DeepSeek API 已有超时/错误处理，失败时降级为简洁错误提示 |
| LLM 可能对简单问候做工具调用（过度推理） | System Prompt 明确引导"问候闲聊直接回复，不调用工具" |
| 每次闲聊都消耗 token | 闲聊类消息本身短小，token 消耗极低；未来可考虑本地缓存常见问候 |

## 7. 测试要点

- 问候类消息（"你好"、"hi"）→ LLM 自然回复，不调用工具
- 身份类问题（"你是谁"、"你是什么模型"）→ LLM 自我介绍，不调用工具
- 编程问题（"Python 报错 TypeError"、"解释 goroutine"）→ LLM 调用对应 MCP 工具
- 混合问题 → LLM 自主判断是否调用工具
- LLM 不可用时的降级行为
