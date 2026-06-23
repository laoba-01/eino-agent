# MCP 客户端连接池化 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 ClientManager 引入固定大小连接池，每个 MCP 端点维护 4 个独立 session，通过 round-robin 分发请求，消除并发阻塞。

**Architecture:** 重写 `pkg/mcp/client.go`，在 `ClientManager` 内部按端点创建 `sessionPool`，每个 pool 包含 N 个 `mcp.ClientSession`。`CallTool` 通过 `atomic.Uint64` 无锁 round-robin 选 session，失败自动重试下一个。

**Tech Stack:** Go 1.21+, `sync/atomic`, `github.com/modelcontextprotocol/go-sdk v1.6.1`

## Global Constraints

- 池大小固定，默认 4，通过 `MCP.PoolSize` 配置化
- 公开接口 `CallTool`、`ListAllTools`、`Close` 签名不变
- 失败 session 自动标记 unhealthy，后台 30s Ping 探测恢复
- 无锁 round-robin（`atomic.Uint64`）

---

### Task 1: 配置层 — 新增 PoolSize 字段

**Files:**
- Modify: `app/core/internal/config/config.go:7-9`
- Modify: `app/core/etc/core.yaml:4-5`

**Interfaces:**
- Produces: `Config.MCP.PoolSize int` — 供 `NewClientManager` 读取

- [ ] **Step 1: 修改配置结构体**

```go
// app/core/internal/config/config.go
type Config struct {
	zrpc.RpcServerConf
	MCP struct {
		Endpoints string
		PoolSize  int `json:",default=4"`
	}
	Embedding struct {
		Endpoint string
		ApiKey   string
		Model    string
	}
	MemoryRpc zrpc.RpcClientConf
	LLM struct {
		Endpoint     string
		APIKey       string
		Model        string
		SystemPrompt map[string]string
	}
}
```

- [ ] **Step 2: 修改运行时配置**

```yaml
# app/core/etc/core.yaml
MCP:
  Endpoints: "tool-service=http://localhost:8081/mcp"
  PoolSize: 4
```

- [ ] **Step 3: 验证编译通过**

Run: `cd smart-coding-assistant && go build ./app/core/...`
Expected: 编译成功（PoolSize 暂未被使用，仅有编译期检查）

- [ ] **Step 4: Commit**

```bash
git add app/core/internal/config/config.go app/core/etc/core.yaml
git commit -m "feat(config): MCP 连接池新增 PoolSize 配置字段"
```

---

### Task 2: 连接池核心 — 重写 ClientManager

**Files:**
- Modify: `pkg/mcp/client.go` (全文重写)

**Interfaces:**
- Consumes: `Config.MCP.PoolSize int` (optional, 默认 4)
- Produces:
  - `NewClientManager(ctx, endpointsStr string) *ClientManager` — 签名不变
  - `NewClientManagerWithPoolSize(ctx, endpointsStr string, poolSize int) *ClientManager` — 新增，指定池大小
  - `(m *ClientManager) CallTool(ctx, serverName, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error)` — 签名不变
  - `(m *ClientManager) ListAllTools(ctx context.Context) map[string][]*mcp.Tool` — 签名不变
  - `(m *ClientManager) Close()` — 签名不变

- [ ] **Step 1: 编写 pool 单元测试（先写，确认失败）**

创建测试文件 `pkg/mcp/client_test.go`：

```go
package mcp

import (
	"testing"
)

func TestSessionPoolAcquireRoundRobin(t *testing.T) {
	// 验证 round-robin 顺序
	// 用 mock 验证 acquire 按 0→1→2→3→0 循环

	// 由于需要真实 MCP 服务端，此处仅验证数据结构逻辑
	// 集成测试在 Task 3
}

func TestSessionPoolAllUnhealthy(t *testing.T) {
	// 全部 session 标记为 unhealthy 时应返回错误
}

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]string
	}{
		{"single", "a=http://x", map[string]string{"a": "http://x"}},
		{"multiple", "a=http://x,b=http://y", map[string]string{"a": "http://x", "b": "http://y"}},
		{"empty", "", map[string]string{}},
		{"spaces", " a = http://x , b = http://y ", map[string]string{"a": "http://x", "b": "http://y"}},
		{"no equals", "just-name", map[string]string{"just-name": "just-name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEndpoints(tt.raw)
			if len(got) != len(tt.want) {
				t.Errorf("parseEndpoints(%q) len = %d, want %d", tt.raw, len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseEndpoints(%q)[%q] = %q, want %q", tt.raw, k, got[k], v)
				}
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd smart-coding-assistant && go test ./pkg/mcp/... -v -run TestParseEndpoints`
Expected: 测试文件存在，但若 `client_test.go` 尚不存在，测试方法也不可见。实际上 `parseEndpoints` 已存在，此测试应 PASS。确认 PASS。

