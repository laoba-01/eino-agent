package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"eino/core-service/proto"
	memoryProto "eino/memory-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultGLMAPIURL   = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	defaultModel       = "glm-5.1"
	memoryServiceAddr  = "localhost:50053"
)

func loadEnvFile() {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
	}

	var content []byte
	var envPath string
	var err error

	for _, p := range candidates {
		content, err = os.ReadFile(p)
		if err == nil {
			envPath = p
			break
		}
	}

	if content == nil {
		log.Printf("未找到.env配置文件，将使用系统环境变量")
		return
	}

	log.Printf("已加载.env配置文件: %s", envPath)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
			log.Printf("加载环境变量: %s", key)
		}
	}
}

type GLMRequest struct {
	Model    string       `json:"model"`
	Messages []GLMMessage `json:"messages"`
}

type GLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GLMResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Choices []GLMChoice `json:"choices"`
	Usage   GLMUsage    `json:"usage"`
}

type GLMChoice struct {
	Index        int        `json:"index"`
	Message      GLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type GLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type knowledgePointDelta struct {
	Topic        string  `json:"topic"`
	MasteryDelta float64 `json:"mastery_delta"`
}

type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
	glmAPIKey    string
	glmAPIURL    string
	modelName    string
	httpClient   *http.Client
	memoryClient memoryProto.MemoryServiceClient
}

func NewCoreServiceServer() *CoreServiceServer {
	loadEnvFile()

	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		log.Println("警告: GLM_API_KEY 环境变量未设置，将使用模拟响应")
	} else {
		log.Println("GLM API Key 已成功加载")
	}

	apiURL := os.Getenv("GLM_API_URL")
	if apiURL == "" {
		apiURL = defaultGLMAPIURL
	}

	model := os.Getenv("GLM_MODEL")
	if model == "" {
		model = defaultModel
	}

	log.Printf("使用模型: %s", model)

	return &CoreServiceServer{
		glmAPIKey: apiKey,
		glmAPIURL: apiURL,
		modelName: model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *CoreServiceServer) initMemoryClient() error {
	conn, err := grpc.Dial(memoryServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接记忆服务失败: %v", err)
	}
	s.memoryClient = memoryProto.NewMemoryServiceClient(conn)
	log.Println("成功连接到记忆管理服务:", memoryServiceAddr)
	return nil
}

const systemPrompt = `你是一个专业的智能编程学习助手，专门帮助用户解决编程学习中的问题。你擅长：1. 分析代码错误并提供修复建议 2. 解释编程语言的语法知识并提供示例 3. 为算法和编程题目提供解题思路和代码实现 4. 回答各种编程相关的学习问题。

请用中文友好地回答用户问题，同时识别本次对话涉及的编程知识点。

你必须严格按照以下格式回复：
<response>
你的正常回答内容
</response>
<knowledge_points>
[{"topic": "知识点名称", "mastery_delta": 0.1}]
</knowledge_points>

规则：
1. <response>中是对用户的正常回答
2. <knowledge_points>中是识别的知识点列表（JSON数组）
3. mastery_delta 表示本次交互对该知识点的掌握度提升值，范围0.05-0.3
4. 知识点应具体且有意义（如"for循环"、"递归"、"哈希表"而非"编程基础"）
5. 每次识别1-3个知识点
6. 如果对话未涉及具体编程知识点，返回空列表 []`

