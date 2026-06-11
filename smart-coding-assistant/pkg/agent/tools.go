package agent

import (
	"context"
	"encoding/json"
	"fmt"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	mcpmgr "smart-coding-assistant/pkg/mcp"
)

// mcpToolAdapter 将已有的 MCP 工具适配为 Eino tool.BaseTool + tool.InvokableTool
type mcpToolAdapter struct {
	info       *schema.ToolInfo
	mcpClient  *mcpmgr.ClientManager
	serverName string
}

func (t *mcpToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.info, nil
}

func (t *mcpToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("mcp tool %q: parse args: %w", t.info.Name, err)
	}

	result, err := t.mcpClient.CallTool(ctx, t.serverName, t.info.Name, args)
	if err != nil {
		return "", fmt.Errorf("mcp tool %q call: %w", t.info.Name, err)
	}

	if result.IsError {
		return "", fmt.Errorf("mcp tool %q returned error", t.info.Name)
	}

	// 提取文本内容
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text, nil
		}
	}

	// 回退：JSON 序列化
	b, _ := json.Marshal(result)
	return string(b), nil
}

// BuildEinoTools 将已有 MCP ClientManager 中的所有工具转换为 Eino tools
func BuildEinoTools(ctx context.Context, mcpClient *mcpmgr.ClientManager) []tool.BaseTool {
	allTools := mcpClient.ListAllTools(ctx)
	var result []tool.BaseTool

	for serverName, tools := range allTools {
		for _, t := range tools {
			result = append(result, &mcpToolAdapter{
				info:       mcpToolToEinoInfo(t),
				mcpClient:  mcpClient,
				serverName: serverName,
			})
		}
	}
	return result
}

// mcpToolToEinoInfo 将 mcp.Tool 转为 Eino ToolInfo
func mcpToolToEinoInfo(t *mcp.Tool) *schema.ToolInfo {
	params := inputSchemaToParams(t.InputSchema)
	return &schema.ToolInfo{
		Name: t.Name,
		Desc: t.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}
}

// inputSchemaToParams 将 MCP InputSchema (map[string]any) 转为 Eino 参数定义
func inputSchemaToParams(raw any) map[string]*schema.ParameterInfo {
	if raw == nil {
		return nil
	}
	schemaMap, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	props, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		return nil
	}

	requiredSet := make(map[string]bool)
	if req, ok := schemaMap["required"].([]any); ok {
		for _, r := range req {
			if s, ok := r.(string); ok {
				requiredSet[s] = true
			}
		}
	}

	result := make(map[string]*schema.ParameterInfo, len(props))
	for key, val := range props {
		propMap, ok := val.(map[string]any)
		if !ok {
			continue
		}
		pi := &schema.ParameterInfo{
			Required: requiredSet[key],
		}
		if desc, ok := propMap["description"].(string); ok {
			pi.Desc = desc
		}
		if typ, ok := propMap["type"].(string); ok {
			pi.Type = mapSchemaType(typ)
		} else {
			pi.Type = schema.String
		}
		result[key] = pi
	}
	return result
}

func mapSchemaType(t string) schema.DataType {
	switch t {
	case "string":
		return schema.String
	case "number", "integer":
		return schema.Integer
	case "boolean":
		return schema.Boolean
	case "array":
		return schema.Array
	case "object":
		return schema.Object
	default:
		return schema.String
	}
}
