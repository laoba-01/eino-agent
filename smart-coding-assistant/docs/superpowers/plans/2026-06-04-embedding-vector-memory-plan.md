# 向量化 + 语义记忆 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补上 Embedding 客户端 + Core↔Memory 连线，实现聊天语义记忆

**Architecture:** 新增 `embedding` 包调用 DeepSeek API 做文本向量化；Core 通过 go-zero gRPC 客户端连接 Memory 服务；在 ChatLogic 中插入向量召回→上下文增强→异步存储的语义记忆流程

**Tech Stack:** Go 1.25, go-zero v1.10.1, DeepSeek API (OpenAI 兼容 embedding), Milvus v2.4

---

### Task 1: 创建 EmbeddingClient

**Files:**
- Create: `app/core/internal/embedding/client.go`

- [ ] **Step 1: 写入 EmbeddingClient 实现**

```go
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// EmbeddingClient 调用 OpenAI 兼容的 Embedding API 将文本转为向量
type EmbeddingClient struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// NewEmbeddingClient 创建 Embedding 客户端
// endpoint: API 地址，如 https://api.deepseek.com/v1
// apiKey: API 密钥
// model: 模型名称，如 deepseek-chat
func NewEmbeddingClient(endpoint, apiKey, model string) *EmbeddingClient {
	return &EmbeddingClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed 将单条文本转换为向量
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("embedding API 返回空结果")
	}
	return vecs[0], nil
}

// EmbedBatch 批量将多条文本转换为向量
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Embedding API key 未配置")
	}

	results := make([][]float32, len(texts))

	for i, text := range texts {
		vec, err := c.embedWithRetry(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embedding 第 %d 条失败: %w", i, err)
		}
		results[i] = vec
	}

	return results, nil
}

// embedWithRetry 带重试的单条 embedding 请求
func (c *EmbeddingClient) embedWithRetry(ctx context.Context, text string) ([]float32, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			log.Printf("[Embedding] 重试 %d/3 (等待 %v): %v", attempt, backoff, lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		vec, err := c.doEmbed(ctx, text)
		if err == nil {
			return vec, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("embedding 请求失败(已重试3次): %w", lastErr)
}

// doEmbed 执行单次 embedding API 调用
func (c *EmbeddingClient) doEmbed(ctx context.Context, text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Model: c.model,
		Input: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.endpoint + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态 %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(embResp.Data) == 0 || len(embResp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("API 返回的 embedding 为空")
	}

	return embResp.Data[0].Embedding, nil
}
```

- [ ] **Step 2: 提交**

```bash
git add app/core/internal/embedding/client.go
git commit -m "feat: 新增 EmbeddingClient，支持 DeepSeek API 文本向量化"
```

---

### Task 2: 更新 Core 配置

**Files:**
- Modify: `app/core/internal/config/config.go`
- Modify: `app/core/etc/core.yaml`

- [ ] **Step 1: 更新 Config 结构体**

修改 `app/core/internal/config/config.go`，在 `MCP` 后新增 `Embedding` 和 `MemoryRpc`：

```go
package config

import "github.com/zeromicro/go-zero/zrpc"

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

- [ ] **Step 2: 更新 YAML 配置**

修改 `app/core/etc/core.yaml`：

```yaml
Name: core.rpc
ListenOn: 0.0.0.0:50051

MCP:
  Endpoints: "tool-service=http://localhost:8081/mcp"

Embedding:
  Endpoint: https://api.deepseek.com/v1
  ApiKey: ${DEEPSEEK_API_KEY}
  Model: deepseek-chat

MemoryRpc:
  Endpoints:
  - localhost:50053
```

- [ ] **Step 3: 提交**

```bash
git add app/core/internal/config/config.go app/core/etc/core.yaml
git commit -m "feat: core 配置新增 Embedding 和 MemoryRpc"
```

---

### Task 3: 更新 ServiceContext，接入 MemoryRpc 和 EmbeddingClient

**Files:**
- Modify: `app/core/internal/svc/servicecontext.go`

- [ ] **Step 1: 在 ServiceContext 中新增字段并初始化**

将 `app/core/internal/svc/servicecontext.go` 替换为：

```go
package svc

