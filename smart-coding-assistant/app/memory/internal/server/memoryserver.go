package server

import (
	"context"

	"smart-coding-assistant/app/memory/internal/logic"
	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"
)

type MemoryServer struct {
	pb.UnimplementedMemoryServiceServer
	svcCtx *svc.ServiceContext
}

func NewMemoryServer(svcCtx *svc.ServiceContext) *MemoryServer {
	return &MemoryServer{svcCtx: svcCtx}
}

func (s *MemoryServer) SaveContext(ctx context.Context, in *pb.SaveContextRequest) (*pb.SaveContextResponse, error) {
	l := logic.NewSaveContextLogic(ctx, s.svcCtx)
	return l.SaveContext(in)
}

func (s *MemoryServer) GetContext(ctx context.Context, in *pb.GetContextRequest) (*pb.GetContextResponse, error) {
	l := logic.NewGetContextLogic(ctx, s.svcCtx)
	return l.GetContext(in)
}

func (s *MemoryServer) DeleteContext(ctx context.Context, in *pb.DeleteContextRequest) (*pb.DeleteContextResponse, error) {
	l := logic.NewDeleteContextLogic(ctx, s.svcCtx)
	return l.DeleteContext(in)
}

func (s *MemoryServer) UpdateContext(ctx context.Context, in *pb.UpdateContextRequest) (*pb.UpdateContextResponse, error) {
	l := logic.NewUpdateContextLogic(ctx, s.svcCtx)
	return l.UpdateContext(in)
}

func (s *MemoryServer) SaveVector(ctx context.Context, in *pb.SaveVectorRequest) (*pb.SaveVectorResponse, error) {
	l := logic.NewSaveVectorLogic(ctx, s.svcCtx)
	return l.SaveVector(in)
}

func (s *MemoryServer) SearchSimilar(ctx context.Context, in *pb.SearchSimilarRequest) (*pb.SearchSimilarResponse, error) {
	l := logic.NewSearchSimilarLogic(ctx, s.svcCtx)
	return l.SearchSimilar(in)
}

func (s *MemoryServer) DeleteVector(ctx context.Context, in *pb.DeleteVectorRequest) (*pb.DeleteVectorResponse, error) {
	l := logic.NewDeleteVectorLogic(ctx, s.svcCtx)
	return l.DeleteVector(in)
}
