package logic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	corepb "smart-coding-assistant/app/core/pb"
	"smart-coding-assistant/app/gateway/internal/svc"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type ChatLogic struct {
	svcCtx *svc.ServiceContext
}

func NewChatLogic(svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{svcCtx: svcCtx}
}

func (l *ChatLogic) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	userID := r.Context().Value("user_id").(string)

	// 在 JSON 解码之前读取原始 body，处理 GBK→UTF-8 转码
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read request body"})
		return
	}

	// 检测并修复 GBK 编码（Windows 终端 curl 常见问题）
	bodyBytes = fixGarbledGBK(bodyBytes)

	var req struct {
		Message string            `json:"message"`
		Context map[string]string `json:"context"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 在网关层拦截系统/闲聊问题，直接返回（避免后端中文编码问题）
	if sysResp := handleSystemQuestion(req.Message); sysResp != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(corepb.ChatResponse{
			Response:   sysResp,
			IsFinished: true,
			Context:    req.Context,
		})
		return
	}

	resp, err := l.svcCtx.CoreRpc.Chat(r.Context(), &corepb.ChatRequest{
		UserId:  userID,
		Message: req.Message,
		Context: req.Context,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// fixGarbledGBK 在 JSON 解码前检测并转换 GBK→UTF-8
func fixGarbledGBK(raw []byte) []byte {
	// 如果已经是合法 UTF-8，直接返回
	if isUTF8(raw) {
		return raw
	}

	// 尝试将字节解释为 GBK 并转换为 UTF-8
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Bytes, err := io.ReadAll(transform.NewReader(bytes.NewReader(raw), decoder))
	if err != nil {
		return raw
	}

	// 验证转换后的结果包含合法 JSON 结构
	if bytes.Contains(utf8Bytes, []byte(`"message"`)) || bytes.Contains(utf8Bytes, []byte(`"context"`)) {
		return utf8Bytes
	}

	return raw
}

// isUTF8 检查字节是否已经是合法 UTF-8
func isUTF8(data []byte) bool {
	// 简单检查：如果已经包含常见的中文 UTF-8 序列，说明不需要转换
	for i := 0; i < len(data)-2; i++ {
		b := data[i]
		if b >= 0xE4 && b <= 0xE9 {
			// UTF-8 中文首字节范围 (U+4E00-U+9FFF 的 UTF-8 编码以 E4-E9 开头)
			if data[i+1] >= 0x80 && data[i+1] <= 0xBF &&
				data[i+2] >= 0x80 && data[i+2] <= 0xBF {
				return true
			}
		}
	}
	return false
}

// handleSystemQuestion 在网关层处理系统/闲聊问题
func handleSystemQuestion(msg string) string {
	msgLower := strings.ToLower(msg)

	// 模型相关
	if strings.Contains(msgLower, "模型") || strings.Contains(msgLower, "model") ||
		strings.Contains(msgLower, "什么模型") || strings.Contains(msgLower, "你用的是什么") {
		return "我目前是基于规则匹配的编程学习助手，尚未接入大语言模型。\n\n" +
			"我能通过关键词识别你的意图，并调用对应的工具：\n" +
			"- 分析代码错误 — 发送报错信息帮你定位问题\n" +
			"- 查询编程语法 — 解释编程语言的各种概念\n" +
			"- 生成解题方案 — 提供算法思路和代码实现\n\n" +
			"未来接入 LLM 后会有更强的对话能力！"
	}

	// 身份相关
	if strings.Contains(msgLower, "你是谁") || strings.Contains(msgLower, "你叫什么") ||
		strings.Contains(msgLower, "你的名字") || strings.Contains(msgLower, "你是什么") ||
		strings.Contains(msgLower, "who are you") {
		return "我是**智能编程学习助手**，一个帮助你学习编程的 AI Agent。\n\n可以分析代码错误、解释语法概念、生成解题方案。试试问我编程问题吧！"
	}

	// 功能相关
	if strings.Contains(msgLower, "你能做什么") || strings.Contains(msgLower, "有什么功能") ||
		strings.Contains(msgLower, "help") || strings.Contains(msgLower, "功能") {
		return "我能帮你：\n\n" +
			"1. 分析代码错误 — 发送包含错误信息的消息\n" +
			"   示例: 我的 Python 代码报错 TypeError: ...\n\n" +
			"2. 查询语法概念 — 询问编程语言语法\n" +
			"   示例: 解释一下 Go 的 goroutine\n\n" +
			"3. 生成解题方案 — 描述编程问题\n" +
			"   示例: 用 Go 实现 two sum 算法"
	}

	// 打招呼
	if msgLower == "hi" || msgLower == "hello" || msgLower == "你好" ||
		strings.Contains(msgLower, "在吗") {
		return "你好！有什么编程问题需要帮助吗？\n\n" +
			"- 代码报错了？把错误信息发给我\n" +
			"- 想了解某个语法？直接问我\n" +
			"- 需要解题思路？告诉我题目"
	}

	// 感谢
	if strings.Contains(msgLower, "谢谢") || strings.Contains(msgLower, "感谢") ||
		strings.Contains(msgLower, "thank") || strings.Contains(msgLower, "thanks") {
		return "不客气！有问题随时问我"
	}

	return ""
}
