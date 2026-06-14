package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	BizRedis RedisConf
	Milvus   struct {
		Addr string
	}
}

// RedisConf Redis 连接池配置
type RedisConf struct {
	Addr         string
	PoolSize     int    // 连接池最大连接数，默认 10*cpu
	MinIdleConns int    // 最小空闲连接数，预热用，默认 0
	DialTimeout  int    // 拨号超时 (ms)，默认 5000
	ReadTimeout  int    // 读超时 (ms)，默认 3000
	WriteTimeout int    // 写超时 (ms)，默认 3000
	MaxRetries   int    // 命令失败最大重试次数，默认 3
	DB           int    // 数据库编号，默认 0
	Password     string // 密码
}