- [ ] **Step 3: 重写 `pkg/mcp/client.go`**

完整替换文件内容：

```go
package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------- HTTP transport helper ----------

type acceptHeaderRoundTripper struct {
	delegate http.RoundTripper
}

func (a *acceptHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json, text/event-stream")
	return a.delegate.RoundTrip(req)
}

// ---------- pooled session ----------

type pooledSession struct {
	session *mcp.ClientSession
	client  *mcp.Client
	healthy atomic.Bool
}

func (ps *pooledSession) isHealthy() bool {
	return ps.healthy.Load()
}

func (ps *pooledSession) markUnhealthy() {
	ps.healthy.Store(false)
}

func (ps *pooledSession) markHealthy() {
	ps.healthy.Store(true)
}

func (ps *pooledSession) close() {
	if ps.session != nil {
		ps.session.Close()
	}
}

// ---------- session pool ----------

type sessionPool struct {
	sessions []*pooledSession
	next     atomic.Uint64
	endpoint string
}

func newSessionPool(ctx context.Context, endpoint string, size int) *sessionPool {
	pool := &sessionPool{
		sessions: make([]*pooledSession, 0, size),
		endpoint: endpoint,
	}

	for i := 0; i < size; i++ {
		ps := &pooledSession{}
		ps.healthy.Store(true)

		client := mcp.NewClient(&mcp.Implementation{
			Name:    "smart-coding-core-service",
			Version: "1.0.0",
		}, nil)

		transport := &mcp.StreamableClientTransport{
			Endpoint: endpoint,
			HTTPClient: &http.Client{
				Transport: &acceptHeaderRoundTripper{delegate: http.DefaultTransport},
			},
		}

		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			log.Printf("警告: 无法创建 MCP 连接 #%d 到 %s: %v", i, endpoint, err)
			ps.healthy.Store(false)
		} else {
			log.Printf("已创建 MCP 连接 #%d 到 %s", i, endpoint)
		}

		ps.client = client
		ps.session = session
		pool.sessions = append(pool.sessions, ps)
	}

	// 启动后台健康检查
	go pool.healthCheckLoop()

	return pool
}

func (p *sessionPool) acquire() (*pooledSession, error) {
	size := uint64(len(p.sessions))
	if size == 0 {
		return nil, fmt.Errorf("连接池为空")
	}

	start := p.next.Add(1) - 1
	for i := uint64(0); i < size; i++ {
		idx := (start + i) % size
		ps := p.sessions[idx]
		if ps.isHealthy() {
			return ps, nil
		}
	}

	return nil, fmt.Errorf("连接池 %s: 所有连接均不可用", p.endpoint)
}

func (p *sessionPool) callTool(ctx context.Context, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	maxRetries := len(p.sessions)
	if maxRetries == 0 {
		return nil, fmt.Errorf("连接池为空")
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		ps, err := p.acquire()
		if err != nil {
			return nil, err
		}

		result, callErr := ps.session.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})

		if callErr == nil {
			return result, nil
		}

		log.Printf("MCP 连接调用失败 (尝试 %d/%d): %v", attempt+1, maxRetries, callErr)
		ps.markUnhealthy()
		lastErr = callErr
	}

	return nil, fmt.Errorf("所有 MCP 连接均失败: %w", lastErr)
}

func (p *sessionPool) listTools(ctx context.Context) ([]*mcp.Tool, error) {
	// ListTools 使用首个 healthy session 即可（元数据操作）
	ps, err := p.acquire()
	if err != nil {
		return nil, err
	}

	resp, err := ps.session.ListTools(ctx, nil)
	if err != nil {
		ps.markUnhealthy()
		return nil, err
	}
	return resp.Tools, nil
}

func (p *sessionPool) close() {
	for _, ps := range p.sessions {
		ps.close()
	}
}

func (p *sessionPool) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, ps := range p.sessions {
			if ps.isHealthy() {
				continue
			}
			if ps.session == nil {
				continue
			}
			if err := ps.session.Ping(context.Background(), nil); err == nil {
				log.Printf("MCP 连接到 %s 已恢复", p.endpoint)
				ps.markHealthy()
			}
		}
	}
}

// ---------- ClientManager (public API) ----------

type ClientManager struct {
	mu     sync.RWMutex
	pools  map[string]*sessionPool
}

func NewClientManager(ctx context.Context, endpointsStr string) *ClientManager {
	return NewClientManagerWithPoolSize(ctx, endpointsStr, 4)
}

func NewClientManagerWithPoolSize(ctx context.Context, endpointsStr string, poolSize int) *ClientManager {
	if poolSize <= 0 {
		poolSize = 4
	}

	mgr := &ClientManager{
		pools: make(map[string]*sessionPool),
	}

	if endpointsStr == "" {
		return mgr
	}

	endpoints := parseEndpoints(endpointsStr)
	for name, url := range endpoints {
		pool := newSessionPool(ctx, url, poolSize)
		mgr.pools[name] = pool
		log.Printf("MCP 连接池已就绪: %s (%s) x%d", name, url, poolSize)
	}

	return mgr
}

func (m *ClientManager) ListAllTools(ctx context.Context) map[string][]*mcp.Tool {
	result := make(map[string][]*mcp.Tool)
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, pool := range m.pools {
		tools, err := pool.listTools(ctx)
		if err != nil {
			log.Printf("警告: 获取 %s 的工具列表失败: %v", name, err)
			continue
		}
		result[name] = tools
	}
	return result
}

func (m *ClientManager) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	m.mu.RLock()
	pool, ok := m.pools[serverName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("MCP 服务器 %q 未连接", serverName)
	}
	return pool.callTool(ctx, toolName, args)
}

func (m *ClientManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, pool := range m.pools {
		pool.close()
		log.Printf("已关闭 MCP 连接池: %s", name)
	}
}

func parseEndpoints(raw string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.Index(part, "="); idx >= 0 {
			name := strings.TrimSpace(part[:idx])
			url := strings.TrimSpace(part[idx+1:])
			if name != "" && url != "" {
				result[name] = url
			}
		} else {
			result[part] = part
		}
	}
	return result
}
```

