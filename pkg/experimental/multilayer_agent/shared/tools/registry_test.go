package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 用于测试的模拟工具
type mockTool struct {
	name        string
	description string
	callResult  string
	callError   error
}

func (t *mockTool) Name() string {
	return t.name
}

func (t *mockTool) Description() string {
	return t.description
}

func (t *mockTool) Call(ctx context.Context, input string) (string, error) {
	return t.callResult, t.callError
}

// TestNewToolRegistry 测试创建新的工具注册表
func TestNewToolRegistry(t *testing.T) {
	// 创建新的工具注册表
	registry := NewToolRegistry()
	
	// 验证初始状态
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.tools)
	assert.Empty(t, registry.tools)
	assert.Empty(t, registry.ListTools())
}

// TestRegisterTool 测试工具注册功能
func TestRegisterTool(t *testing.T) {
	// 创建注册表
	registry := NewToolRegistry()
	
	// 创建模拟工具
	tool1 := &mockTool{name: "测试工具1", description: "测试描述1"}
	tool2 := &mockTool{name: "测试工具2", description: "测试描述2"}
	
	// 测试成功注册
	t.Run("成功注册工具", func(t *testing.T) {
		err := registry.RegisterTool(tool1)
		assert.NoError(t, err)
		
		// 验证工具已注册
		registeredTool, err := registry.GetTool("测试工具1")
		assert.NoError(t, err)
		assert.Equal(t, tool1, registeredTool)
	})
	
	// 测试注册重名工具
	t.Run("注册重名工具应返回错误", func(t *testing.T) {
		// 尝试注册同名工具
		duplicateTool := &mockTool{name: "测试工具1", description: "不同描述"}
		err := registry.RegisterTool(duplicateTool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已存在")
		
		// 原工具应保持不变
		registeredTool, err := registry.GetTool("测试工具1")
		assert.NoError(t, err)
		assert.Equal(t, tool1, registeredTool)
	})
	
	// 测试注册多个工具
	t.Run("可以注册多个不同名称的工具", func(t *testing.T) {
		err := registry.RegisterTool(tool2)
		assert.NoError(t, err)
		
		// 验证两个工具都已注册
		tool1Registered, err := registry.GetTool("测试工具1")
		assert.NoError(t, err)
		assert.Equal(t, tool1, tool1Registered)
		
		tool2Registered, err := registry.GetTool("测试工具2")
		assert.NoError(t, err)
		assert.Equal(t, tool2, tool2Registered)
	})
}

// TestGetTool 测试获取工具功能
func TestGetTool(t *testing.T) {
	// 创建注册表并注册工具
	registry := NewToolRegistry()
	tool := &mockTool{name: "测试工具", description: "测试描述"}
	_ = registry.RegisterTool(tool)
	
	// 测试获取存在的工具
	t.Run("获取存在的工具应成功", func(t *testing.T) {
		registeredTool, err := registry.GetTool("测试工具")
		assert.NoError(t, err)
		assert.Equal(t, tool, registeredTool)
	})
	
	// 测试获取不存在的工具
	t.Run("获取不存在的工具应返回错误", func(t *testing.T) {
		_, err := registry.GetTool("不存在的工具")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不存在")
	})
}

// TestListTools 测试列出所有工具功能
func TestListTools(t *testing.T) {
	// 创建空注册表
	registry := NewToolRegistry()
	
	// 测试空注册表
	t.Run("空注册表应返回空列表", func(t *testing.T) {
		tools := registry.ListTools()
		assert.Empty(t, tools)
	})
	
	// 注册多个工具
	tool1 := &mockTool{name: "工具1"}
	tool2 := &mockTool{name: "工具2"}
	tool3 := &mockTool{name: "工具3"}
	_ = registry.RegisterTool(tool1)
	_ = registry.RegisterTool(tool2)
	_ = registry.RegisterTool(tool3)
	
	// 测试列出所有工具
	t.Run("应列出所有已注册工具", func(t *testing.T) {
		tools := registry.ListTools()
		
		// 验证返回长度
		assert.Equal(t, 3, len(tools))
		
		// 验证所有工具都在列表中
		// 注意：由于map遍历顺序不固定，这里只检查数量和包含关系
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name()] = true
		}
		
		assert.True(t, toolNames["工具1"])
		assert.True(t, toolNames["工具2"])
		assert.True(t, toolNames["工具3"])
	})
}
