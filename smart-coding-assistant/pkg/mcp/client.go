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
	mu    sync.RWMutex
	pools map[string]*sessionPool
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
