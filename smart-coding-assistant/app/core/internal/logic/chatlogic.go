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

type Intent string

const (
	IntentAnalyzeCode     Intent = "analyze_code"
	IntentQuerySyntax     Intent = "query_syntax"
	IntentGenerateProblem Intent = "generate_problem"
	IntentChat            Intent = "chat"
	IntentUnknown         Intent = "unknown"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	message := in.GetMessage()
	msgLower := strings.ToLower(message)

	// 优先处理系统/闲聊问题（在 detectIntent 之前兜底）
	if isSystemQuestion(msgLower) {
		response := l.chatResponse(message)
		// 异步保存闲聊到记忆
		go l.rememberMessage(in.GetUserId(), message, response)
		return &pb.ChatResponse{
			Response:   response,
			IsFinished: true,
			Context:    in.Context,
		}, nil
	}

	// 语义召回：搜索相似历史对话
	recalledContext := l.recallSimilarHistory(in.GetUserId(), message)

	// 合并召回的历史到上下文
	enrichedCtx := mergeContext(in.GetContext(), recalledContext)

	intent := detectIntent(msgLower)

	response := l.handleIntent(intent, message, msgLower, enrichedCtx)

	// 异步存入向量记忆
	go l.rememberMessage(in.GetUserId(), message, response)

	return &pb.ChatResponse{
		Response:   response,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}

// isSystemQuestion 在 detectIntent 之前做一层兜底，避免因编译/缓存问题导致系统问题漏入工具调用
// 注意：使用 ASCII 和常见中文字符匹配，但某些环境下中文字符可能在传输中被损坏
// 因此匹配逻辑优先基于英文/拼音，中文作为辅助
func isSystemQuestion(msgLower string) bool {
	// ASCII 模式（最可靠）
	asciiPatterns := []string{"hi", "hello", "thanks", "thank", "help", "what can you do",
		"who are you", "what are you", "your name"}
	for _, kw := range asciiPatterns {
		if strings.Contains(msgLower, kw) {
			return true
		}
	}

	// 中文模式 — 逐个字节匹配，即使终端乱码也能工作
	// "你是什么" "什么模型" "你是谁" "你叫什么" "你的名字"
	// "你能做什么" "有什么功能" "你好" "谢谢" "在吗"
	cnPatterns := []string{
		"你是什么", "什么模型", "你是谁", "你叫什么", "你的名字",
		"你能做什么", "有什么功能", "你好", "谢谢", "在吗",
		"模型", // 单独匹配"模型"更宽泛
	}
	for _, kw := range cnPatterns {
		if strings.Contains(msgLower, kw) {
			return true
		}
	}
	return false
}

// detectIntent 根据用户消息关键词判断意图
func detectIntent(msgLower string) Intent {
	// 0. 系统/元问题：关于系统本身的问题，优先级最高
	chatKeywords := []string{"你是什么", "什么模型", "你是谁", "你叫什么", "你的名字",
		"你能做什么", "有什么功能", "你是什么模型", "hi", "hello", "你好", "thanks", "谢谢"}
	for _, kw := range chatKeywords {
		if strings.Contains(msgLower, kw) {
			return IntentChat
		}
	}

	// 1. 代码错误分析：包含错误相关关键词
	errorKeywords := []string{"error", "bug", "报错", "错误", "fix", "修复", "exception", "panic", "crash",
		"not working", "不工作", "failed", "失败", "undefined", "null pointer", "traceback"}
	for _, kw := range errorKeywords {
		if strings.Contains(msgLower, kw) {
			return IntentAnalyzeCode
		}
	}

	// 2. 语法查询：必须同时包含「疑问词 + 编程概念」或「明确的编程概念查询」
	//    单独 "是什么" "怎么用" 不触发，防止误匹配
	syntaxConcepts := []string{"decorator", "装饰器", "async", "await", "closure", "闭包",
		"goroutine", "channel", "generator", "生成器", "interface", "接口",
		"指针", "pointer", "继承", "inheritance", "多态", "polymorphism", "递归", "recursion",
		"泛型", "generic", "协程", "goroutine", "promise", "迭代器", "iterator",
		"list comprehension", "列表推导", "lambda", "正则", "regex", "序列化", "反序列化"}
	for _, c := range syntaxConcepts {
		if strings.Contains(msgLower, c) {
			return IntentQuerySyntax
		}
	}
	// 包含 what is / explain + programming context
	syntaxAskWords := []string{"what is", "how to use", "explain", "如何使用", "语法", "syntax", "用法", "是什么意思"}
	programmingContext := []string{"python", "go ", "golang", "javascript", "js ", "typescript", "java", "rust",
		"c++", "cpp", "函数", "function", "变量", "variable", "类型", "class", "类", "方法", "method",
		"模块", "module", "包", "package", "循环", "loop", "条件", "condition", "数组", "array"}
	hasAskWord := false
	for _, kw := range syntaxAskWords {
		if strings.Contains(msgLower, kw) {
			hasAskWord = true
			break
		}
	}
	hasProgCtx := false
	for _, kw := range programmingContext {
		if strings.Contains(msgLower, kw) {
			hasProgCtx = true
			break
		}
	}
	if hasAskWord && hasProgCtx {
		return IntentQuerySyntax
	}

	// 3. 解题：要求生成代码/解题
	solveKeywords := []string{"solve", "解题", "写一个", "实现", "implement", "solution", "解决方案",
		"算法", "algorithm", "write code", "编写", "generate", "生成", "two sum", "排序",
		"搜索", "反转", "链表", "二叉树", "动态规划", "dp ", "dp)"}
	for _, kw := range solveKeywords {
		if strings.Contains(msgLower, kw) {
			return IntentGenerateProblem
		}
	}

	return IntentUnknown
}

// handleIntent 根据意图调用对应的 MCP 工具
func (l *ChatLogic) handleIntent(intent Intent, message, msgLower string, ctx map[string]string) string {
	switch intent {
	case IntentChat:
		return l.chatResponse(message)

	case IntentAnalyzeCode:
		return l.callAnalyzeCodeError(message, ctx)

	case IntentQuerySyntax:
		return l.callQuerySyntax(message, ctx)

	case IntentGenerateProblem:
		return l.callGenerateProblemSolution(message, ctx)

	default:
		return l.defaultResponse()
	}
}

// callAnalyzeCodeError 调用代码错误分析工具
func (l *ChatLogic) callAnalyzeCodeError(message string, ctx map[string]string) string {
	if l.svcCtx.MCPClient == nil {
		return "MCP 客户端未初始化，无法分析代码错误。"
	}

	// 从消息中提取代码、错误信息和语言
	code := extractCodeBlock(message)
	errMsg := extractErrorMsg(message)
	lang := detectLanguage(message)
	if l, ok := ctx["language"]; ok && l != "" {
		lang = l
	}

	result, err := l.svcCtx.MCPClient.CallTool(l.ctx, "tool-service", "analyze_code_error", map[string]interface{}{
		"code":          code,
		"error_message": errMsg,
		"language":      lang,
	})
	if err != nil {
		return "调用代码分析工具失败: " + err.Error()
	}

	return formatToolResult(result, "代码错误分析")
}

// callQuerySyntax 调用语法查询工具
func (l *ChatLogic) callQuerySyntax(message string, ctx map[string]string) string {
	if l.svcCtx.MCPClient == nil {
		return "MCP 客户端未初始化，无法查询语法。"
	}

	lang, query := extractLangAndQuery(message)
	if l, ok := ctx["language"]; ok && l != "" {
		lang = l
	}

	result, err := l.svcCtx.MCPClient.CallTool(l.ctx, "tool-service", "query_syntax", map[string]interface{}{
		"language": lang,
		"query":    query,
		"context":  message,
	})
	if err != nil {
		return "语法查询失败: " + err.Error()
	}

	return formatToolResult(result, "语法查询")
}

// callGenerateProblemSolution 调用解题方案生成工具
func (l *ChatLogic) callGenerateProblemSolution(message string, ctx map[string]string) string {
	if l.svcCtx.MCPClient == nil {
		return "MCP 客户端未初始化，无法生成解题方案。"
	}

	lang := detectLanguage(message)
	if l, ok := ctx["language"]; ok && l != "" {
		lang = l
	}
	difficulty := "medium"
	if d, ok := ctx["difficulty"]; ok && d != "" {
		difficulty = d
	}

	result, err := l.svcCtx.MCPClient.CallTool(l.ctx, "tool-service", "generate_problem_solution", map[string]interface{}{
		"problem":    message,
		"difficulty": difficulty,
		"language":   lang,
	})
	if err != nil {
		return "生成解题方案失败: " + err.Error()
	}

	return formatToolResult(result, "解题方案")
}

// chatResponse 处理闲聊/系统性问题
func (l *ChatLogic) chatResponse(message string) string {
	msgLower := strings.ToLower(message)

	if strings.Contains(msgLower, "模型") || strings.Contains(msgLower, "model") {
		return "我目前是基于规则匹配的编程学习助手，尚未接入大语言模型。\n\n" +
			"我能通过关键词识别你的意图，并调用对应的工具来帮你：\n" +
			"- 分析代码错误\n" +
			"- 查询编程语法\n" +
			"- 生成解题方案\n\n" +
			"未来接入 LLM 后会有更强的对话能力！"
	}
	if strings.Contains(msgLower, "你是谁") || strings.Contains(msgLower, "你叫什么") || strings.Contains(msgLower, "你的名字") {
		return "我是**智能编程学习助手** 🤖，一个帮助你学习编程的 AI Agent。\n\n" +
			"我可以分析代码错误、解释语法概念、生成解题方案。试试问我编程相关的问题吧！"
	}
	if strings.Contains(msgLower, "你能做什么") || strings.Contains(msgLower, "有什么功能") || strings.Contains(msgLower, "help") {
		return l.defaultResponse()
	}
	if strings.Contains(msgLower, "hi") || strings.Contains(msgLower, "hello") || strings.Contains(msgLower, "你好") {
		return "你好！👋 我是智能编程学习助手，有什么编程问题需要帮助吗？\n\n" +
			"- 遇到代码报错？把错误信息发给我\n" +
			"- 想了解某个语法？直接问我\n" +
			"- 需要解题思路？告诉我题目"
	}
	if strings.Contains(msgLower, "thanks") || strings.Contains(msgLower, "谢谢") || strings.Contains(msgLower, "thank") {
		return "不客气！有任何编程问题随时问我 😊"
	}

	return l.defaultResponse()
}

// defaultResponse 无明确意图时的默认响应
func (l *ChatLogic) defaultResponse() string {
	var sb strings.Builder
	sb.WriteString("👋 你好！我是智能编程学习助手。\n\n")
	sb.WriteString("我可以帮你：\n\n")
	sb.WriteString("1. **分析代码错误** — 发送包含错误信息和代码的消息\n")
	sb.WriteString("   示例: `我的代码报错了 TypeError: ...`\n\n")
	sb.WriteString("2. **查询语法概念** — 询问编程语言语法\n")
	sb.WriteString("   示例: `解释一下 Python 的装饰器`\n\n")
	sb.WriteString("3. **生成解题方案** — 描述编程问题，获取解题思路和代码\n")
	sb.WriteString("   示例: `用 Go 实现 two sum 算法`\n\n")

	if l.svcCtx.MCPClient != nil {
		allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
		if len(allTools) > 0 {
			sb.WriteString("---\n**可用工具**:\n")
			for serverName, tools := range allTools {
				for _, t := range tools {
					sb.WriteString("- `" + t.Name + "` (" + serverName + "): " + t.Description + "\n")
				}
			}
		}
	}

	return sb.String()
}

// formatToolResult 格式化工具调用结果为可读文本
func formatToolResult(result interface{}, title string) string {
	var sb strings.Builder
	sb.WriteString("## " + title + "\n\n")

	if result == nil {
		return sb.String() + "(无结果)"
	}

	// mcp.CallToolResult 包含 Content 数组
	type contentItem struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type toolResult struct {
		Content []contentItem `json:"content"`
	}

	// 先尝试序列化整体结果
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return sb.String() + fmt.Sprintf("%v", result)
	}

	var tr toolResult
	if err := json.Unmarshal(jsonBytes, &tr); err == nil {
		for _, c := range tr.Content {
			if c.Type == "text" {
				// 尝试美化 JSON 文本
				var parsed interface{}
				if json.Unmarshal([]byte(c.Text), &parsed) == nil {
					prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
					sb.WriteString(string(prettyJSON))
					sb.WriteString("\n")
				} else {
					sb.WriteString(c.Text)
					sb.WriteString("\n")
				}
			}
		}
		return sb.String()
	}

	return sb.String() + fmt.Sprintf("%v", result)
}

