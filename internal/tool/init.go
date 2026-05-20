package tool

import (
	"fmt"
)

// GlobalRegistry 全局工具注册表实例
var GlobalRegistry *ToolRegistry

// GlobalExecutor 全局工具执行器实例
var GlobalExecutor *ToolExecutor

// InitTools 初始化工具系统
// 在应用启动时调用一次
func InitTools(configPath string) error {
	// 创建工具注册表
	registry := NewToolRegistry()

	// 加载内置工具
	if err := registry.LoadBuiltinTools(); err != nil {
		return fmt.Errorf("failed to load builtin tools: %w", err)
	}

	// 从配置文件加载工具配置
	if configPath != "" {
		if err := registry.LoadFromYAML(configPath); err != nil {
			// 配置文件加载失败不阻断启动，只记录警告
			fmt.Printf("[WARNING] failed to load tools config: %v\n", err)
		}
	}

	// 创建工具执行器
	executor := NewToolExecutor(registry)

	// 保存到全局变量
	GlobalRegistry = registry
	GlobalExecutor = executor

	fmt.Printf("[INFO] initialized %d tools\n", len(registry.GetAll()))
	return nil
}

// GetGlobalRegistry 获取全局工具注册表
func GetGlobalRegistry() *ToolRegistry {
	if GlobalRegistry == nil {
		panic("tools not initialized, call InitTools first")
	}
	return GlobalRegistry
}

// GetGlobalExecutor 获取全局工具执行器
func GetGlobalExecutor() *ToolExecutor {
	if GlobalExecutor == nil {
		panic("tools not initialized, call InitTools first")
	}
	return GlobalExecutor
}
