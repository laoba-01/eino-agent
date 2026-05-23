package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultEmbeddingAPIURL = "https://open.bigmodel.cn/api/paas/v4/embeddings"
	defaultEmbeddingModel  = "embedding-3"
)

var (
	apiKey string
	apiURL string
	model  string
	client *http.Client
)

func init() {
	apiURL = os.Getenv("GLM_EMBEDDING_API_URL")
	if apiURL == "" {
		apiURL = defaultEmbeddingAPIURL
	}

	model = os.Getenv("GLM_EMBEDDING_MODEL")
	if model == "" {
		model = defaultEmbeddingModel
	}

	apiKey = os.Getenv("GLM_API_KEY")

	client = &http.Client{Timeout: 30 * time.Second}
}

func loadEnvFile() {
	candidates := []string{".env", filepath.Join("..", ".env")}
	for _, p := range candidates {
		content, err := os.ReadFile(p)
		if err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					os.Setenv(key, val)
				}
			}
			return
		}
	}
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func Embed(text string) ([]float32, error) {
	loadEnvFile()
	key := os.Getenv("GLM_API_KEY")
	if key == "" {
		key = apiKey
	}
	if key == "" {
		return nil, fmt.Errorf("GLM_API_KEY not set")
	}

	reqBody := embeddingRequest{Model: model, Input: text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vec := make([]float32, len(embResp.Data[0].Embedding))
	for i, v := range embResp.Data[0].Embedding {
		vec[i] = float32(v)
	}
	return vec, nil
}

func EmbedBatch(texts []string) ([][]float32, error) {
	vecs := make([][]float32, 0, len(texts))
	for i, text := range texts {
		vec, err := Embed(text)
		if err != nil {
			return nil, fmt.Errorf("embed text %d: %w", i, err)
		}
		vecs = append(vecs, vec)
		if (i+1)%5 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return vecs, nil
}
