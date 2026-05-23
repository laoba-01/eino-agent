package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"rag-eval/embedding"

	proto "eino/memory-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Document struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

var memoryAddr = getEnv("MEMORY_SERVICE_ADDR", "localhost:50053")

func main() {
	dataDir := resolveDataDir()
	docsPath := filepath.Join(dataDir, "data", "documents.json")

	docs, err := loadDocuments(docsPath)
	if err != nil {
		log.Fatalf("加载文档失败: %v", err)
	}
	log.Printf("已加载 %d 个知识文档", len(docs))

	conn, err := grpc.Dial(memoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接 memory-service 失败: %v", err)
	}
	defer conn.Close()
	client := proto.NewMemoryServiceClient(conn)

	const collection = "rag_knowledge"

	for i, doc := range docs {
		log.Printf("[%d/%d] 正在嵌入文档 #%d: %s", i+1, len(docs), doc.ID, doc.Title)

		vec, err := embedding.Embed(doc.Content)
		if err != nil {
			log.Fatalf("嵌入文档 #%d 失败: %v", doc.ID, err)
		}
		log.Printf("  向量维度: %d", len(vec))

		metadata := map[string]string{
			"title":   doc.Title,
			"content": doc.Content,
		}

		resp, err := client.SaveVector(context.Background(), &proto.SaveVectorRequest{
			Collection: collection,
			Vectors: []*proto.VectorData{
				{
					Id:       int64(doc.ID),
					Vector:   vec,
					Metadata: metadata,
				},
			},
		})
		if err != nil {
			log.Fatalf("存储向量 #%d 失败: %v", doc.ID, err)
		}
		if !resp.Success {
			log.Fatalf("存储向量 #%d 返回失败: %s", doc.ID, resp.Error)
		}

		log.Printf("  已入库 (IDs: %v)", resp.InsertedIds)

		if (i+1)%5 == 0 {
			log.Printf("已处理 %d/%d，暂停 500ms 避免 API 限频...", i+1, len(docs))
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("入库完成！共 %d 个文档存入 collection %q", len(docs), collection)
}

func loadDocuments(path string) ([]Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var docs []Document
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return docs, nil
}

func resolveDataDir() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	dir := filepath.Dir(f)
	for {
		if _, err := os.Stat(filepath.Join(dir, "data")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
