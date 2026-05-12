package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	coreProto "eino/core-service/proto"
	authProto "eino/auth-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	authServiceAddr = "localhost:50054"
	coreServiceAddr = "localhost:50051"
)

type ChatRequest struct {
	Message string            `json:"message"`
	Context map[string]string `json:"context"`
}

type ChatResponse struct {
	Response   string            `json:"response"`
	IsFinished bool              `json:"is_finished"`
	Context    map[string]string `json:"context"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	coreClient coreProto.CoreServiceClient
	authClient authProto.AuthServiceClient
)

func initCoreServiceClient() error {
	conn, err := grpc.Dial(coreServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接核心服务失败: %v", err)
	}
	coreClient = coreProto.NewCoreServiceClient(conn)
	log.Println("成功连接到核心学习服务:", coreServiceAddr)
	return nil
}

func initAuthServiceClient() error {
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接认证服务失败: %v", err)
	}
	authClient = authProto.NewAuthServiceClient(conn)
	log.Println("成功连接到认证服务:", authServiceAddr)
	return nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username and password are required"})
		return
	}

	if authClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := authClient.Register(ctx, &authProto.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("注册失败: %v", err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(RegisterResponse{
			Success: resp.Success,
			Message: resp.Message,
			UserID:  resp.UserId,
		})
		return
	}

	resp := RegisterResponse{
		Success: true,
		Message: "Registration successful (mock)",
		UserID:  "mock-user-id",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username and password are required"})
		return
	}

	if authClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := authClient.Login(ctx, &authProto.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("登录失败: %v", err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if resp.Success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		json.NewEncoder(w).Encode(LoginResponse{
			Success: resp.Success,
			Message: resp.Message,
			Token:   resp.Token,
			UserID:  resp.UserId,
		})
		return
	}

	resp := LoginResponse{
		Success: true,
		Message: "Login successful (mock)",
		Token:   "mock-jwt-token",
		UserID:  "mock-user-id",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "No token provided"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")

	if authClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := authClient.Logout(ctx, &authProto.LogoutRequest{Token: token})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("登出失败: %v", err)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": resp.Success})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "No token provided"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		if authClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			resp, err := authClient.ValidateToken(ctx, &authProto.ValidateTokenRequest{Token: token})
			if err != nil || !resp.Valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid token"})
				return
			}
			ctx = context.WithValue(r.Context(), "user_id", resp.UserId)
			next(w, r.WithContext(ctx))
			return
		}

		userID := "mock-user-id"
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next(w, r.WithContext(ctx))
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(string)

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("收到聊天请求，用户ID: %s, 消息: %s", userID, req.Message)

	grpcReq := &coreProto.ChatRequest{
		UserId:  userID,
		Message: req.Message,
		Context: req.Context,
	}

	grpcCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	grpcResp, err := coreClient.Chat(grpcCtx, grpcReq)
	if err != nil {
		log.Printf("调用核心服务出错: %v", err)
		http.Error(w, fmt.Sprintf("Service error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ChatResponse{
		Response:   grpcResp.Response,
		IsFinished: grpcResp.IsFinished,
		Context:    grpcResp.Context,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleLearningReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(string)

	grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := coreClient.GetLearningReport(grpcCtx, &coreProto.GetLearningReportRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("获取学习报告出错: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("获取学习报告失败: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleLearningPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(string)

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "last_seen"
	}

	grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := coreClient.GetKnowledgePoints(grpcCtx, &coreProto.GetKnowledgePointsRequest{
		UserId: userID,
		SortBy: sortBy,
	})
	if err != nil {
		log.Printf("获取知识点列表出错: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("获取知识点列表失败: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := initCoreServiceClient(); err != nil {
		log.Printf("警告: 无法连接到核心学习服务: %v", err)
	}

	if err := initAuthServiceClient(); err != nil {
		log.Printf("警告: 无法连接到认证服务: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/chat", authMiddleware(handleChat))
	mux.HandleFunc("/api/learning/report", authMiddleware(handleLearningReport))
	mux.HandleFunc("/api/learning/points", authMiddleware(handleLearningPoints))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "d:/agent/smart-coding-assistant/frontend/index.html")
			return
		}
		filePath := "d:/agent/smart-coding-assistant/frontend" + r.URL.Path
		http.ServeFile(w, r, filePath)
	})

	handler := corsMiddleware(mux)

	port := "8080"
	fmt.Printf("API Gateway listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
