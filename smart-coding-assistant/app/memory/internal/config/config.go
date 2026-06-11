package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	BizRedis struct {
		Addr string
	}
	Milvus struct {
		Addr string
	}
	
}
