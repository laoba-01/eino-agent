package logic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

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
		Message  string            `json:"message"`
		Context  map[string]string `json:"context"`
		Language string            `json:"language"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 所有请求透传到核心服务（不再在网关层拦截）
	resp, err := l.svcCtx.CoreRpc.Chat(r.Context(), &corepb.ChatRequest{
		UserId:   userID,
		Message:  req.Message,
		Context:  req.Context,
		Language: req.Language,
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


