package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// GetCurrentTime 获取当前时间工具
func GetCurrentTime() *Tool {
	return &Tool{
		Definition: Definition{
			Name:        "get_current_time",
			Description: "获取当前的日期和时间",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "时间格式，例如: 2006-01-02 15:04:05",
						"enum":        []string{"datetime", "date", "time", "unix"},
					},
				},
				"required": []string{},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*Result, error) {
			format := "datetime"
			if f, ok := args["format"].(string); ok {
				format = f
			}

			now := time.Now()
			var content string

			switch format {
			case "date":
				content = now.Format("2006-01-02")
			case "time":
				content = now.Format("15:04:05")
			case "unix":
				content = fmt.Sprintf("%d", now.Unix())
			default:
				content = now.Format("2006-01-02 15:04:05")
			}

			return &Result{Content: content}, nil
		},
	}
}

// Calculate 计算器工具
func Calculate() *Tool {
	return &Tool{
		Definition: Definition{
			Name:        "calculate",
			Description: "执行数学计算，支持加减乘除、幂运算等",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "数学表达式，例如: 2+2, 10*5, 100/3",
					},
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "运算类型",
						"enum":        []string{"add", "subtract", "multiply", "divide", "power"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "第一个操作数",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "第二个操作数",
					},
				},
				"required": []string{},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*Result, error) {
			// 支持表达式计算
			if expr, ok := args["expression"].(string); ok {
				// 简单实现，实际应该使用更安全的表达式解析器
				result, err := evaluateExpression(expr)
				if err != nil {
					return &Result{Error: err.Error()}, err
				}
				return &Result{Content: fmt.Sprintf("%v", result)}, nil
			}

			// 支持参数计算
			operation, _ := args["operation"].(string)
			a, ok1 := args["a"].(float64)
			b, ok2 := args["b"].(float64)

			if !ok1 || !ok2 {
				return &Result{Error: "缺少必要的参数 a 和 b"}, nil
			}

			var result float64
			switch operation {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return &Result{Error: "除数不能为零"}, nil
				}
				result = a / b
			case "power":
				result = 1
				for i := 0; i < int(b); i++ {
					result *= a
				}
			default:
				return &Result{Error: fmt.Sprintf("不支持的运算类型: %s", operation)}, nil
			}

			return &Result{Content: fmt.Sprintf("%v", result)}, nil
		},
	}
}

// 简单的表达式求值（仅支持基础运算）
func evaluateExpression(expr string) (float64, error) {
	// 这里使用简单的解析，生产环境建议使用专门的表达式库
	// 为了演示，这里实现一个非常简单的解析器
	
	// 移除空格
	result := 0.0
	parts := make([]float64, 0)
	ops := make([]string, 0)
	
	current := ""
	for _, ch := range expr {
		if ch >= '0' && ch <= '9' || ch == '.' || ch == '-' {
			current += string(ch)
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			if current != "" {
				var val float64
				if _, err := fmt.Sscanf(current, "%f", &val); err != nil {
					return 0, fmt.Errorf("invalid number: %s", current)
				}
				parts = append(parts, val)
				current = ""
			}
			ops = append(ops, string(ch))
		}
	}
	
	if current != "" {
		var val float64
		if _, err := fmt.Sscanf(current, "%f", &val); err != nil {
			return 0, fmt.Errorf("invalid number: %s", current)
		}
		parts = append(parts, val)
	}
	
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	
	// 简单的从左到右计算（不考虑优先级）
	result = parts[0]
	for i, op := range ops {
		if i+1 >= len(parts) {
			break
		}
		switch op {
		case "+":
			result += parts[i+1]
		case "-":
			result -= parts[i+1]
		case "*":
			result *= parts[i+1]
		case "/":
			if parts[i+1] == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			result /= parts[i+1]
		}
	}
	
	return result, nil
}

// SearchInternet 搜索互联网工具（占位符）
func SearchInternet() *Tool {
	return &Tool{
		Definition: Definition{
			Name:        "search_internet",
			Description: "搜索互联网获取实时信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回结果数量",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*Result, error) {
			query, ok := args["query"].(string)
			if !ok {
				return &Result{Error: "缺少搜索关键词"}, nil
			}

			// TODO: 实现实际的搜索功能
			// 可以集成搜索引擎 API
			return &Result{
				Content: fmt.Sprintf("搜索 '%s' 的功能尚未实现，需要配置搜索 API", query),
				Error:   "not_implemented",
			}, nil
		},
	}
}

// GetWeather 获取天气信息工具（占位符）
func GetWeather() *Tool {
	return &Tool{
		Definition: Definition{
			Name:        "get_weather",
			Description: "获取指定城市的天气信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "城市名称",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"description": "温度单位",
						"enum":        []string{"celsius", "fahrenheit"},
						"default":     "celsius",
					},
				},
				"required": []string{"city"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*Result, error) {
			city, ok := args["city"].(string)
			if !ok {
				return &Result{Error: "缺少城市名称"}, nil
			}

			// TODO: 实现实际的天气查询功能
			// 可以集成天气 API
			return &Result{
				Content: fmt.Sprintf("查询 '%s' 天气的功能尚未实现，需要配置天气 API", city),
				Error:   "not_implemented",
			}, nil
		},
	}
}

// FormatJSON JSON 格式化工具
func FormatJSON() *Tool {
	return &Tool{
		Definition: Definition{
			Name:        "format_json",
			Description: "格式化 JSON 字符串，使其更易读",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"json_string": map[string]interface{}{
						"type":        "string",
						"description": "需要格式化的 JSON 字符串",
					},
					"indent": map[string]interface{}{
						"type":        "integer",
						"description": "缩进空格数",
						"default":     2,
					},
				},
				"required": []string{"json_string"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*Result, error) {
			jsonStr, ok := args["json_string"].(string)
			if !ok {
				return &Result{Error: "缺少 JSON 字符串"}, nil
			}

			indent := 2
			if ind, ok := args["indent"].(float64); ok {
				indent = int(ind)
			}

			var obj interface{}
			if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
				return &Result{Error: fmt.Sprintf("JSON 解析失败: %v", err)}, err
			}

			formatted, err := json.MarshalIndent(obj, "", fmt.Sprintf("%d", indent))
			if err != nil {
				return &Result{Error: fmt.Sprintf("JSON 格式化失败: %v", err)}, err
			}

			return &Result{Content: string(formatted)}, nil
		},
	}
}
