package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type DeleteContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteContextLogic {
	return &DeleteContextLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteContextLogic) DeleteContext(in *pb.DeleteContextRequest) (*pb.DeleteContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.DeleteContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	if len(in.Keys) > 0 {
		if err := l.svcCtx.Redis.HDel(l.ctx, key, in.Keys...).Err(); err != nil {
			return &pb.DeleteContextResponse{Success: false, Error: err.Error()}, nil
		}
	} else {
		if err := l.svcCtx.Redis.Del(l.ctx, key).Err(); err != nil {
			return &pb.DeleteContextResponse{Success: false, Error: err.Error()}, nil
		}
	}

	return &pb.DeleteContextResponse{Success: true}, nil
}
