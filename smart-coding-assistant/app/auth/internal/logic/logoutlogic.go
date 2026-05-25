package logic

import (
	"context"

	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *LogoutLogic) Logout(in *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := l.svcCtx.Redis.Del(l.ctx, "token:"+in.Token).Err()
	if err != nil {
		return &pb.LogoutResponse{Success: false, Message: "Failed to logout"}, err
	}

	return &pb.LogoutResponse{Success: true, Message: "Logout successful"}, nil
}
