package logic

import (
	"context"
	"fmt"
	"strconv"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type SaveContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveContextLogic {
	return &SaveContextLogic{ctx: ctx, svcCtx: svcCtx}
}

// SaveContext 原子写入上下文（Lua 脚本: HSET + EXPIRE + _version 自增, 1 RTT）
func (l *SaveContextLogic) SaveContext(in *pb.SaveContextRequest) (*pb.SaveContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.SaveContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	// 构建 ARGV: [ttl, k1, v1, k2, v2, ...]
	fields := mapToSlice(in.Context)
	args := make([]interface{}, 0, 1+len(fields))
	args = append(args, in.Ttl)
	for _, f := range fields {
		args = append(args, f)
	}

	result, err := saveScript.Run(l.ctx, l.svcCtx.Redis, []string{key}, args...).Result()
	if err != nil {
		return &pb.SaveContextResponse{Success: false, Error: err.Error()}, nil
	}

	vals := result.([]interface{})
	if vals[0].(int64) == 0 {
		return &pb.SaveContextResponse{Success: false, Error: fmt.Sprintf("save failed: %v", vals[1])}, nil
	}

	newVersion := vals[1].(int64)
	return &pb.SaveContextResponse{Success: true, NewVersion: newVersion}, nil
}

// strToInt64 converts string to int64, returns 0 on error.
func strToInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
