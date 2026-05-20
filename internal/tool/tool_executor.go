package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolExecutor 工具执行器，用于在 AI 服务中集成工具调用
type ToolExecutor struct {
	registry *ToolRegistry
}

// NewToolExecutor 创建新的工具执行器
func NewToolExecutor(registry *ToolRegistry) *ToolExecutor {
	return &ToolExecutor{
		registry: registry,
	}
}

// ExecuteToolCall 执行工具调用
// 当 LLM 返回 tool_calls 时，调用此方法执行实际的工具
// 参数 toolCallName: 工具名称
// 参数 toolCallArgs: 工具参数（JSON 字符串或 map）
func (te *ToolExecutor) ExecuteToolCall(ctx context.Context, toolCallName string, toolCallArgs interface{}) (*Result, error) {
	// 解析工具调用参数
	var args map[string]interface{}
	
	switch v := toolCallArgs.(type) {
	case string:
		// JSON 字符串
		if v != "" {
			if err := json.Unmarshal([]byte(v), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}
		}
	case map[string]interface{}:
		// 已经是 map
		args = v
	default:
		return nil, fmt.Errorf("unsupported arguments type: %T", toolCallArgs)
	}

	// 创建 Call 对象
	call := &Call{
		Name:      toolCallName,
		Arguments: args,
	}

	// 执行工具
	result, err := te.registry.Execute(ctx, call)
	if err != nil {
		return &Result{Error: fmt.Sprintf("Error: %v", err)}, err
	}

	return result, nil
}

// HandleToolCalls 处理多个工具调用
// 参数 calls: 工具调用列表，每个元素包含 Name 和 Arguments
func (te *ToolExecutor) HandleToolCalls(ctx context.Context, calls []map[string]interface{}) ([]*Result, error) {
	results := make([]*Result, 0, len(calls))
	for i, callData := range calls {
		name, _ := callData["name"].(string)
		args := callData["arguments"]
		
		result, err := te.ExecuteToolCall(ctx, name, args)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tool call '%d' (%s): %w", i, name, err)
		}
		results = append(results, result)
	}
	return results, nil
}

// GetAvailableTools 获取当前可用的工具列表
func (te *ToolExecutor) GetAvailableTools() []Definition {
	return te.registry.GetDefinitions()
}

// GetToolsJSON 获取工具定义的 JSON 表示（用于传递给 LLM）
func (te *ToolExecutor) GetToolsJSON() ([]byte, error) {
	definitions := te.registry.GetDefinitions()
	return json.Marshal(definitions)
}