// === 辅助函数 ===

// extractCodeBlock 从消息中提取代码块
func extractCodeBlock(msg string) string {
	// 查找 ``` 包围的代码块
	start := strings.Index(msg, "```")
	if start == -1 {
		return msg // 无代码块则返回全部文本
	}
	start += 3
	// 跳过语言标识
	if idx := strings.Index(msg[start:], "\n"); idx != -1 {
		start += idx + 1
	}
	end := strings.Index(msg[start:], "```")
	if end == -1 {
		return msg[start:]
	}
	return msg[start : start+end]
}

// extractErrorMsg 从消息中提取错误信息
func extractErrorMsg(msg string) string {
	errorMarkers := []string{"Error:", "error:", "错误:", "错误信息:", "报错:", "panic:", "Exception:", "Traceback"}
	for _, marker := range errorMarkers {
		if idx := strings.Index(msg, marker); idx != -1 {
			rest := msg[idx:]
			if end := strings.Index(rest, "\n\n"); end != -1 {
				return rest[:end]
			}
			return rest
		}
	}
	return msg
}

// detectLanguage 检测消息中提到的编程语言
func detectLanguage(msg string) string {
	msgLower := strings.ToLower(msg)
	languages := []string{"python", "go", "golang", "javascript", "js", "typescript", "ts", "java", "rust", "c++", "cpp", "c#", "csharp", "ruby", "php", "swift", "kotlin", "scala"}
	for _, lang := range languages {
		if strings.Contains(msgLower, lang) {
			switch lang {
			case "golang":
				return "go"
			case "js":
				return "javascript"
			case "ts":
				return "typescript"
			case "cpp":
				return "c++"
			case "csharp":
				return "c#"
			default:
				return lang
			}
		}
	}
	return "go" // 默认
}

