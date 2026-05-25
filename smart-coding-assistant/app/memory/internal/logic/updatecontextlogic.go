package logic

import (
	"context"
	"fmt"
	"time"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type UpdateContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContextLogic {
	return &UpdateContextLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateContextLogic) UpdateContext(in *pb.UpdateContextRequest) (*pb.UpdateContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.UpdateContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)
	pipe := l.svcCtx.Redis.Pipeline()
	pipe.HSet(l.ctx, key, mapToInterface(in.Context)...)
	if in.Ttl > 0 {
		pipe.Expire(l.ctx, key, time.Duration(in.Ttl)*time.Second)
	}
	if _, err := pipe.Exec(l.ctx); err != nil {
		return &pb.UpdateContextResponse{Success: false, Error: err.Error()}, nil
	}

	return &pb.UpdateContextResponse{Success: true}, nil
}
