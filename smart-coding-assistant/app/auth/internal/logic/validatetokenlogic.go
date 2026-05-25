package logic

import (
	"context"

	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ValidateTokenLogic) ValidateToken(in *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, valid := validateToken(l.svcCtx.Redis, l.svcCtx.Config.JWT.Secret, in.Token)
	if !valid {
		return &pb.ValidateTokenResponse{Valid: false, Message: "Invalid token"}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:   true,
		UserId:  userID,
		Message: "Token is valid",
	}, nil
}
