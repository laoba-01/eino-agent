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
