package logic

import (
	"context"
	"fmt"

	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ChatLogic) Chat(in *pb.ChatRequest) (*pb.ChatResponse, error) {
	var toolInfo string
	if l.svcCtx.MCPClient != nil {
		allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
		if len(allTools) > 0 {
			toolInfo = "\n可用 MCP 工具:\n"
			for serverName, tools := range allTools {
				for _, t := range tools {
					toolInfo += fmt.Sprintf("- [%s] %s: %s\n", serverName, t.Name, t.Description)
				}
			}
		}
	}

	return &pb.ChatResponse{
		Response:   "Hello from Core Service!" + toolInfo,
		IsFinished: true,
		Context:    in.Context,
	}, nil
}
