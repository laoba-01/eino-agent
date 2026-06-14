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

// DeleteContext 带可选乐观锁的原子删除（Lua 脚本, 1 RTT）
//
// - 指定 keys → 删除指定字段（HDEL）
// - 不指定 keys → 删除整个 context key（DEL）
// - expected_version > 0 → CAS 检查后才执行
func (l *DeleteContextLogic) DeleteContext(in *pb.DeleteContextRequest) (*pb.DeleteContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.DeleteContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	// ARGV[1] = expected_version, ARGV[2..N] = keys to delete (optional)
	args := make([]interface{}, 0, 1+len(in.Keys))
	args = append(args, in.ExpectedVersion)
	for _, k := range in.Keys {
		args = append(args, k)
	}

	result, err := deleteScript.Run(l.ctx, l.svcCtx.Redis, []string{key}, args...).Result()
	if err != nil {
		return &pb.DeleteContextResponse{Success: false, Error: err.Error()}, nil
	}

	vals := result.([]interface{})
	status := vals[0].(int64)

	if status == 0 {
		// 版本冲突
		return &pb.DeleteContextResponse{
			Success:        false,
			Error:          "VERSION_CONFLICT",
			CurrentVersion: vals[1].(int64),
		}, nil
	}

	// status == 1, vals[1] = deleted count
	return &pb.DeleteContextResponse{Success: true}, nil
}
