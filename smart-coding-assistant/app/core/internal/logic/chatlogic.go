package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"time"

	memorypb "smart-coding-assistant/app/memory/pb"

	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

const memoryCollection = "chat_history"

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

// Chat 是 Agent 的主入口：
//   1. RAG 语义召回历史
//   2. Eino ReAct Agent 自主推理 + MCP 工具调用
//   3. 异步回存对话到向量记忆
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()
	msgLower := strings.ToLower(message)

	// ========== 闲聊快速路径 ==========
	if isSystemQuestion(msgLower) {
		response := l.chatResponse(message)
		go l.rememberMessage(in.GetUserId(), message, response)
		return &pb.ChatResponse{Response: response, IsFinished: true, Context: in.Context}, nil
	}

	// ========== RAG 语义召回 ==========
	recalledHistory := l.recallSimilarHistory(in.GetUserId(), message)
	enrichedCtx := mergeContext(in.GetContext(), recalledHistory)

	// ========== Eino Agent ==========
	agentInput := message
	if recalledHistory != "" {
		agentInput = fmt.Sprintf("【历史上下文】\n%s\n\n【当前问题】\n%s", recalledHistory, message)
	}

	response, err := l.svcCtx.Agent.Run(l.ctx, agentInput)
	if err != nil {
		response = fmt.Sprintf("抱歉，执行过程中出现错误: %v\n\n请尝试更具体地描述你的问题。", err)
	}

	// ========== 异步存储 ==========
	go l.rememberMessage(in.GetUserId(), message, response)

	// 清理 context 引用避免泄露
	_ = enrichedCtx

	return &pb.ChatResponse{
		Response:   response,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}

// ==============================
// 闲聊 / 兜底
// ==============================

func isSystemQuestion(msgLower string) bool {
	asciiPatterns := []string{"hi", "hello", "thanks", "thank", "help", "what can you do",
		"who are you", "what are you", "your name"}
	for _, kw := range asciiPatterns {
		if strings.Contains(msgLower, kw) { return true }
	}
	cnPatterns := []string{"你是什么", "什么模型", "你是谁", "你叫什么", "你的名字",
		"你能做什么", "有什么功能", "你好", "谢谢", "在吗", "模型"}
	for _, kw := range cnPatterns {
		if strings.Contains(msgLower, kw) { return true }
	}
	return false
}

func (l *ChatLogic) chatResponse(message string) string {
	msgLower := strings.ToLower(message)
	if strings.Contains(msgLower, "模型") || strings.Contains(msgLower, "model") {
		return "我是一个 AI Agent，基于 Eino ReAct 架构，具备自主规划、MCP 工具调用和语义记忆能力。\n\n" +
			"我会理解你的问题后自主调用合适的工具，进行多步推理，最终给出答案。"
	}
	if strings.Contains(msgLower, "你是谁") || strings.Contains(msgLower, "你的名字") {
		return "我是**智能编程学习助手** 🤖，一个基于 Eino ReAct + go-zero 的 AI Agent。"
	}
	if strings.Contains(msgLower, "hi") || strings.Contains(msgLower, "hello") || strings.Contains(msgLower, "你好") {
		return "你好！👋 我是智能编程学习助手，有什么编程问题需要帮助吗？"
	}
	if strings.Contains(msgLower, "thanks") || strings.Contains(msgLower, "谢谢") {
		return "不客气！有任何编程问题随时问我 😊"
	}
	return l.defaultResponse()
}

func (l *ChatLogic) defaultResponse() string {
	var sb strings.Builder
	sb.WriteString("👋 你好！我是 AI Agent 编程助手。\n\n")
	sb.WriteString("我可以自主调用工具完成：\n\n")
	sb.WriteString("- 分析代码错误\n")
	sb.WriteString("- 查询语法概念\n")
	sb.WriteString("- 生成解题方案\n")
	sb.WriteString("- 多步推理任务\n\n")
	if l.svcCtx.MCPClient != nil {
		allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
		if len(allTools) > 0 {
			sb.WriteString("---\n**可用 MCP 工具**:\n")
			for serverName, tools := range allTools {
				for _, t := range tools {
					sb.WriteString(fmt.Sprintf("- `%s` (%s): %s\n", t.Name, serverName, t.Description))
				}
			}
		}
	}
	return sb.String()
}

