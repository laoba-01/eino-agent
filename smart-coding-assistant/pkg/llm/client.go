package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 大模型 API HTTP 客户端
type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

// Config LLM 客户端配置
type Config struct {
	Endpoint string
	APIKey   string
	Model    string
	Timeout  time.Duration
}

// ChatMessage LLM 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Client{
		endpoint: cfg.Endpoint,
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := chatRequest{Model: c.model, Messages: messages}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm api status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	var cr chatResponse
	if err := json.Unmarshal(respBytes, &cr); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if cr.Error != nil {
		return "", fmt.Errorf("llm api error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("llm api returned no choices")
	}
	return cr.Choices[0].Message.Content, nil
}
