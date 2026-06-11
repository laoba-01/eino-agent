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

	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/planner"
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
//   2. Planner（LLM）生成执行计划
//   3. Executor 逐步执行（MCP 工具 + LLM 推理）
//   4. 异步回存对话到向量记忆
func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()
	msgLower := strings.ToLower(message)

	// ========== 闲聊快速路径 ==========
	if isSystemQuestion(msgLower) {
		response := l.chatResponse(message)
		go l.rememberMessage(in.GetUserId(), message, response)
		return &pb.ChatResponse{
			Response:   response,
			IsFinished: true,
			Context:    in.Context,
		}, nil
	}

	// ========== RAG 语义召回 ==========
	recalledHistory := l.recallSimilarHistory(in.GetUserId(), message)
	enrichedCtx := mergeContext(in.GetContext(), recalledHistory)

	// ========== Agent: 收集可用工具 ==========
	availableTools := l.collectTools()

	// ========== Agent: Planner 生成执行计划 ==========
	plan, err := l.svcCtx.Planner.GeneratePlan(l.ctx, message, enrichedCtx, availableTools)
	if err != nil {
		// 降级：计划失败时用 LLM 直接回答
		response := fmt.Sprintf("抱歉，无法制定执行计划: %v\n\n请尝试更具体地描述你的问题。", err)
		go l.rememberMessage(in.GetUserId(), message, response)
		return &pb.ChatResponse{
			Response:   response,
			IsFinished: true,
		}, nil
	}

	// ========== Agent: Executor 逐步执行 ==========
	resp := &pb.ChatResponse{
		IsFinished: false,
		Plan: &pb.PlanInfo{
			PlanId:     plan.ID,
			Goal:       plan.Goal,
			TotalSteps: int32(len(plan.Steps)),
		},
		Context: in.Context,
	}

	reporter := &chatReporter{resp: resp, totalSteps: len(plan.Steps)}
	if err := l.svcCtx.Executor.Execute(l.ctx, plan, reporter); err != nil {
		resp.Response = fmt.Sprintf("执行终止: %v", err)
		resp.IsFinished = true
		go l.rememberMessage(in.GetUserId(), message, resp.Response)
		return resp, nil
	}

	// ========== 构建最终响应 + 异步存储 ==========
	response := l.buildFinalResponse(plan)
	resp.Response = response
	resp.IsFinished = true

	go l.rememberMessage(in.GetUserId(), message, response)

	return resp, nil
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
		return "我是一个 AI Agent，使用 Planner + Executor 架构。\n\n" +
			"我会理解你的问题后自主制定执行计划，按步骤调用工具并推理，最终给出答案。\n\n" +
			"支持：代码错误分析 / 语法查询 / 编程解题 / 多步推理。"
	}
	if strings.Contains(msgLower, "你是谁") || strings.Contains(msgLower, "你的名字") {
		return "我是**智能编程学习助手** 🤖，一个基于 Plan-and-Execute 架构的 AI Agent。\n\n" +
			"具备自主规划、MCP 工具调用和语义记忆能力。"
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
	sb.WriteString("我可以自主规划并执行复杂任务：\n\n")
	sb.WriteString("1. **分析代码错误** — 发送包含错误信息和代码的消息\n")
	sb.WriteString("2. **查询语法概念** — 询问编程语言语法\n")
	sb.WriteString("3. **生成解题方案** — 描述编程问题，获取解题思路和代码\n")
	sb.WriteString("4. **多步推理任务** — 我会自行分解目标并逐步完成\n\n")

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
// Agent: 工具收集
// ==============================

func (l *ChatLogic) collectTools() []string {
	var tools []string
	if l.svcCtx.MCPClient == nil {
		return tools
	}
	allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
	for _, serverTools := range allTools {
		for _, t := range serverTools {
			tools = append(tools, t.Name)
		}
	}
	return tools
}

// ==============================
// Agent: 最终响应构建
// ==============================

func (l *ChatLogic) buildFinalResponse(plan *planner.Plan) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s\n\n", plan.Goal)
	for _, step := range plan.Steps {
		icon := "✅"
		if step.Status == string(planner.StepStatusFailed) {
			icon = "❌"
		}
		fmt.Fprintf(&sb, "%s **步骤%d**: %s\n", icon, step.Index, step.Description)
		if step.Result != "" {
			fmt.Fprintf(&sb, "   %s\n\n", step.Result)
		}
		if step.Error != "" {
			fmt.Fprintf(&sb, "   错误: %s\n\n", step.Error)
		}
	}
	return sb.String()
}

// ==============================
// Agent: 步骤进度回调
// ==============================

type chatReporter struct {
	resp       *pb.ChatResponse
	totalSteps int
}

func (r *chatReporter) OnStepStart(step planner.Step) {
	r.resp.Response = fmt.Sprintf("正在执行步骤 %d/%d: %s", step.Index, r.totalSteps, step.Description)
}

func (r *chatReporter) OnStepDone(step planner.Step) {
	r.resp.Steps = append(r.resp.Steps, &pb.StepResult{
		Index:       int32(step.Index),
		Description: step.Description,
		Status:      step.Status,
		Result:      step.Result,
		Error:       step.Error,
	})
	r.resp.Plan.CompletedSteps = int32(len(r.resp.Steps))
}

func (r *chatReporter) OnAllDone(plan planner.Plan) {
	r.resp.Plan.CompletedSteps = int32(r.totalSteps)
}

var _ executor.StepReporter = (*chatReporter)(nil)

// ==============================
// RAG: 语义记忆
// ==============================

// recallSimilarHistory 搜索与当前消息语义相似的历史对话
func (l *ChatLogic) recallSimilarHistory(userId, message string) string {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return ""
	}

	queryVec, err := l.svcCtx.Embedding.Embed(l.ctx, message)
	if err != nil {
		log.Printf("[Memory] 向量化失败(召回): %v", err)
		return ""
	}

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
		if q == "" {
			continue
		}
		fmt.Fprintf(&sb, "%d. 问: %s\n", i+1, q)
		if a != "" {
			fmt.Fprintf(&sb, "   答: %s\n", a)
		}
	}
	return sb.String()
}

// rememberMessage 异步将对话存入向量记忆
func (l *ChatLogic) rememberMessage(userId, message, response string) {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vec, err := l.svcCtx.Embedding.Embed(ctx, message)
	if err != nil {
		log.Printf("[Memory] 向量化失败(存储): %v", err)
		return
	}

	id := messageID(userId, message)
	_, err = l.svcCtx.MemoryRpc.SaveVector(ctx, &memorypb.SaveVectorRequest{
		Collection: memoryCollection,
		Vectors: []*memorypb.VectorData{
			{
				Id:     id,
				Vector: vec,
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
// MCP 工具结果格式化（保留供 Executor 输出）
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
