package svc

import (
	"github.com/redis/go-redis/v9"
	"smart-coding-assistant/app/auth/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BizRedis.Addr,
		Password: "",
		DB:       0,
	})

	return &ServiceContext{
		Config: c,
		Redis:  rdb,
	}
}