func (s *CoreServiceServer) callGLMAPI(userMessage string) (string, error) {
	if s.glmAPIKey == "" {
		log.Println("GLM API Key未设置，返回模拟响应")
		return "<response>你好！我是智能编程学习助手，很高兴为你服务！请先设置GLM_API_KEY环境变量以启用真实的大模型功能。</response>\n<knowledge_points>[]</knowledge_points>", nil
	}

	messages := []GLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	reqBody := GLMRequest{
		Model:    s.modelName,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", s.glmAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.glmAPIKey)

	log.Printf("正在调用GLM API，模型: %s...", s.modelName)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("GLM API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API返回错误: %s", string(body))
	}

	var glmResp GLMResponse
	if err := json.Unmarshal(body, &glmResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(glmResp.Choices) == 0 {
		return "", fmt.Errorf("API返回空的choices")
	}

	responseText := glmResp.Choices[0].Message.Content
	log.Printf("GLM API调用成功，消耗token: 总计 %d", glmResp.Usage.TotalTokens)

	return responseText, nil
}

var (
	responseRe        = regexp.MustCompile(`(?s)<response>(.*?)</response>`)
	knowledgePointsRe = regexp.MustCompile(`(?s)<knowledge_points>(.*?)</knowledge_points>`)
)

func parseLLMResponse(raw string) (chatText string, points []knowledgePointDelta) {
	respMatch := responseRe.FindStringSubmatch(raw)
	if len(respMatch) >= 2 {
		chatText = strings.TrimSpace(respMatch[1])
	} else {
		kpMatch := knowledgePointsRe.FindStringSubmatch(raw)
		if len(kpMatch) >= 2 {
			chatText = strings.TrimSpace(strings.Replace(raw, kpMatch[0], "", 1))
		} else {
			chatText = strings.TrimSpace(raw)
		}
	}

	kpMatch := knowledgePointsRe.FindStringSubmatch(raw)
	if len(kpMatch) >= 2 {
		kpJSON := strings.TrimSpace(kpMatch[1])
		if kpJSON != "" && kpJSON != "[]" {
			var deltas []knowledgePointDelta
			if err := json.Unmarshal([]byte(kpJSON), &deltas); err == nil {
				points = deltas
			} else {
				log.Printf("Warning: failed to parse knowledge points JSON: %v", err)
			}
		}
	}

	return chatText, points
}

func (s *CoreServiceServer) saveKnowledgePoints(ctx context.Context, userID string, points []knowledgePointDelta) {
	if s.memoryClient == nil || len(points) == 0 {
		return
	}

	for _, kp := range points {
		delta := kp.MasteryDelta
		if delta < 0.05 {
			delta = 0.05
		} else if delta > 0.3 {
			delta = 0.3
		}

		_, err := s.memoryClient.SaveKnowledgePoint(ctx, &memoryProto.SaveKnowledgePointRequest{
			UserId: userID,
			Point: &memoryProto.KnowledgePoint{
				Topic:   kp.Topic,
				Mastery: delta,
			},
			Merge: true,
		})
		if err != nil {
			log.Printf("Warning: failed to save knowledge point %q: %v", kp.Topic, err)
		} else {
			log.Printf("Saved knowledge point: %s (delta=%.2f)", kp.Topic, delta)
		}
	}
}

func (s *CoreServiceServer) Chat(ctx context.Context, req *proto.ChatRequest) (*proto.ChatResponse, error) {
	log.Printf("收到聊天请求，用户ID: %s, 消息: %s", req.UserId, req.Message)

	rawResponse, err := s.callGLMAPI(req.Message)
	if err != nil {
		log.Printf("调用GLM API出错: %v", err)
		return &proto.ChatResponse{
			Response:   fmt.Sprintf("抱歉，处理您的请求时出错了：%v", err),
			IsFinished: true,
			Context:    req.Context,
		}, nil
	}

	chatText, knowledgePoints := parseLLMResponse(rawResponse)

	s.saveKnowledgePoints(ctx, req.UserId, knowledgePoints)

	return &proto.ChatResponse{
		Response:   chatText,
		IsFinished: true,
		Context:    req.Context,
	}, nil
}

func (s *CoreServiceServer) GetHistory(ctx context.Context, req *proto.GetHistoryRequest) (*proto.GetHistoryResponse, error) {
	log.Printf("获取历史记录，用户ID: %s", req.UserId)
	return &proto.GetHistoryResponse{
		Messages: []*proto.ChatMessage{},
	}, nil
}

func (s *CoreServiceServer) GetLearningReport(ctx context.Context, req *proto.GetLearningReportRequest) (*proto.GetLearningReportResponse, error) {
	log.Printf("获取学习报告，用户ID: %s", req.UserId)

	if s.memoryClient == nil {
		return &proto.GetLearningReportResponse{}, nil
	}

	statsResp, err := s.memoryClient.GetLearningStats(ctx, &memoryProto.GetLearningStatsRequest{
		UserId: req.UserId,
	})
	if err != nil {
		log.Printf("获取学习统计失败: %v", err)
		return &proto.GetLearningReportResponse{}, nil
	}

	pointsResp, err := s.memoryClient.GetKnowledgePoints(ctx, &memoryProto.GetKnowledgePointsRequest{
		UserId:  req.UserId,
		SortBy:  "last_seen",
	})
	if err != nil {
		log.Printf("获取知识点列表失败: %v", err)
		return &proto.GetLearningReportResponse{
			TotalTopics:      statsResp.TotalTopics,
			AverageMastery:   statsResp.AverageMastery,
			TotalInteractions: statsResp.TotalInteractions,
			MasteredCount:    statsResp.MasteredCount,
			LearningCount:    statsResp.LearningCount,
			WeakCount:        statsResp.WeakCount,
		}, nil
	}

	var recentPoints []*proto.KnowledgePointDetail
	for i, p := range pointsResp.Points {
		if i >= 10 {
			break
		}
		recentPoints = append(recentPoints, &proto.KnowledgePointDetail{
			Topic:       p.Topic,
			Mastery:     p.Mastery,
			Interactions: p.Interactions,
			LastSeen:    p.LastSeen,
		})
	}

	var weakPoints []*proto.KnowledgePointDetail
	for _, p := range pointsResp.Points {
		if p.Mastery < 0.3 {
			weakPoints = append(weakPoints, &proto.KnowledgePointDetail{
				Topic:       p.Topic,
				Mastery:     p.Mastery,
				Interactions: p.Interactions,
				LastSeen:    p.LastSeen,
			})
		}
	}
	if len(weakPoints) > 5 {
		weakPoints = weakPoints[:5]
	}

	return &proto.GetLearningReportResponse{
		TotalTopics:      statsResp.TotalTopics,
		AverageMastery:   statsResp.AverageMastery,
		TotalInteractions: statsResp.TotalInteractions,
		MasteredCount:    statsResp.MasteredCount,
		LearningCount:    statsResp.LearningCount,
		WeakCount:        statsResp.WeakCount,
		RecentPoints:     recentPoints,
		WeakPoints:       weakPoints,
	}, nil
}

func (s *CoreServiceServer) GetKnowledgePoints(ctx context.Context, req *proto.GetKnowledgePointsRequest) (*proto.GetKnowledgePointsResponse, error) {
	log.Printf("获取知识点列表，用户ID: %s", req.UserId)

	if s.memoryClient == nil {
		return &proto.GetKnowledgePointsResponse{Success: false, Error: "memory service not available"}, nil
	}

	pointsResp, err := s.memoryClient.GetKnowledgePoints(ctx, &memoryProto.GetKnowledgePointsRequest{
		UserId: req.UserId,
		Limit:  req.Limit,
		SortBy: req.SortBy,
	})
	if err != nil {
		return &proto.GetKnowledgePointsResponse{Success: false, Error: err.Error()}, nil
	}

	var points []*proto.KnowledgePointDetail
	for _, p := range pointsResp.Points {
		points = append(points, &proto.KnowledgePointDetail{
			Topic:       p.Topic,
			Mastery:     p.Mastery,
			Interactions: p.Interactions,
			LastSeen:    p.LastSeen,
		})
	}

	return &proto.GetKnowledgePointsResponse{Success: true, Points: points}, nil
}

func main() {
	server := NewCoreServiceServer()

	if err := server.initMemoryClient(); err != nil {
		log.Printf("警告: 无法连接到记忆管理服务: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterCoreServiceServer(s, server)

	log.Println("Core Service (with GLM API + Learning Tracking) listening on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
