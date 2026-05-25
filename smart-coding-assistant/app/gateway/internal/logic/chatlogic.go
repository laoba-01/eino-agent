package logic

import (
	"encoding/json"
	"net/http"

	corepb "smart-coding-assistant/app/core/pb"
	"smart-coding-assistant/app/gateway/internal/svc"
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

	var req struct {
		Message string            `json:"message"`
		Context map[string]string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
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
