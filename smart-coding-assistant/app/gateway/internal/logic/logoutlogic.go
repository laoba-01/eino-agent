package logic

import (
	"encoding/json"
	"net/http"
	"strings"

	authpb "smart-coding-assistant/app/auth/pb"
	"smart-coding-assistant/app/gateway/internal/svc"
)

type LogoutLogic struct {
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{svcCtx: svcCtx}
}

func (l *LogoutLogic) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	resp, err := l.svcCtx.AuthRpc.Logout(r.Context(), &authpb.LogoutRequest{
		Token: token,
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