// extractLangAndQuery 从消息中提取语言和查询概念
func extractLangAndQuery(msg string) (string, string) {
	lang := detectLanguage(msg)
	msgLower := strings.ToLower(msg)

	// 常见查询模式: "解释 X 的 Y", "what is X in Go", "如何使用 X"
	query := msg // 默认为整个消息

	// 尝试提取具体查询概念
	concepts := []string{"decorator", "async", "await", "closure", "闭包", "goroutine", "channel",
		"generator", "interface", "接口", "pointer", "指针", "inheritance", "继承",
		"polymorphism", "多态", "recursion", "递归", "generic", "泛型"}
	for _, c := range concepts {
		if strings.Contains(msgLower, c) {
			query = c
			break
		}
	}

	return lang, query
}

// === 语义记忆 ===

const memoryCollection = "chat_history"

// recallSimilarHistory 搜索与当前消息语义相似的历史对话
// 失败时返回空字符串，不影响主流程
func (l *ChatLogic) recallSimilarHistory(userId, message string) string {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return ""
	}

	// 将用户消息向量化
	queryVec, err := l.svcCtx.Embedding.Embed(l.ctx, message)
	if err != nil {
		log.Printf("[Memory] 向量化失败(召回): %v", err)
		return ""
	}

	// 搜索相似历史
	resp, err := l.svcCtx.MemoryRpc.SearchSimilar(l.ctx, &memorypb.SearchSimilarRequest{
		Collection:  memoryCollection,
		QueryVector: queryVec,
		TopK:        3,
	})
	if err != nil {
		log.Printf("[Memory] 相似搜索失败: %v", err)
		return ""
	}

	if !resp.Success || len(resp.Results) == 0 {
		return ""
	}

	// 格式化历史为上下文文本
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

// rememberMessage 异步将消息和回答存入向量记忆
func (l *ChatLogic) rememberMessage(userId, message, response string) {
	if l.svcCtx.Embedding == nil || l.svcCtx.MemoryRpc == nil {
		return
	}

	// 使用独立 context，避免请求结束后被取消
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

// messageID 生成幂等的向量 ID
func messageID(userId, message string) int64 {
	h := fnv.New64a()
	h.Write([]byte(userId + "|" + message))
	return int64(h.Sum64())
}

// mergeContext 合并原始上下文和召回的历史上下文
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
