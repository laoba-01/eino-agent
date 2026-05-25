package svc

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"smart-coding-assistant/app/memory/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	Milvus client.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BizRedis.Addr,
		Password: "",
		DB:       0,
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

	return &ServiceContext{
		Config: c,
		Redis:  rdb,
		Milvus: milvusClient,
	}
}
