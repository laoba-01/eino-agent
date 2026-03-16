package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// 定义请求和响应结构
type ChatRequest struct {
	UserID  string            `json:"user_id"`
	Message string            `json:"message"`
	Context map[string]string `json:"context"`
}

type ChatResponse struct {
	Response   string            `json:"response"`
	IsFinished bool              `json:"is_finished"`
	Context    map[string]string `json:"context"`
}

// 处理聊天请求
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 这里将实现与核心对话服务的gRPC调用
	// 暂时返回模拟响应
	resp := ChatResponse{
		Response:   "Hello from API Gateway!",
		IsFinished: true,
		Context:    req.Context,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// 健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
}

func main() {
	// 注册路由
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/health", handleHealth)

	// 启动HTTP服务器
	port := "8080"
	fmt.Printf("API Gateway listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}