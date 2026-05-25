package logic

import (
	"context"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
)

type QuerySyntaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuerySyntaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySyntaxLogic {
	return &QuerySyntaxLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *QuerySyntaxLogic) QuerySyntax(in *pb.QuerySyntaxRequest) (*pb.QuerySyntaxResponse, error) {
	return &pb.QuerySyntaxResponse{
		Explanation: "Syntax explanation placeholder",
		Example:     "Syntax example placeholder",
		Success:     true,
	}, nil
}