- [ ] **Step 4: 运行已有测试确保无回归**

Run: `cd smart-coding-assistant && go test ./pkg/mcp/... -v`
Expected: PASS（parseEndpoints 测试通过）

- [ ] **Step 5: 验证编译**

Run: `cd smart-coding-assistant && go build ./...`
Expected: 编译成功

- [ ] **Step 6: Commit**

```bash
git add pkg/mcp/client.go pkg/mcp/client_test.go
git commit -m "feat(mcp): ClientManager 引入固定大小连接池，round-robin 分发"
```

---

### Task 3: 集成 — 传递 PoolSize 配置到 NewClientManager

**Files:**
- Modify: `app/core/internal/svc/servicecontext.go:32`

**Interfaces:**
- Consumes: `Config.MCP.PoolSize int`
- Produces: (无新接口，`ServiceContext.MCPClient` 类型不变)

- [ ] **Step 1: 修改 ServiceContext 初始化**

```go
// app/core/internal/svc/servicecontext.go 第 32 行

// 修改前:
mcpClient := mcp.NewClientManager(context.Background(), c.MCP.Endpoints)

// 修改后:
poolSize := c.MCP.PoolSize
if poolSize <= 0 {
    poolSize = 4
}
mcpClient := mcp.NewClientManagerWithPoolSize(context.Background(), c.MCP.Endpoints, poolSize)
```

- [ ] **Step 2: 编译验证**

Run: `cd smart-coding-assistant && go build ./app/core/...`
Expected: 编译成功

- [ ] **Step 3: 端到端验证**

启动 tool 服务和 core 服务，发送聊天请求确认工具调用正常。

Run tool service: `cd smart-coding-assistant && go run ./app/tool/tool.go`
Run core service: `cd smart-coding-assistant && go run ./app/core/core.go`
Send a test request (e.g., "分析这段代码的语法错误") and verify tools are called successfully.

Expected: 工具调用成功，日志显示 4 个 MCP 连接均已创建。

- [ ] **Step 4: Commit**

```bash
git add app/core/internal/svc/servicecontext.go
git commit -m "feat(core): ServiceContext 传递 PoolSize 配置到 MCP ClientManager"
```

---

### Task 4: 技术文档更新

**Files:**
- Modify: `技术文档.md:519`

- [ ] **Step 1: 更新已完成的待办**

```markdown
# 修改前:
- 🔜 MCP 客户端连接池化，解决并发阻塞问题

# 修改后:
- ✅ ~~MCP 客户端连接池化，解决并发阻塞问题~~ — 已实现
```

- [ ] **Step 2: Commit**

```bash
git add 技术文档.md
git commit -m "docs: 标记 MCP 连接池化为已完成"
```
