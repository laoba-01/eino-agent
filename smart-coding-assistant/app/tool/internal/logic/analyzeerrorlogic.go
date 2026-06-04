package logic

import (
	"context"
	"fmt"
	"strings"

	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"
)

type AnalyzeCodeErrorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnalyzeCodeErrorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeCodeErrorLogic {
	return &AnalyzeCodeErrorLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *AnalyzeCodeErrorLogic) AnalyzeCodeError(in *pb.AnalyzeCodeErrorRequest) (*pb.AnalyzeCodeErrorResponse, error) {
	analysis, suggestedFix := analyzeError(in.GetCode(), in.GetErrorMessage(), in.GetLanguage())
	return &pb.AnalyzeCodeErrorResponse{
		Analysis:     analysis,
		SuggestedFix: suggestedFix,
		Success:      true,
	}, nil
}

func analyzeError(code, errMsg, lang string) (string, string) {
	errLower := strings.ToLower(errMsg)

	// 构建分析报告
	var sb strings.Builder
	sb.WriteString("## 错误分析\n\n")
	sb.WriteString(fmt.Sprintf("**语言**: %s\n", lang))
	sb.WriteString(fmt.Sprintf("**错误信息**: %s\n\n", errMsg))
	if code != "" {
		sb.WriteString(fmt.Sprintf("**代码**:\n```%s\n%s\n```\n\n", lang, code))
	}

	// 基于错误信息关键词进行模式匹配
	switch {
	case matchAny(errLower, "undefined", "is not defined", "nameerror", "unresolved"):
		sb.WriteString("**根因**: 变量或标识符未定义。代码中使用了尚未声明或不在作用域内的名称。\n")
		sb.WriteString(fmt.Sprintf("**影响**: %s 编译器/解释器无法解析该标识符。\n", lang))

	case matchAny(errLower, "type", "typeerror", "cannot convert", "incompatible"):
		sb.WriteString("**根因**: 类型不匹配。操作符或函数接收到了预期之外的数据类型。\n")
		sb.WriteString("**影响**: 运行时错误，可能在特定输入条件下触发。\n")

	case matchAny(errLower, "nil", "null", "nullpointerexception", "nonetype", "panic", "nil pointer"):
		sb.WriteString("**根因**: 空指针/空值引用。在未检查 nil/null 的情况下访问了对象的属性或方法。\n")
		sb.WriteString("**影响**: 运行时 panic 或异常。\n")

	case matchAny(errLower, "index", "indexerror", "out of range", "arrayindexoutofbounds"):
		sb.WriteString("**根因**: 数组/切片/列表索引越界。访问了超出集合长度范围的索引位置。\n")
		sb.WriteString("**影响**: 运行时错误，边界条件未正确处理。\n")

	case matchAny(errLower, "syntax", "syntaxerror", "unexpected", "expected"):
		sb.WriteString("**根因**: 语法错误。代码不符合编程语言的语法规则。\n")
		sb.WriteString(fmt.Sprintf("**影响**: %s 编译器/解释器无法解析该代码。\n", lang))

	default:
		sb.WriteString("**根因**: 需要进一步分析错误上下文和调用堆栈来确定具体原因。\n")
		sb.WriteString("**影响**: 建议添加更详细的日志记录来追踪错误路径。\n")
	}

	// 构建修复建议
	var fix strings.Builder
	fix.WriteString("\n## 修复建议\n\n")

	switch {
	case matchAny(errLower, "undefined", "is not defined", "nameerror", "unresolved"):
		fix.WriteString("1. 检查变量声明：确保在使用前已声明变量\n")
		fix.WriteString("2. 检查作用域：确认变量在当前作用域内是可见的\n")
		fix.WriteString("3. 检查导入：如有跨文件引用，确认 import/include 语句正确\n")

	case matchAny(errLower, "type", "typeerror", "cannot convert", "incompatible"):
		fix.WriteString("1. 添加显式类型转换（如 int()、str()、.toString()）\n")
		fix.WriteString("2. 检查函数签名，确认传入参数类型与预期一致\n")
		fix.WriteString("3. 使用类型断言或类型守卫处理接口类型\n")

	case matchAny(errLower, "nil", "null", "nullpointerexception", "nonetype", "panic", "nil pointer"):
		fix.WriteString("1. 在访问前添加 nil/null 检查：`if obj != nil { ... }`\n")
		fix.WriteString("2. 使用可选链操作符（如果语言支持，如 `obj?.field`）\n")
		fix.WriteString("3. 为变量设置合理的默认值，避免 nil 状态\n")

	case matchAny(errLower, "index", "indexerror", "out of range", "arrayindexoutofbounds"):
		fix.WriteString("1. 访问前检查索引：`if idx >= 0 && idx < len(arr) { ... }`\n")
		fix.WriteString("2. 优先使用 for-range 遍历，避免手动索引\n")
		fix.WriteString("3. 确认循环边界条件正确\n")

	case matchAny(errLower, "syntax", "syntaxerror", "unexpected", "expected"):
		fix.WriteString("1. 检查括号/引号是否成对闭合\n")
		fix.WriteString("2. 检查语句末尾的分号/缩进是否符合语言规范\n")
		fix.WriteString("3. 检查关键字拼写是否正确\n")

	default:
		fix.WriteString("1. 仔细阅读错误信息和堆栈跟踪，定位出错的具体行号\n")
		fix.WriteString("2. 添加单元测试覆盖出错场景，便于复现和修复\n")
		fix.WriteString("3. 搜索该错误信息，查看社区中类似问题的解决方案\n")
	}

	return sb.String(), fix.String()
}

// matchAny 检查 s 是否包含 keywords 中的任意一个
func matchAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
