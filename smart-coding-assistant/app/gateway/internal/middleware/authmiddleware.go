package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	authpb "smart-coding-assistant/app/auth/pb"
)

type AuthMiddleware struct {
	authRpc authpb.AuthServiceClient
}

func NewAuthMiddleware(authRpc authpb.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{authRpc: authRpc}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authorization token required"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		resp, err := m.authRpc.ValidateToken(r.Context(), &authpb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil || !resp.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired token"})
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", resp.UserId)
		next(w, r.WithContext(ctx))
	}
}