// ==============================
// RAG: Eino Embedder + MemoryRpc
// ==============================

func (l *ChatLogic) recallSimilarHistory(userId, message string) string {
	if l.svcCtx.Embedder == nil || l.svcCtx.MemoryRpc == nil {
		return ""
	}

	// Eino Embedder: [][]float64
	vecs, err := l.svcCtx.Embedder.EmbedStrings(l.ctx, []string{message})
	if err != nil || len(vecs) == 0 {
		if err != nil {
			log.Printf("[Memory] 向量化失败(召回): %v", err)
		}
		return ""
	}

	// [][]float64 → []float32
	queryVec := float64sToFloat32s(vecs[0])

	resp, err := l.svcCtx.MemoryRpc.SearchSimilar(l.ctx, &memorypb.SearchSimilarRequest{
		Collection:  memoryCollection,
		QueryVector: queryVec,
		TopK:        3,
	})
	if err != nil || !resp.Success || len(resp.Results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("【相关历史对话】\n")
	for i, r := range resp.Results {
		q := r.Metadata["q"]
		a := r.Metadata["a"]
		if q == "" { continue }
		fmt.Fprintf(&sb, "%d. 问: %s\n", i+1, q)
		if a != "" {
			fmt.Fprintf(&sb, "   答: %s\n", a)
		}
	}
	return sb.String()
}

func (l *ChatLogic) rememberMessage(userId, message, response string) {
	if l.svcCtx.Embedder == nil || l.svcCtx.MemoryRpc == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vecs, err := l.svcCtx.Embedder.EmbedStrings(ctx, []string{message})
	if err != nil || len(vecs) == 0 {
		if err != nil {
			log.Printf("[Memory] 向量化失败(存储): %v", err)
		}
		return
	}

	queryVec := float64sToFloat32s(vecs[0])
	id := messageID(userId, message)
	_, err = l.svcCtx.MemoryRpc.SaveVector(ctx, &memorypb.SaveVectorRequest{
		Collection: memoryCollection,
		Vectors: []*memorypb.VectorData{
			{
				Id:     id,
				Vector: queryVec,
				Metadata: map[string]string{
					"q":       message,
					"a":       response,
					"user_id": userId,
					"ts":      time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	})
	if err != nil {
		log.Printf("[Memory] 存储向量失败: %v", err)
	}
}

func float64sToFloat32s(in []float64) []float32 {
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}

func messageID(userId, message string) int64 {
	h := fnv.New64a()
	h.Write([]byte(userId + "|" + message))
	return int64(h.Sum64())
}

func mergeContext(original map[string]string, recalled string) map[string]string {
	if recalled == "" {
		return original
	}
	merged := make(map[string]string, len(original)+1)
	for k, v := range original {
		merged[k] = v
	}
	merged["history"] = recalled
	return merged
}

// ==============================
// MCP 工具结果格式化（保留）
// ==============================

func formatToolResult(result interface{}, title string) string {
	var sb strings.Builder
	sb.WriteString("## " + title + "\n\n")
	if result == nil {
		return sb.String() + "(无结果)"
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return sb.String() + fmt.Sprintf("%v", result)
	}
	type contentItem struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type toolResult struct {
		Content []contentItem `json:"content"`
	}
	var tr toolResult
	if err := json.Unmarshal(jsonBytes, &tr); err == nil {
		for _, c := range tr.Content {
			if c.Type == "text" {
				var parsed interface{}
				if json.Unmarshal([]byte(c.Text), &parsed) == nil {
					prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
					sb.WriteString(string(prettyJSON) + "\n")
				} else {
					sb.WriteString(c.Text + "\n")
				}
			}
		}
		return sb.String()
	}
	return sb.String() + fmt.Sprintf("%v", result)
}
