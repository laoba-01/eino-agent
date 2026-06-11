package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
	"smart-coding-assistant/pkg/llm"
)

type QuerySyntaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuerySyntaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuerySyntaxLogic {
	return &QuerySyntaxLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *QuerySyntaxLogic) QuerySyntax(in *pb.QuerySyntaxRequest) (*pb.QuerySyntaxResponse, error) {
	systemPrompt := fmt.Sprintf(
		"你是一位%s语言专家。解释用户询问的语法概念，给出清晰的定义和实用的代码示例。"+
			"回复格式：先解释概念，然后附上一个具体代码示例。",
		in.Language,
	)

	userMessage := fmt.Sprintf("请解释 %s 中的语法概念: %s", in.Language, in.Query)
	if in.Context != "" {
		userMessage += fmt.Sprintf("\n\n用户上下文: %s", in.Context)
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	response, err := l.svcCtx.LLMClient.Chat(l.ctx, messages)
	if err != nil {
		return &pb.QuerySyntaxResponse{
			Explanation: "",
			Example:     "",
			Success:     false,
		}, fmt.Errorf("llm查询语法: %w", err)
	}

	return &pb.QuerySyntaxResponse{
		Explanation: response,
		Example:     response,
		Success:     true,
	}, nil
}
