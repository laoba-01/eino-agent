package svc

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"smart-coding-assistant/app/memory/internal/config"
)

type ServiceContext struct {
	Config     config.Config
	Redis      *redis.Client
	Milvus     client.Client
	LoadedCols sync.Map // 已加载的 Milvus collection 标记（懒加载用）
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.BizRedis.Addr,
		Password:     c.BizRedis.Password,
		DB:           c.BizRedis.DB,
		PoolSize:     applyDefault(c.BizRedis.PoolSize, 50),
		MinIdleConns: applyDefault(c.BizRedis.MinIdleConns, 10),
		MaxRetries:   applyDefault(c.BizRedis.MaxRetries, 2),
		DialTimeout:  durMs(c.BizRedis.DialTimeout, 1*time.Second),
		ReadTimeout:  durMs(c.BizRedis.ReadTimeout, 500*time.Millisecond),
		WriteTimeout: durMs(c.BizRedis.WriteTimeout, 500*time.Millisecond),
		PoolTimeout:  2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: c.Milvus.Addr,
	})
	if err != nil {
		log.Printf("Warning: Failed to connect to Milvus: %v", err)
		return &ServiceContext{Config: c, Redis: rdb}
	}

	ver, err := milvusClient.GetVersion(ctx)
	if err != nil {
		log.Printf("Warning: Milvus version check failed: %v", err)
	} else {
		log.Printf("Connected to Milvus %s", ver)
	}

	log.Printf("Redis 连接池已初始化 (pool=%d, idle=%d)", rdb.PoolStats().TotalConns, rdb.PoolStats().IdleConns)

	return &ServiceContext{
		Config: c,
		Redis:  rdb,
		Milvus: milvusClient,
	}
}

// applyDefault 返回非零值 val，否则返回 def
func applyDefault(val, def int) int {
	if val > 0 {
		return val
	}
	return def
}

// durMs 毫秒数转 time.Duration, 零值则返回默认值
func durMs(ms int, def time.Duration) time.Duration {
	if ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return def
}
