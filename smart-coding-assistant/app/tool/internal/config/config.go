package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MCP struct {
		Port string
	}
	LLM struct {
		Endpoint string
		APIKey   string
		Model    string
	}
}
