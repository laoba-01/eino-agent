package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type GetContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContextLogic {
	return &GetContextLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GetContextLogic) GetContext(in *pb.GetContextRequest) (*pb.GetContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.GetContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	if len(in.Keys) > 0 {
		vals, err := l.svcCtx.Redis.HMGet(l.ctx, key, in.Keys...).Result()
		if err != nil {
			return &pb.GetContextResponse{Success: false, Error: err.Error()}, nil
		}
		result := make(map[string]string)
		for i, k := range in.Keys {
			if vals[i] != nil {
				result[k] = vals[i].(string)
			}
		}
		return &pb.GetContextResponse{Context: result, Success: true}, nil
	}

	result, err := l.svcCtx.Redis.HGetAll(l.ctx, key).Result()
	if err != nil {
		return &pb.GetContextResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.GetContextResponse{Context: result, Success: true}, nil
}
