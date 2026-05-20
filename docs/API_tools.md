# Tools 系统使用文档

本文档说明如何在 trading-forum-go 项目中使用 AI Agent 工具系统。

## 目录结构

```
internal/tool/
├── tool_entity.go      # 工具相关的数据结构定义
├── tool_init.go        # 工具注册表和管理逻辑
├── builtin_tools.go    # 内置工具实现
├── tool_executor.go    # 工具执行器，用于与 Eino 框架集成
└── init.go             # 全局初始化和访问接口

config/
└── tools.yaml          # 工具配置文件
```

## 快速开始

### 1. 在应用启动时初始化工具系统

在 `main.go` 或应用初始化代码中：

```go
import (
    "blog.alphazer01214.top/internal/tool"
)

func main() {
    // 其他初始化代码...
    
    // 初始化工具系统
    if err := tool.InitTools("./config/tools.yaml"); err != nil {
        log.Printf("Failed to init tools: %v", err)
    }
    
    // 继续其他初始化...
}
```

### 2. 在 AI Service 中使用工具

修改 `internal/service/ai_service.go`，在调用 LLM 时集成工具：

```go
import (
    "blog.alphazer01214.top/internal/tool"
    "github.com/cloudwego/eino/schema"
)

func (ai *AIService) InvokeAgentWithTools(ctx context.Context, userId uint, agentId uint, req *request.InvokeAgentRequest) (*response.StandardAiResponse, error) {
    // 获取 agent 和 chatModel...
    
    // 获取工具执行器
    executor := tool.GetGlobalExecutor()
    
    // 构建消息
    messages := []*schema.Message{
        schema.UserMessage(req.UsrPrompt),
    }
    
    // 调用 LLM（需要支持工具调用）
    // 注意：具体实现取决于 Eino 版本的 API
    result, err := chatModel.Generate(ctx, messages)
    if err != nil {
        return nil, err
    }
    
    // 检查是否有工具调用
    if len(result.ToolCalls) > 0 {
        // 执行工具调用
        toolResults, err := executor.HandleToolCalls(ctx, result.ToolCalls)
        if err != nil {
            return nil, err
        }
        
        // 将工具结果添加到消息中
        messages = append(messages, result)
        messages = append(messages, toolResults...)
        
        // 再次调用 LLM 获取最终结果
        finalResult, err := chatModel.Generate(ctx, messages)
        if err != nil {
            return nil, err
        }
        
        return &response.StandardAiResponse{
            Content: finalResult.Content,
            Status:  true,
        }, nil
    }
    
    return &response.StandardAiResponse{
        Content: result.Content,
        Status:  true,
    }, nil
}
```

## 内置工具

### 1. get_current_time - 获取当前时间

获取当前的日期和时间，支持多种格式。

**参数：**
- `format` (可选): 时间格式
  - `datetime` (默认): "2006-01-02 15:04:05"
  - `date`: "2006-01-02"
  - `time`: "15:04:05"
  - `unix`: Unix 时间戳

**示例调用：**
```json
{
  "name": "get_current_time",
  "arguments": {
    "format": "datetime"
  }
}
```

### 2. calculate - 数学计算

执行数学计算，支持加减乘除、幂运算等。

**参数方式 1 - 表达式：**
- `expression`: 数学表达式字符串，如 "2+2", "10*5"

**参数方式 2 - 结构化：**
- `operation`: 运算类型 (add, subtract, multiply, divide, power)
- `a`: 第一个操作数
- `b`: 第二个操作数

**示例调用：**
```json
{
  "name": "calculate",
  "arguments": {
    "expression": "100 + 50"
  }
}
```

或

```json
{
  "name": "calculate",
  "arguments": {
    "operation": "multiply",
    "a": 10,
    "b": 5
  }
}
```

### 3. format_json - JSON 格式化

格式化 JSON 字符串，使其更易读。

**参数：**
- `json_string` (必需): 需要格式化的 JSON 字符串
- `indent` (可选): 缩进空格数，默认 2

**示例调用：**
```json
{
  "name": "format_json",
  "arguments": {
    "json_string": "{\"name\":\"John\",\"age\":30}",
    "indent": 2
  }
}
```

### 4. search_internet - 互联网搜索（占位符）

搜索互联网获取实时信息。

**状态：** 默认禁用，需要实现实际的搜索 API 集成。

### 5. get_weather - 天气查询（占位符）

获取指定城市的天气信息。

**状态：** 默认禁用，需要实现实际的天气 API 集成。

## 配置文件说明

`config/tools.yaml` 用于控制哪些工具可用：

```yaml
tools:
  - name: "get_current_time"
    description: "自定义描述（可选）"
    enabled: true  # true 启用，false 禁用
  
  - name: "calculate"
    enabled: true
  
  - name: "search_internet"
    enabled: false  # 禁用
```

## 自定义工具

### 创建新工具

1. 在 `builtin_tools.go` 或新文件中定义工具：

```go
func MyCustomTool() *tool.Tool {
    return &tool.Tool{
        Definition: tool.Definition{
            Name:        "my_tool",
            Description: "我的自定义工具",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "param1": map[string]interface{}{
                        "type":        "string",
                        "description": "参数1的描述",
                    },
                },
                "required": []string{"param1"},
            },
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (*tool.Result, error) {
            // 实现工具逻辑
            param1 := args["param1"].(string)
            
            return &tool.Result{
                Content: fmt.Sprintf("Result: %s", param1),
            }, nil
        },
    }
}
```

2. 在 `LoadBuiltinTools` 中注册：

```go
func (r *ToolRegistry) LoadBuiltinTools() error {
    builtinTools := []*Tool{
        GetCurrentTime(),
        Calculate(),
        SearchInternet(),
        GetWeather(),
        FormatJSON(),
        MyCustomTool(),  // 添加新工具
    }
    // ...
}
```

3. 在 `tools.yaml` 中配置（可选）：

```yaml
tools:
  - name: "my_tool"
    enabled: true
```

## API 集成示例

### 在 Agent 创建时保存工具配置

Agent 实体已经包含 `Tools` 字段（`datatypes.JSONMap`），可以在创建/更新 Agent 时保存工具配置。

### 查询可用工具

```go
// 获取所有可用工具的定义
definitions := tool.GetGlobalRegistry().GetDefinitions()

// 转换为 JSON 返回给前端
toolsJSON, _ := json.Marshal(definitions)
```

## 注意事项

1. **工具安全**：工具执行时应该进行适当的权限检查和参数验证
2. **错误处理**：工具应该返回有意义的错误信息，帮助 LLM 理解问题
3. **超时控制**：长时间运行的工具应该支持超时和取消
4. **并发安全**：工具注册表和执行器都是并发安全的
5. **性能考虑**：避免在工具中执行过于耗时的操作

## 扩展建议

1. **数据库查询工具**：允许 agent 查询数据库（需要严格控制权限）
2. **文件操作工具**：读取/写入文件（需要沙箱环境）
3. **HTTP 请求工具**：发送 HTTP 请求（需要白名单限制）
4. **代码执行工具**：执行代码片段（需要沙箱环境）
5. **第三方 API 集成**：集成各种 SaaS 服务

## 故障排查

### 工具未找到

检查：
1. 工具是否在 `LoadBuiltinTools` 中注册
2. `tools.yaml` 中是否设置为 `enabled: true`
3. 是否在应用启动时调用了 `InitTools`

### 工具执行失败

检查：
1. 参数是否正确传递
2. Handler 函数中的逻辑是否有错误
3. 查看日志中的错误信息

## 更新日志

- 2026-05-20: 初始版本，包含 5 个基础工具
