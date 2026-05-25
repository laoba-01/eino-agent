package server

import (
	"context"

	"smart-coding-assistant/app/tool/internal/logic"
	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
)

type ToolServer struct {
	pb.UnimplementedToolServiceServer
	svcCtx *svc.ServiceContext
}

func NewToolServer(svcCtx *svc.ServiceContext) *ToolServer {
	return &ToolServer{svcCtx: svcCtx}
}

func (s *ToolServer) AnalyzeCodeError(ctx context.Context, in *pb.AnalyzeCodeErrorRequest) (*pb.AnalyzeCodeErrorResponse, error) {
	l := logic.NewAnalyzeCodeErrorLogic(ctx, s.svcCtx)
	return l.AnalyzeCodeError(in)
}

func (s *ToolServer) QuerySyntax(ctx context.Context, in *pb.QuerySyntaxRequest) (*pb.QuerySyntaxResponse, error) {
	l := logic.NewQuerySyntaxLogic(ctx, s.svcCtx)
	return l.QuerySyntax(in)
}

func (s *ToolServer) GenerateProblemSolution(ctx context.Context, in *pb.GenerateProblemSolutionRequest) (*pb.GenerateProblemSolutionResponse, error) {
	l := logic.NewGenerateProblemSolutionLogic(ctx, s.svcCtx)
	return l.GenerateProblemSolution(in)
}
