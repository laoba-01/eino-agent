package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
	"smart-coding-assistant/pkg/llm"
)

type AnalyzeCodeErrorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnalyzeCodeErrorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeCodeErrorLogic {
	return &AnalyzeCodeErrorLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *AnalyzeCodeErrorLogic) AnalyzeCodeError(in *pb.AnalyzeCodeErrorRequest) (*pb.AnalyzeCodeErrorResponse, error) {
	systemPrompt := fmt.Sprintf(
		"你是一位资深软件工程师。分析给出的%s代码和错误信息，找出根因并提供可操作的修复方案。"+
			"回复格式：先简洁描述错误根因，然后给出具体的修复代码或步骤。",
		in.Language,
	)

	userMessage := fmt.Sprintf(
		"源代码:\n```%s\n%s\n```\n\n错误信息:\n%s",
		in.Language, in.Code, in.ErrorMessage,
	)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	response, err := l.svcCtx.LLMClient.Chat(l.ctx, messages)
	if err != nil {
		return &pb.AnalyzeCodeErrorResponse{
			Analysis:     "",
			SuggestedFix: "",
			Success:      false,
		}, fmt.Errorf("llm分析错误: %w", err)
	}

	return &pb.AnalyzeCodeErrorResponse{
		Analysis:     response,
		SuggestedFix: response,
		Success:      true,
	}, nil
}
