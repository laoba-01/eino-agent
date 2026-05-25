package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	BizRedis struct {
		Addr string
	}
	JWT struct {
		Secret string
		Expire int64
	}
}
