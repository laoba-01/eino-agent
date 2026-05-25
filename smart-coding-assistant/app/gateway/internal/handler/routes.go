package handler

import (
	"net/http"

	"smart-coding-assistant/app/gateway/internal/logic"
	"smart-coding-assistant/app/gateway/internal/middleware"
	"smart-coding-assistant/app/gateway/internal/svc"
)

func RegisterHandlers(s *http.ServeMux, svcCtx *svc.ServiceContext) {
	cors := middleware.NewCorsMiddleware()
	auth := middleware.NewAuthMiddleware(svcCtx.AuthRpc)

	// 公开路由（CORS 中间件）
	s.HandleFunc("/api/auth/register", cors.Handle(logic.NewRegisterLogic(svcCtx).Handle))
	s.HandleFunc("/api/auth/login", cors.Handle(logic.NewLoginLogic(svcCtx).Handle))
	s.HandleFunc("/health", cors.Handle(logic.NewHealthCheckLogic().Handle))

	// 受保护路由（CORS + Auth 中间件）
	s.HandleFunc("/api/auth/logout", cors.Handle(auth.Handle(logic.NewLogoutLogic(svcCtx).Handle)))
	s.HandleFunc("/api/chat", cors.Handle(auth.Handle(logic.NewChatLogic(svcCtx).Handle)))
}
