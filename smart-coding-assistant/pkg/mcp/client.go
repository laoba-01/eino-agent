package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type ClientManager struct {
	sessions map[string]*mcp.ClientSession
	clients  map[string]*mcp.Client
}

type acceptHeaderRoundTripper struct {
	delegate http.RoundTripper
}

func (a *acceptHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json, text/event-stream")
	return a.delegate.RoundTrip(req)
}

func NewClientManager(ctx context.Context, endpointsStr string) *ClientManager {
	mgr := &ClientManager{
		sessions: make(map[string]*mcp.ClientSession),
		clients:  make(map[string]*mcp.Client),
	}

	if endpointsStr == "" {
		return mgr
	}

	endpoints := parseEndpoints(endpointsStr)
	for name, url := range endpoints {
		client := mcp.NewClient(&mcp.Implementation{
			Name:    "smart-coding-core-service",
			Version: "1.0.0",
		}, nil)

		transport := &mcp.StreamableClientTransport{
			Endpoint: url,
			HTTPClient: &http.Client{
				Transport: &acceptHeaderRoundTripper{delegate: http.DefaultTransport},
			},
		}

		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			log.Printf("警告: 无法连接到 MCP 服务器 %s (%s): %v", name, url, err)
			continue
		}
		mgr.sessions[name] = session
		mgr.clients[name] = client
		log.Printf("已连接到 MCP 服务器: %s (%s)", name, url)
	}

	return mgr
}

func (m *ClientManager) ListAllTools(ctx context.Context) map[string][]*mcp.Tool {
	result := make(map[string][]*mcp.Tool)
	for name, session := range m.sessions {
		resp, err := session.ListTools(ctx, nil)
		if err != nil {
			log.Printf("警告: 获取 %s 的工具列表失败: %v", name, err)
			continue
		}
		result[name] = resp.Tools
	}
	return result
}

func (m *ClientManager) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	session, ok := m.sessions[serverName]
	if !ok {
		return nil, fmt.Errorf("MCP 服务器 %q 未连接", serverName)
	}
	return session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
}

func (m *ClientManager) Close() {
	for name, session := range m.sessions {
		if err := session.Close(); err != nil {
			log.Printf("关闭 MCP 会话 %s 失败: %v", name, err)
		}
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
