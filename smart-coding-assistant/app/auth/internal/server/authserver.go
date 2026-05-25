package server

import (
	"context"

	"smart-coding-assistant/app/auth/internal/logic"
	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	svcCtx *svc.ServiceContext
}

func NewAuthServer(svcCtx *svc.ServiceContext) *AuthServer {
	return &AuthServer{svcCtx: svcCtx}
}

func (s *AuthServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	l := logic.NewRegisterLogic(ctx, s.svcCtx)
	return l.Register(in)
}

func (s *AuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	l := logic.NewLoginLogic(ctx, s.svcCtx)
	return l.Login(in)
}

func (s *AuthServer) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	l := logic.NewValidateTokenLogic(ctx, s.svcCtx)
	return l.ValidateToken(in)
}

func (s *AuthServer) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	l := logic.NewLogoutLogic(ctx, s.svcCtx)
	return l.Logout(in)
}
