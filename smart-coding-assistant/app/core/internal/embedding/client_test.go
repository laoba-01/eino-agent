package embedding

import (
	"context"
	"os"
	"testing"
	"time"
)

// 测试用环境变量，不写死任何厂商
func testConfig() (endpoint, apiKey, model string, ok bool) {
	endpoint = os.Getenv("EMBEDDING_ENDPOINT")
	apiKey = os.Getenv("EMBEDDING_API_KEY")
	model = os.Getenv("EMBEDDING_MODEL")

	if endpoint == "" {
		endpoint = "https://open.bigmodel.cn/api/paas/v4"
	}
	if model == "" {
		model = "embedding-2"
	}
	if apiKey == "" {
		return "", "", "", false
	}
	return endpoint, apiKey, model, true
}

func TestEmbeddingClient_Embed(t *testing.T) {
	endpoint, apiKey, model, ok := testConfig()
	if !ok {
		t.Skip("EMBEDDING_API_KEY 未设置，跳过集成测试")
	}

	client := NewEmbeddingClient(endpoint, apiKey, model)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	text := "Goroutine 是 Go 语言中的轻量级协程，用于并发编程"
	vec, err := client.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Embed 失败: %v", err)
	}

	if len(vec) == 0 {
		t.Fatal("返回的向量为空")
	}

	t.Logf("✅ 向量化成功，维度: %d (provider=%s, model=%s)", len(vec), endpoint, model)
	t.Logf("   文本: %s", text)
	t.Logf("   向量前 5 维: [%.4f, %.4f, %.4f, %.4f, %.4f]", vec[0], vec[1], vec[2], vec[3], vec[4])
}

func TestEmbeddingClient_EmbedBatch(t *testing.T) {
	endpoint, apiKey, model, ok := testConfig()
	if !ok {
		t.Skip("EMBEDDING_API_KEY 未设置，跳过集成测试")
	}

	client := NewEmbeddingClient(endpoint, apiKey, model)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	texts := []string{
		"Python 的装饰器是一种高阶函数",
		"如何在 Go 中实现并发安全的计数器",
		"二分查找的时间复杂度是 O(log n)",
	}

	vecs, err := client.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedBatch 失败: %v", err)
	}

	if len(vecs) != len(texts) {
		t.Fatalf("返回向量数量不匹配: 期望 %d, 实际 %d", len(texts), len(vecs))
	}

	for i, vec := range vecs {
		if len(vec) == 0 {
			t.Fatalf("第 %d 条向量为空", i)
		}
	}

	t.Logf("✅ 批量向量化成功，共 %d 条，每条维度: %d", len(vecs), len(vecs[0]))
	for i, text := range texts {
		t.Logf("   [%d] %s", i, text)
	}
}

func TestEmbeddingClient_EmptyKey(t *testing.T) {
	client := NewEmbeddingClient("https://open.bigmodel.cn/api/paas/v4", "", "embedding-2")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Embed(ctx, "test")
	if err == nil {
		t.Fatal("期望返回错误(空 API key)，但成功了")
	}
	t.Logf("✅ 空 API key 正确返回错误: %v", err)
}

func TestEmbeddingClient_Timeout(t *testing.T) {
	_, apiKey, _, ok := testConfig()
	if !ok {
		apiKey = "test-key" // 超时测试不需要真实 key
	}

	client := NewEmbeddingClient("https://192.0.2.1:9999", apiKey, "embedding-2")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := client.Embed(ctx, "test")
	if err == nil {
		t.Fatal("期望超时错误，但成功了")
	}
	t.Logf("✅ 超时正确触发: %v", err)
}
