// Package tools 实现多层代理系统的工具调用功能
// 这个包提供工具注册、调用和管理的核心功能
package tools

import (
	"fmt"
	"sync"
	
	"github.com/tmc/langchaingo/tools"
)

// ToolRegistry 管理系统中所有可用的工具
// 提供注册、获取和列举功能，确保线程安全
type ToolRegistry struct {
	// tools 存储所有已注册的工具，以工具名称为键
	tools map[string]tools.Tool
	// mu 用于保护 tools 映射的并发访问
	mu    sync.RWMutex
}

// NewToolRegistry 创建新的工具注册表实例
// 返回一个初始化好的 ToolRegistry 指针
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]tools.Tool),
	}
}

// RegisterTool 注册一个工具到注册表
// 参数:
//   - tool: 要注册的工具，必须实现 tools.Tool 接口
// 返回:
//   - error: 如果同名工具已存在，将返回错误
func (r *ToolRegistry) RegisterTool(tool tools.Tool) error {
	// 加锁确保线程安全
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 获取工具名称
	name := tool.Name()
	
	// 检查工具是否已存在
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("工具 %s 已存在", name)
	}
	
	// 注册新工具
	r.tools[name] = tool
	return nil
}

// GetTool 根据名称获取工具
// 参数:
//   - name: 工具名称
// 返回:
//   - tools.Tool: 找到的工具
//   - error: 如果工具不存在，将返回错误
func (r *ToolRegistry) GetTool(name string) (tools.Tool, error) {
	// 使用读锁，允许并发读取
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// 查找工具
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("工具 %s 不存在", name)
	}
	
	return tool, nil
}

// ListTools 列出所有已注册的工具
// 返回:
//   - []tools.Tool: 所有已注册工具的切片
func (r *ToolRegistry) ListTools() []tools.Tool {
	// 使用读锁，允许并发读取
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// 创建切片存储所有工具
	list := make([]tools.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		list = append(list, tool)
	}
	
	return list
}
