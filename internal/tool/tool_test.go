package tool

import (
	"context"
	"testing"
)

func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	// 测试加载内置工具
	err := registry.LoadBuiltinTools()
	if err != nil {
		t.Fatalf("Failed to load builtin tools: %v", err)
	}

	// 验证工具数量
	allTools := registry.GetAll()
	if len(allTools) != 5 {
		t.Errorf("Expected 5 tools, got %d", len(allTools))
	}

	// 测试获取特定工具
	timeTool, ok := registry.Get("get_current_time")
	if !ok {
		t.Fatal("Failed to get get_current_time tool")
	}
	if timeTool.Definition.Name != "get_current_time" {
		t.Errorf("Expected tool name 'get_current_time', got '%s'", timeTool.Definition.Name)
	}
}

func TestGetCurrentTime(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(GetCurrentTime())

	ctx := context.Background()
	call := &Call{
		Name:      "get_current_time",
		Arguments: map[string]interface{}{"format": "datetime"},
	}

	result, err := registry.Execute(ctx, call)
	if err != nil {
		t.Fatalf("Failed to execute get_current_time: %v", err)
	}

	if result.Content == "" {
		t.Error("Expected non-empty time content")
	}

	t.Logf("Current time: %s", result.Content)
}

func TestCalculate(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(Calculate())

	ctx := context.Background()

	// 测试加法
	call := &Call{
		Name: "calculate",
		Arguments: map[string]interface{}{
			"operation": "add",
			"a":         float64(10),
			"b":         float64(5),
		},
	}

	result, err := registry.Execute(ctx, call)
	if err != nil {
		t.Fatalf("Failed to execute calculate: %v", err)
	}

	if result.Content != "15" {
		t.Errorf("Expected '15', got '%s'", result.Content)
	}

	// 测试除法
	call = &Call{
		Name: "calculate",
		Arguments: map[string]interface{}{
			"operation": "divide",
			"a":         float64(100),
			"b":         float64(4),
		},
	}

	result, err = registry.Execute(ctx, call)
	if err != nil {
		t.Fatalf("Failed to execute calculate divide: %v", err)
	}

	if result.Content != "25" {
		t.Errorf("Expected '25', got '%s'", result.Content)
	}
}

func TestFormatJSON(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(FormatJSON())

	ctx := context.Background()
	call := &Call{
		Name: "format_json",
		Arguments: map[string]interface{}{
			"json_string": `{"name":"John","age":30,"city":"New York"}`,
			"indent":      float64(2),
		},
	}

	result, err := registry.Execute(ctx, call)
	if err != nil {
		t.Fatalf("Failed to execute format_json: %v", err)
	}

	expected := `{
  "age": 30,
  "city": "New York",
  "name": "John"
}`

	if result.Content != expected {
		t.Errorf("Expected formatted JSON:\n%s\nGot:\n%s", expected, result.Content)
	}

	t.Logf("Formatted JSON:\n%s", result.Content)
}

func TestGlobalInit(t *testing.T) {
	// 清理全局状态
	GlobalRegistry = nil
	GlobalExecutor = nil

	// 测试初始化
	err := InitTools("")
	if err != nil {
		t.Fatalf("Failed to init tools: %v", err)
	}

	// 验证全局实例
	if GlobalRegistry == nil {
		t.Fatal("GlobalRegistry is nil")
	}
	if GlobalExecutor == nil {
		t.Fatal("GlobalExecutor is nil")
	}

	// 测试获取全局实例
	registry := GetGlobalRegistry()
	if registry == nil {
		t.Fatal("GetGlobalRegistry returned nil")
	}

	executor := GetGlobalExecutor()
	if executor == nil {
		t.Fatal("GetGlobalExecutor returned nil")
	}

	tools := registry.GetAll()
	if len(tools) == 0 {
		t.Error("Expected some tools to be registered")
	}

	t.Logf("Initialized %d tools", len(tools))
}

func TestToolDefinitions(t *testing.T) {
	registry := NewToolRegistry()
	registry.LoadBuiltinTools()

	definitions := registry.GetDefinitions()

	if len(definitions) != 5 {
		t.Errorf("Expected 5 definitions, got %d", len(definitions))
	}

	for _, def := range definitions {
		if def.Name == "" {
			t.Error("Tool definition has empty name")
		}
		if def.Description == "" {
			t.Errorf("Tool '%s' has empty description", def.Name)
		}
		if def.Parameters == nil {
			t.Errorf("Tool '%s' has nil parameters", def.Name)
		}

		t.Logf("Tool: %s - %s", def.Name, def.Description)
	}
}

func TestExecuteNonExistentTool(t *testing.T) {
	registry := NewToolRegistry()

	ctx := context.Background()
	call := &Call{
		Name:      "non_existent_tool",
		Arguments: map[string]interface{}{},
	}

	result, err := registry.Execute(ctx, call)
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}

	if result == nil {
		t.Fatal("Expected result even for error case")
	}

	if result.Error == "" {
		t.Error("Expected error message in result")
	}

	t.Logf("Got expected error: %s", result.Error)
}
