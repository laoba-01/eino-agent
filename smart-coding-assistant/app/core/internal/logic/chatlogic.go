package logic

import (
	"context"
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
//   2. Eino ReAct Agent 自主推理 + MCP 工具调用（统一处理闲聊和编程问题）
//   3. 异步回存对话到向量记忆
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()

	// 按请求语言选择 system prompt
	lang := in.GetLanguage()
	if lang == "" {
		lang = "zh"
	}
	sysPrompt := l.svcCtx.SystemPrompt[lang]
	if sysPrompt == "" {
		sysPrompt = l.svcCtx.SystemPrompt["zh"] // fallback
	}

	// ========== RAG 语义召回 ==========
	recalledHistory := l.recallSimilarHistory(in.GetUserId(), message, lang)
	// ========== Eino Agent ==========
	agentInput := message
	if recalledHistory != "" {
		if lang == "en" {
			agentInput = fmt.Sprintf("[Historical Context]\n%s\n\n[Current Question]\n%s", recalledHistory, message)
		} else {
			agentInput = fmt.Sprintf("【历史上下文】\n%s\n\n【当前问题】\n%s", recalledHistory, message)
		}
	}

	response, err := l.svcCtx.Agent.Run(l.ctx, sysPrompt, agentInput)
	if err != nil {
		if lang == "en" {
			response = fmt.Sprintf("Sorry, an error occurred: %v\n\nPlease try again or describe your problem more specifically.", err)
		} else {
			response = fmt.Sprintf("抱歉，执行过程中出现错误: %v\n\n请稍后重试或更具体地描述你的问题。", err)
		}
	}

	// ========== 异步存储 ==========
	// 仅在 Agent 成功响应时存储，避免错误回复污染 RAG 记忆
	if err == nil {
		go l.rememberMessage(in.GetUserId(), message, response)
	}

	return &pb.ChatResponse{
		Response:   response,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}

// ==============================
// RAG: Eino Embedder + MemoryRpc
// ==============================

func (l *ChatLogic) recallSimilarHistory(userId, message, lang string) string {
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
	if lang == "en" {
		sb.WriteString("[Relevant Chat History]\n")
	} else {
		sb.WriteString("【相关历史对话】\n")
	}
	for i, r := range resp.Results {
		q := r.Metadata["q"]
		a := r.Metadata["a"]
		if q == "" { continue }
		if lang == "en" {
			fmt.Fprintf(&sb, "%d. Q: %s\n", i+1, q)
			if a != "" {
				fmt.Fprintf(&sb, "   A: %s\n", a)
			}
		} else {
			fmt.Fprintf(&sb, "%d. 问: %s\n", i+1, q)
			if a != "" {
				fmt.Fprintf(&sb, "   答: %s\n", a)
			}
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

