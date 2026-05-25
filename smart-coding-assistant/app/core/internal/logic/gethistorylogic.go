package logic

import (
	"context"

	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

type GetHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryLogic {
	return &GetHistoryLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GetHistoryLogic) GetHistory(in *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	return &pb.GetHistoryResponse{
		Messages: []*pb.ChatMessage{},
	}, nil
}
