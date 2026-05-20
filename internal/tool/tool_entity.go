package tool

import (
	"context"
)

// Definition 工具定义，用于注册到 LLM
type Definition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Call 工具调用请求
type Call struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Result 工具调用结果
type Result struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// Handler 工具处理函数类型
type Handler func(ctx context.Context, args map[string]interface{}) (*Result, error)

// Tool 完整的工具定义
type Tool struct {
	Definition Definition `json:"definition"`
	Handler    Handler    `json:"-"` // 不序列化到 JSON
}
