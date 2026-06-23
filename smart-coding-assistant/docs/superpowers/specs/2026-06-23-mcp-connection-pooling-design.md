# MCP 客户端连接池化设计

**日期**: 2026-06-23
**状态**: 已确认

## 1. 问题

当前 `ClientManager` 对每个 MCP 端点仅维护一个 `ClientSession`，所有并发工具调用共享同一个 `jsonrpc2.Connection`。SDK 内部的 `writeMu` 互斥锁序列化所有写入操作，导致并发请求相互阻塞。

Tool 服务已开启 `Stateless: true`，每个 session 独立且可互换，天然适合池化。

## 2. 方案

固定大小连接池，启动时对每个端点创建 N 个独立 session，运行时通过 round-robin 分发请求。

### 2.1 池大小

- 默认 **4** 个连接/端点
- 通过 `core.yaml` 的 `MCP.PoolSize` 配置化

### 2.2 选路策略

`atomic.Uint64` 无锁递增 — 无争用，均匀分发，无需维护每个 session 的状态计数。

### 2.3 错误处理

- 单 session 调用失败 → 标记 unhealthy，自动重试下一个（最多遍历全部）
- 全部 session 失败 → 返回错误
- 后台 goroutine 每 30s 对 unhealthy session 做 Ping 探测，恢复后标记 healthy

### 2.4 API 兼容

公开接口 `CallTool`、`ListAllTools`、`Close` 签名不变，调用方（`tools.go`、`agent.go`）零修改。

## 3. 数据结构

```go
type ClientManager struct {
    mu    sync.RWMutex
    pools map[string]*sessionPool  // serverName → pool
}

type sessionPool struct {
    sessions []*pooledSession
    next     atomic.Uint64
    endpoint string
}

type pooledSession struct {
    session *mcp.ClientSession
    client  *mcp.Client
    healthy atomic.Bool
}
```

## 4. 生命周期

```
NewClientManager()
  └→ parseEndpoints()
       └→ 对每个端点 newSessionPool(size=4)
            └→ 循环 4 次: mcp.NewClient() + client.Connect()

CallTool(serverName, toolName, args)
  └→ pool.Acquire()
       └→ idx = next.Add(1) % size
       └→ 若 session[idx].healthy: 返回 session
       └→ 否则线性扫描下一个 healthy session
  └→ session.CallTool()
  └→ 失败则标记 unhealthy，重试下一个

Close()
  └→ 遍历 pools，逐一 session.Close()
```

## 5. 配置

```yaml
# core.yaml
MCP:
  Endpoints: "tool-service=http://localhost:8081/mcp"
  PoolSize: 4
```

```go
// config.go
type McpConfig struct {
    Endpoints string
    PoolSize  int  `json:",default=4"`
}
```

## 6. 影响范围

| 文件 | 变更 |
|---|---|
| `pkg/mcp/client.go` | 重写，引入 `sessionPool` 和 `pooledSession` |
| `app/core/internal/config/config.go` | `McpConfig` 加 `PoolSize` 字段 |
| `app/core/etc/core.yaml` | 加 `PoolSize: 4` |
| 其他文件 | 无变更 |

## 7. 后续迭代

- 连接池指标暴露（Prometheus metrics）：活跃连接数、失败数、等待时间
- 动态扩缩（按需）：若观测到持续高负载再考虑