import (
	"context"
	"log"
	"os"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/embedding"
	"smart-coding-assistant/app/core/internal/mcp"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	MCPClient *mcp.ClientManager
	Embedding *embedding.EmbeddingClient
	MemoryRpc memorypb.MemoryServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 解析 Embedding API Key（环境变量替换）
	apiKey := os.ExpandEnv(c.Embedding.ApiKey)

	var embClient *embedding.EmbeddingClient
	if apiKey != "" && c.Embedding.Endpoint != "" {
		embClient = embedding.NewEmbeddingClient(c.Embedding.Endpoint, apiKey, c.Embedding.Model)
		log.Printf("[Core] Embedding 客户端已初始化 (endpoint=%s, model=%s)", c.Embedding.Endpoint, c.Embedding.Model)
	} else {
		log.Printf("[Core] 警告: Embedding 未配置，语义记忆功能将不可用")
	}

	// 连接 Memory RPC
	memoryConn := zrpc.MustNewClient(c.MemoryRpc)
	memoryRpc := memorypb.NewMemoryServiceClient(memoryConn.Conn())
	log.Printf("[Core] Memory RPC 客户端已连接")

	return &ServiceContext{
		Config:    c,
		MCPClient: mcp.NewClientManager(context.Background(), c.MCP.Endpoints),
		Embedding: embClient,
		MemoryRpc: memoryRpc,
	}
}
```

- [ ] **Step 2: 提交**

```bash
git add app/core/internal/svc/servicecontext.go
git commit -m "feat: ServiceContext 接入 MemoryRpc 和 EmbeddingClient"
```

---

### Task 4: 更新 core.go 入口

**Files:**
- Modify: `app/core/core.go`

- [ ] **Step 1: 增加 Memory 连接关闭**

修改 `app/core/core.go`，在 `defer ctx.MCPClient.Close()` 后不新增额外的 Close（go-zero 的 `zrpc.MustNewClient` 返回的连接由框架管理生命周期，无需手动 Close）。

当前代码已满足需求，仅需确认编译通过。无需修改。

- [ ] **Step 2: 验证编译**

```bash
cd app/core && go build -o /dev/null .
```

预期：编译成功（如有缺失 import 会在 Task 3 中解决）。

- [ ] **Step 3: 提交**

```bash
# 如无修改则跳过
```

---

### Task 5: ChatLogic 集成语义记忆

**Files:**
- Modify: `app/core/internal/logic/chatlogic.go`

- [ ] **Step 1: 添加 import 和辅助函数**

在 `app/core/internal/logic/chatlogic.go` 顶部新增 import：

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"time"

	memorypb "smart-coding-assistant/app/memory/pb"
	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)
```

- [ ] **Step 2: 修改 Chat 方法，插入语义记忆步骤**

修改 `Chat` 方法（替换原有的第 32-54 行）：

```go
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()
	msgLower := strings.ToLower(message)

	// 优先处理系统/闲聊问题（在 detectIntent 之前兜底）
	if isSystemQuestion(msgLower) {
		response := l.chatResponse(message)
		// 异步保存闲聊到记忆
		go l.rememberMessage(in.GetUserId(), message, response)
		return &pb.ChatResponse{
			Response:   response,
			IsFinished: true,
			Context:    in.Context,
		}, nil
	}

	// 语义召回：搜索相似历史对话
	recalledContext := l.recallSimilarHistory(in.GetUserId(), message)

	// 合并召回的历史到上下文
	enrichedCtx := mergeContext(in.GetContext(), recalledContext)

	intent := detectIntent(msgLower)

	response := l.handleIntent(intent, message, msgLower, enrichedCtx)

	// 异步存入向量记忆
	go l.rememberMessage(in.GetUserId(), message, response)

	return &pb.ChatResponse{
		Response:   response,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}
```

- [ ] **Step 3: 新增 recallSimilarHistory 方法**

在文件末尾（`extractLangAndQuery` 函数之后）新增：

