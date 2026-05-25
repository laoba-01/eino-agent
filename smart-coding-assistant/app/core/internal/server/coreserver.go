package server

import (
	"context"

	"smart-coding-assistant/app/core/internal/logic"
	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

type CoreServer struct {
	pb.UnimplementedCoreServiceServer
	svcCtx *svc.ServiceContext
}

func NewCoreServer(svcCtx *svc.ServiceContext) *CoreServer {
	return &CoreServer{svcCtx: svcCtx}
}

func (s *CoreServer) Chat(ctx context.Context, in *pb.ChatRequest) (*pb.ChatResponse, error) {
	l := logic.NewChatLogic(ctx, s.svcCtx)
	return l.Chat(in)
}

func (s *CoreServer) GetHistory(ctx context.Context, in *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	l := logic.NewGetHistoryLogic(ctx, s.svcCtx)
	return l.GetHistory(in)
}
