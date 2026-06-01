package logic

import (
	"context"
	"fmt"
	"strings"

	"smart-coding-assistant/app/core/internal/executor"
	"smart-coding-assistant/app/core/internal/planner"
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
	// 1. 收集可用工具列表
	availableTools := l.collectTools()

	// 2. Planner 生成执行计划
	plan, err := l.svcCtx.Planner.GeneratePlan(l.ctx, in.Message, in.Context, availableTools)
	if err != nil {
		return &pb.ChatResponse{
			Response:   fmt.Sprintf("计划生成失败: %v", err),
			IsFinished: true,
		}, nil
	}

	// 3. 构建初始响应（计划概览）
	resp := &pb.ChatResponse{
		IsFinished: false,
		Plan: &pb.PlanInfo{
			PlanId:     plan.ID,
			Goal:       plan.Goal,
			TotalSteps: int32(len(plan.Steps)),
		},
		Context: in.Context,
	}

	// 4. Executor 逐步执行
	reporter := &chatReporter{resp: resp, totalSteps: len(plan.Steps)}
	if err := l.svcCtx.Executor.Execute(l.ctx, plan, reporter); err != nil {
		resp.Response = fmt.Sprintf("执行终止: %v", err)
		resp.IsFinished = true
		return resp, nil
	}

	// 5. 全部完成
	resp.Response = l.buildFinalResponse(plan)
	resp.IsFinished = true
	return resp, nil
}

// collectTools 收集可用的 MCP 工具名称列表
func (l *ChatLogic) collectTools() []string {
	var tools []string
	if l.svcCtx.MCPClient == nil {
		return tools
	}
	allTools := l.svcCtx.MCPClient.ListAllTools(l.ctx)
	for _, serverTools := range allTools {
		for _, t := range serverTools {
			tools = append(tools, t.Name)
		}
	}
	return tools
}

// buildFinalResponse 构建最终响应文本
func (l *ChatLogic) buildFinalResponse(plan *planner.Plan) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s\n\n", plan.Goal)
	for _, step := range plan.Steps {
		icon := "✅"
		if step.Status == string(planner.StepStatusFailed) {
			icon = "❌"
		}
		fmt.Fprintf(&sb, "%s **步骤%d**: %s\n", icon, step.Index, step.Description)
		if step.Result != "" {
			fmt.Fprintf(&sb, "   %s\n\n", step.Result)
		}
		if step.Error != "" {
			fmt.Fprintf(&sb, "   错误: %s\n\n", step.Error)
		}
	}
	return sb.String()
}

// chatReporter 实现 executor.StepReporter，将步骤进度写入 ChatResponse
type chatReporter struct {
	resp       *pb.ChatResponse
	totalSteps int
}

func (r *chatReporter) OnStepStart(step planner.Step) {
	r.resp.Response = fmt.Sprintf("正在执行步骤 %d/%d: %s", step.Index, r.totalSteps, step.Description)
}

func (r *chatReporter) OnStepDone(step planner.Step) {
	r.resp.Steps = append(r.resp.Steps, &pb.StepResult{
		Index:       int32(step.Index),
		Description: step.Description,
		Status:      step.Status,
		Result:      step.Result,
		Error:       step.Error,
	})
	r.resp.Plan.CompletedSteps = int32(len(r.resp.Steps))
}

func (r *chatReporter) OnAllDone(plan planner.Plan) {
	r.resp.Plan.CompletedSteps = int32(r.totalSteps)
}

var _ executor.StepReporter = (*chatReporter)(nil)
