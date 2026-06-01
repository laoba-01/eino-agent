package executor

import "smart-coding-assistant/app/core/internal/planner"

// StepReporter 步骤执行回调接口
type StepReporter interface {
	OnStepStart(step planner.Step)
	OnStepDone(step planner.Step)
	OnAllDone(plan planner.Plan)
}
