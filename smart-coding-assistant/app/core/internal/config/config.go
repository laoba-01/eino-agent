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
	LLM struct {
		Endpoint     string
		APIKey       string
		Model        string
		SystemPrompt map[string]string // "zh"/"en" -> prompt text
	}
}
