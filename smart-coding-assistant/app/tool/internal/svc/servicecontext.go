package svc

import (
	"time"

	"smart-coding-assistant/app/tool/internal/config"
	"smart-coding-assistant/pkg/llm"
)

type ServiceContext struct {
	Config    config.Config
	LLMClient *llm.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	llmClient := llm.NewClient(llm.Config{
		Endpoint: c.LLM.Endpoint,
		APIKey:   c.LLM.APIKey,
		Model:    c.LLM.Model,
		Timeout:  60 * time.Second,
	})
	return &ServiceContext{
		Config:    c,
		LLMClient: llmClient,
	}
}
