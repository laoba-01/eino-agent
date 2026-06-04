package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// EmbeddingClient 调用 OpenAI 兼容的 Embedding API 将文本转为向量
type EmbeddingClient struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// NewEmbeddingClient 创建 Embedding 客户端
// endpoint: API 地址，如 https://api.deepseek.com/v1
// apiKey: API 密钥
// model: 模型名称，如 deepseek-chat
func NewEmbeddingClient(endpoint, apiKey, model string) *EmbeddingClient {
	return &EmbeddingClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed 将单条文本转换为向量
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("embedding API 返回空结果")
	}
	return vecs[0], nil
}

// EmbedBatch 批量将多条文本转换为向量
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Embedding API key 未配置")
	}

	results := make([][]float32, len(texts))

	for i, text := range texts {
		vec, err := c.embedWithRetry(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embedding 第 %d 条失败: %w", i, err)
		}
		results[i] = vec
	}

	return results, nil
}

// embedWithRetry 带重试的单条 embedding 请求
func (c *EmbeddingClient) embedWithRetry(ctx context.Context, text string) ([]float32, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			log.Printf("[Embedding] 重试 %d/3 (等待 %v): %v", attempt, backoff, lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		vec, err := c.doEmbed(ctx, text)
		if err == nil {
			return vec, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("embedding 请求失败(已重试3次): %w", lastErr)
}

// doEmbed 执行单次 embedding API 调用
func (c *EmbeddingClient) doEmbed(ctx context.Context, text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Model: c.model,
		Input: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.endpoint + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态 %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(embResp.Data) == 0 || len(embResp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("API 返回的 embedding 为空")
	}

	return embResp.Data[0].Embedding, nil
}