```go
// === 语义记忆 ===

const memoryCollection = "chat_history"

// recallSimilarHistory 搜索与当前消息语义相似的历史对话
// 失败时返回空字符串，不影响主流程
func (l *ChatLogic) recallSimilarHistory(userId, message string) string {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return ""
	}

	// 将用户消息向量化
	queryVec, err := l.svcCtx.Embedding.Embed(l.ctx, message)
	if err != nil {
		log.Printf("[Memory] 向量化失败(召回): %v", err)
		return ""
	}

	// 搜索相似历史
	resp, err := l.svcCtx.MemoryRpc.SearchSimilar(l.ctx, &memorypb.SearchSimilarRequest{
		Collection:  memoryCollection,
		QueryVector: queryVec,
		TopK:        3,
	})
	if err != nil {
		log.Printf("[Memory] 相似搜索失败: %v", err)
		return ""
	}

	if !resp.Success || len(resp.Results) == 0 {
		return ""
	}

	// 格式化历史为上下文文本
	var sb strings.Builder
	sb.WriteString("【相关历史对话】\n")
	for i, r := range resp.Results {
		q := r.Metadata["q"]
		a := r.Metadata["a"]
		if q == "" {
			continue
		}
		fmt.Fprintf(&sb, "%d. 问: %s\n", i+1, q)
		if a != "" {
			fmt.Fprintf(&sb, "   答: %s\n", a)
		}
	}
	return sb.String()
}

// rememberMessage 异步将消息和回答存入向量记忆
func (l *ChatLogic) rememberMessage(userId, message, response string) {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return
	}

	// 使用独立 context，避免请求结束后被取消
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vec, err := l.svcCtx.Embedding.Embed(ctx, message)
	if err != nil {
		log.Printf("[Memory] 向量化失败(存储): %v", err)
		return
	}

	id := messageID(userId, message)
	_, err = l.svcCtx.MemoryRpc.SaveVector(ctx, &memorypb.SaveVectorRequest{
		Collection: memoryCollection,
		Vectors: []*memorypb.VectorData{
			{
				Id:     id,
				Vector: vec,
				Metadata: map[string]string{
					"q":       message,
					"a":       response,
					"user_id": userId,
					"ts":      time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	})
	if err != nil {
		log.Printf("[Memory] 存储向量失败: %v", err)
	}
}

// messageID 生成幂等的向量 ID
func messageID(userId, message string) int64 {
	h := fnv.New64a()
	h.Write([]byte(userId + "|" + message))
	return int64(h.Sum64())
}

// mergeContext 合并原始上下文和召回的历史上下文
func mergeContext(original map[string]string, recalled string) map[string]string {
	if recalled == "" {
		return original
	}
	merged := make(map[string]string, len(original)+1)
	for k, v := range original {
		merged[k] = v
	}
	merged["history"] = recalled
	return merged
}
```

- [ ] **Step 4: 验证编译**

```bash
cd app/core && go build -o /dev/null .
```

预期：编译成功。

- [ ] **Step 5: 提交**

```bash
git add app/core/internal/logic/chatlogic.go
git commit -m "feat: ChatLogic 集成语义记忆（向量召回 + 异步存储）"
```

---

### Task 6: 设置环境变量并验证

**Files:**
- Modify: `docker-compose.yml` (添加 DEEPSEEK_API_KEY 环境变量)

- [ ] **Step 1: 在 docker-compose.yml 中给 core-rpc 加环境变量**

在 `core-rpc` 的 `environment` 中添加：

```yaml
  core-rpc:
    build:
      context: .
      target: core-rpc
    environment:
      MCP_ENDPOINTS: "tool-service=http://tool-rpc:8081/mcp"
      DEEPSEEK_API_KEY: ${DEEPSEEK_API_KEY}
    ports:
      - "50051:50051"
    depends_on:
      - etcd
    restart: unless-stopped
```

- [ ] **Step 2: 设置本地环境变量并验证整体编译**

```bash
export DEEPSEEK_API_KEY="sk-ae73d219e2aa48d98beef3c54588e8fd"
cd app/core && go build -o /dev/null .
```

预期：编译成功。

- [ ] **Step 3: 全量编译验证**

```bash
cd D:/agent/smart-coding-assistant
go build ./...
```

预期：所有服务编译通过。

- [ ] **Step 4: 提交**

```bash
git add docker-compose.yml
git commit -m "chore: docker-compose 添加 DEEPSEEK_API_KEY 环境变量"
```

---

## 边界场景处理

| 场景 | 行为 |
|------|------|
| `DEEPSEEK_API_KEY` 环境变量未设置 | Embedding 客户端为 nil，语义记忆功能静默降级，聊天不受影响 |
| Embedding API 超时/错误 | `recallSimilarHistory` 返回空 → 不影响响应；`rememberMessage` 仅记日志 |
| Memory RPC 不可用 | SearchSimilar/SaveVector 调用失败，记录日志，不影响聊天 |
| 相同消息重复发送 | FNV hash 生成相同 ID，SaveVector 幂等覆盖 |
| Milvus collection 不存在 | Memory 服务的 `ensureCollection` 自动创建 |
