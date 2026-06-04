# 向量化 + 语义记忆 设计文档

日期: 2026-06-04

## 1. 背景与目标

### 当前状态
- Memory 服务已实现向量 CRUD（SaveVector/SearchSimilar/DeleteVector），连接 Milvus
- 但没有 Embedding 模型将文本转为向量
- Core 服务没有接入 Memory RPC，向量能力完全未使用
- 用户聊天历史没有持久化语义记忆

### 目标
补上两个缺失环节：
1. **Embedding 客户端** — 调用 DeepSeek API 将文本转为向量
2. **Core ↔ Memory 连线** — 在聊天流程中嵌入语义记忆

## 2. 架构

```
DeepSeek Embedding API
       ↑
       │ HTTP POST /v1/embeddings
       │
Gateway ──gRPC──→ Core ──gRPC──→ Memory ──→ Milvus (向量)
                     │                       │
                     └──MCP/SSE──→ Tool      └──→ Redis (上下文)
```

### 新增组件

| 组件 | 位置 | 职责 |
|------|------|------|
| EmbeddingClient | `app/core/internal/embedding/client.go` | 调用 DeepSeek API，文本→向量 |
| MemoryRpc 客户端 | `app/core/internal/svc/servicecontext.go` | Core 调用 Memory 的 gRPC 客户端 |

## 3. 数据流

```
用户发消息 "goroutine 怎么用？"
  │
  ├─1─→ embedding.Embed(message) → queryVector ([]float32)
  │
  ├─2─→ memory.SearchSimilar("chat_history", queryVector, topK=3)
  │     返回: [{id: 5, score: 0.92, metadata: {q: "协程怎么通信?", a: "..."}},
  │            {id: 2, score: 0.87, metadata: {q: "channel用法", a: "..."}}]
  │
  ├─3─→ 将召回的历史拼入工具调用的上下文
  │
  ├─4─→ 调用 MCP/Tool 生成回答 (callAnalyzeCodeError / callQuerySyntax / ...)
  │
  └─5─→ memory.SaveVector("chat_history", {
           id: hash(userId+message),
           vector: embedding.Embed(message),
           metadata: {q: message, a: response, user_id: userId, ts: now}
         })
```

### 集合设计

Milvus Collection: `chat_history`
- `id` (int64, PK): hash(userId + message) 生成唯一 ID
- `vector` (FloatVector, dim=由模型决定): 消息的 Embedding
- `metadata` (JSON): `{"q": "用户问题", "a": "回答", "user_id": "xxx", "ts": "2026-06-04T10:00:00Z"}`

## 4. 组件设计

### 4.1 EmbeddingClient

```go
// app/core/internal/embedding/client.go

type EmbeddingClient struct {
    endpoint   string        // https://api.deepseek.com/v1
    apiKey     string
    model      string        // deepseek-chat
    httpClient *http.Client  // 带超时
}

func NewEmbeddingClient(endpoint, apiKey, model string) *EmbeddingClient

// Embed 将单条文本转为向量
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error)

// EmbedBatch 批量转换（用于历史消息）
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
```

**API 调用格式**（OpenAI 兼容）：
```json
POST /v1/embeddings
{
  "model": "deepseek-chat",
  "input": "文本内容"
}
→ { "data": [{ "embedding": [0.123, -0.456, ...] }] }
```

**错误处理**：超时 30s，重试 3 次（指数退避），非 2xx 返回 error。

### 4.2 配置

```yaml
# app/core/etc/core.yaml
Embedding:
  Endpoint: https://api.deepseek.com/v1
  ApiKey: ${DEEPSEEK_API_KEY}
  Model: deepseek-chat
MemoryRpc:
  Endpoints:
  - localhost:50053
```

```go
// app/core/internal/config/config.go
type Config struct {
    zrpc.RpcServerConf
    MCP struct {
        Endpoints string
    }
    Embedding struct {
        Endpoint string
        ApiKey   string
        Model    string
    }
    MemoryRpc zrpc.RpcClientConf
}
```

### 4.3 ServiceContext

```go
type ServiceContext struct {
    Config      config.Config
    MCPClient   *mcp.ClientManager
    Embedding   *embedding.EmbeddingClient       // 新增
    MemoryRpc   memorypb.MemoryServiceClient     // 新增
}
```

### 4.4 ChatLogic 语义记忆集成

在现有 `Chat` 方法中插入记忆步骤：

```go
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
    // ... 现有意图识别逻辑 ...

    // 【新增】语义召回：搜索相似历史
    recalledContext := l.recallSimilarHistory(in.GetMessage())

    // 【新增】将召回的历史合并到上下文
    enrichedCtx := mergeContext(in.GetContext(), recalledContext)

    // 原有：意图路由 + 工具调用
    response := l.handleIntent(intent, message, msgLower, enrichedCtx)

    // 【新增】异步存入向量（不阻塞响应）
    go l.rememberMessage(in.GetUserId(), in.GetMessage(), response)

    return &pb.ChatResponse{Response: response, ...}, nil
}
```

**两个新方法**：
- `recallSimilarHistory(message string) string` — Embedding → SearchSimilar → 格式化为文本上下文
- `rememberMessage(userId, message, response string)` — Embedding → SaveVector

**异步存储**：`rememberMessage` 用 goroutine 异步执行，不影响响应延迟。存储失败只记日志，不抛出错误。

### 4.5 ID 生成

使用 `hash(userId + message)` 作为向量 ID，保证相同消息幂等写入（后写覆盖先写）：

```go
func messageID(userId, message string) int64 {
    h := fnv.New64a()
    h.Write([]byte(userId + "|" + message))
    return int64(h.Sum64())
}
```

## 5. 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `app/core/internal/embedding/client.go` | 新增 | EmbeddingClient 实现 |
| `app/core/internal/svc/servicecontext.go` | 修改 | 加入 EmbeddingClient + MemoryRpc |
| `app/core/internal/config/config.go` | 修改 | 加入 Embedding + MemoryRpc 配置 |
| `app/core/etc/core.yaml` | 修改 | 加配置项 |
| `app/core/internal/logic/chatlogic.go` | 修改 | 加入语义记忆流程 |
| `app/core/core.go` | 修改 | 初始化 embedding 客户端 |

**预估改动量：新增 ~180 行，修改 ~40 行，6 个文件。**

## 6. 无变更部分

- **Memory 服务**: 现有接口（SaveVector/SearchSimilar）已满足需求，无需修改
- **Gateway**: 不需要改动，Chat 接口不变
- **Tool 服务**: 不受影响
- **前端**: 不需要改动
- **Proto 文件**: 不需要修改，现有 `memory_service.proto` 的接口已足够

## 7. 边界与风险

| 场景 | 处理方式 |
|------|----------|
| Embedding API 超时/失败 | `recallSimilarHistory` 返回空字符串，不影响主流程；`rememberMessage` 只记日志 |
| Milvus 不可用 | Memory gRPC 调用失败，但不阻塞聊天响应 |
| 相同消息重复发送 | ID 幂等，SaveVector 自动覆盖旧记录 |
| API Key 未配置 | 启动时打印 warning，语义记忆功能静默降级 |
| 向量维度变化 | 由 API 返回的实际维度决定，自动适配 |
