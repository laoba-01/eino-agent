package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type DeleteVectorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteVectorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVectorLogic {
	return &DeleteVectorLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteVectorLogic) DeleteVector(in *pb.DeleteVectorRequest) (*pb.DeleteVectorResponse, error) {
	if l.svcCtx.Milvus == nil {
		return &pb.DeleteVectorResponse{Success: false, Error: "milvus not connected"}, nil
	}

	expr := fmt.Sprintf("id in [%s]", joinInt64s(in.Ids))
	if err := l.svcCtx.Milvus.Delete(l.ctx, in.Collection, "", expr); err != nil {
		return &pb.DeleteVectorResponse{Success: false, Error: err.Error()}, nil
	}

	return &pb.DeleteVectorResponse{Success: true}, nil
}
