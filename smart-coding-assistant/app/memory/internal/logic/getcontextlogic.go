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

// GetContext 读取上下文，同时返回 _version 供调用方后续 CAS 写入
func (l *GetContextLogic) GetContext(in *pb.GetContextRequest) (*pb.GetContextResponse, error) {
	if l.svcCtx.Redis == nil {
		return &pb.GetContextResponse{Success: false, Error: "redis not connected"}, nil
	}

	key := fmt.Sprintf("context:%s", in.UserId)

	var result map[string]string
	var err error

	if len(in.Keys) > 0 {
		// 指定 key 读取: HMGET
		// 总是额外读取 _version
		fetchKeys := append(in.Keys, "_version")
		vals, e := l.svcCtx.Redis.HMGet(l.ctx, key, fetchKeys...).Result()
		if e != nil {
			return &pb.GetContextResponse{Success: false, Error: e.Error()}, nil
		}
		result = make(map[string]string, len(in.Keys))
		for i, k := range in.Keys {
			if vals[i] != nil {
				result[k] = toString(vals[i])
			}
		}
		// 版本号单独提取
		if vals[len(in.Keys)] != nil {
			result["_version"] = toString(vals[len(in.Keys)])
		}
		err = nil
	} else {
		// 全量读取: HGETALL
		result, err = l.svcCtx.Redis.HGetAll(l.ctx, key).Result()
	}

	if err != nil {
		return &pb.GetContextResponse{Success: false, Error: err.Error()}, nil
	}

	// 提取 _version 并移除（不暴露给调用方作为普通 context field）
	version := strToInt64(result["_version"])
	delete(result, "_version")

	return &pb.GetContextResponse{Context: result, Success: true, Version: version}, nil
}

// toString safely converts interface{} to string.
func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", s)
	}
}
