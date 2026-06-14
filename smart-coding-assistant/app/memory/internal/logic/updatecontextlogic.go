package logic

import (
	"context"
	"fmt"

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

// UpdateContext 带乐观锁的原子更新（CAS Lua 脚本, 1 RTT）
//
// 流程:
//  1. 调用方先 GetContext → 拿到 version
//  2. 本地修改数据
//  3. UpdateContext(expected_version=version) → CAS 检查
//  4. 版本冲突 → 返回 current_version，调用方重试
func (l *UpdateContextLogic) UpdateContext(in *pb.UpdateContextRequest) (*pb.UpdateContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.UpdateContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	// 构建 ARGV: [expected_version, ttl, k1, v1, k2, v2, ...]
	fields := mapToSlice(in.Context)
	args := make([]interface{}, 0, 2+len(fields))
	args = append(args, in.ExpectedVersion, in.Ttl)
	for _, f := range fields {
		args = append(args, f)
	}

	result, err := updateScript.Run(l.ctx, l.svcCtx.Redis, []string{key}, args...).Result()
	if err != nil {
		return &pb.UpdateContextResponse{Success: false, Error: err.Error()}, nil
	}

	vals := result.([]interface{})
	if vals[0].(int64) == 0 {
		// 版本冲突
		return &pb.UpdateContextResponse{
			Success:        false,
			Error:          "VERSION_CONFLICT",
			CurrentVersion: vals[1].(int64),
		}, nil
	}

	return &pb.UpdateContextResponse{
		Success:    true,
		NewVersion: vals[1].(int64),
	}, nil
}
