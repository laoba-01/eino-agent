package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
	"smart-coding-assistant/pkg/llm"
)

type GenerateProblemSolutionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateProblemSolutionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateProblemSolutionLogic {
	return &GenerateProblemSolutionLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GenerateProblemSolutionLogic) GenerateProblemSolution(in *pb.GenerateProblemSolutionRequest) (*pb.GenerateProblemSolutionResponse, error) {
	systemPrompt := fmt.Sprintf(
		"你是一位算法竞赛教练。针对给定的编程问题，给出%s语言的%s难度解法。"+
			"回复格式：1) 解题思路 2) 完整代码实现 3) 关键步骤解释。",
		in.Language, in.Difficulty,
	)

	userMessage := fmt.Sprintf(
		"请为以下编程问题生成解法:\n\n问题: %s\n难度: %s\n目标语言: %s",
		in.Problem, in.Difficulty, in.Language,
	)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	response, err := l.svcCtx.LLMClient.Chat(l.ctx, messages)
	if err != nil {
		return &pb.GenerateProblemSolutionResponse{
			Approach:    "",
			Code:        "",
			Explanation: "",
			Success:     false,
		}, fmt.Errorf("llm生成题解: %w", err)
	}

	return &pb.GenerateProblemSolutionResponse{
		Approach:    response,
		Code:        response,
		Explanation: response,
		Success:     true,
	}, nil
}
