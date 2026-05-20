package tool

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]*Tool
}

// NewToolRegistry 创建新的工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*Tool),
	}
}

// Register 注册一个工具
func (r *ToolRegistry) Register(tool *Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	if tool.Definition.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if r.tools[tool.Definition.Name] != nil {
		return fmt.Errorf("tool '%s' already registered", tool.Definition.Name)
	}
	r.tools[tool.Definition.Name] = tool
	return nil
}

// Get 获取已注册的工具
func (r *ToolRegistry) Get(name string) (*Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetAll 获取所有已注册的工具
func (r *ToolRegistry) GetAll() []*Tool {
	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetDefinitions 获取所有工具的定义（用于注册到 LLM）
func (r *ToolRegistry) GetDefinitions() []Definition {
	definitions := make([]Definition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, tool.Definition)
	}
	return definitions
}

// Execute 执行工具调用
func (r *ToolRegistry) Execute(ctx context.Context, call *Call) (*Result, error) {
	tool, ok := r.Get(call.Name)
	if !ok {
		return &Result{Error: fmt.Sprintf("tool '%s' not found", call.Name)}, fmt.Errorf("tool '%s' not found", call.Name)
	}
	if tool.Handler == nil {
		return &Result{Error: fmt.Sprintf("tool '%s' has no handler", call.Name)}, fmt.Errorf("tool '%s' has no handler", call.Name)
	}
	return tool.Handler(ctx, call.Arguments)
}

// LoadBuiltinTools 加载内置工具
func (r *ToolRegistry) LoadBuiltinTools() error {
	builtinTools := []*Tool{
		GetCurrentTime(),
		Calculate(),
		SearchInternet(),
		GetWeather(),
		FormatJSON(),
	}

	for _, tool := range builtinTools {
		if err := r.Register(tool); err != nil {
			return fmt.Errorf("failed to register builtin tool '%s': %w", tool.Definition.Name, err)
		}
	}

	return nil
}

// LoadFromYAML 从 YAML 文件加载工具配置
func (r *ToolRegistry) LoadFromYAML(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read tools config file: %w", err)
	}

	var config ToolsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse tools config: %w", err)
	}

	// 根据配置启用/禁用工具
	for _, toolConfig := range config.Tools {
		if tool, ok := r.Get(toolConfig.Name); ok {
			tool.Definition.Description = toolConfig.Description
			if toolConfig.Enabled != nil && !*toolConfig.Enabled {
				// 如果禁用，从注册表中移除
				delete(r.tools, toolConfig.Name)
			}
		}
	}

	return nil
}

// ToolsConfig 工具配置文件结构
type ToolsConfig struct {
	Tools []ToolConfig `yaml:"tools"`
}

// ToolConfig 单个工具配置
type ToolConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Enabled     *bool  `yaml:"enabled,omitempty"`
}
