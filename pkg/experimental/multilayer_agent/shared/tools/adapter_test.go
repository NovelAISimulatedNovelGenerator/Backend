package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLangChainAdapter 测试创建新的LangChain工具适配器
func TestNewLangChainAdapter(t *testing.T) {
	// 创建模拟工具
	mockTool := &mockTool{
		name:        "测试工具",
		description: "测试工具描述",
	}
	
	// 创建适配器
	adapter := NewLangChainAdapter(mockTool)
	
	// 验证适配器属性
	assert.NotNil(t, adapter)
	assert.Equal(t, mockTool, adapter.originalTool)
}

// TestLangChainAdapterName 测试适配器的Name方法
func TestLangChainAdapterName(t *testing.T) {
	// 创建模拟工具
	mockTool := &mockTool{
		name: "测试工具名称",
	}
	
	// 创建适配器
	adapter := NewLangChainAdapter(mockTool)
	
	// 验证Name方法
	assert.Equal(t, "测试工具名称", adapter.Name())
}

// TestLangChainAdapterDescription 测试适配器的Description方法
func TestLangChainAdapterDescription(t *testing.T) {
	// 创建模拟工具
	mockTool := &mockTool{
		description: "这是一个测试工具的详细描述",
	}
	
	// 创建适配器
	adapter := NewLangChainAdapter(mockTool)
	
	// 验证Description方法
	assert.Equal(t, "这是一个测试工具的详细描述", adapter.Description())
}

// TestLangChainAdapterCall 测试适配器的Call方法
func TestLangChainAdapterCall(t *testing.T) {
	// 测试成功调用
	t.Run("成功调用工具", func(t *testing.T) {
		// 创建模拟工具
		mockTool := &mockTool{
			name:        "成功工具",
			description: "总是成功的工具",
			callResult:  "调用成功的结果",
			callError:   nil,
		}
		
		// 创建适配器
		adapter := NewLangChainAdapter(mockTool)
		
		// 调用工具
		result, err := adapter.Call(context.Background(), "测试输入")
		
		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "调用成功的结果", result)
	})
	
	// 测试失败调用
	t.Run("调用失败的工具", func(t *testing.T) {
		// 创建模拟工具
		expectedError := errors.New("模拟的工具错误")
		mockTool := &mockTool{
			name:        "失败工具",
			description: "总是失败的工具",
			callResult:  "",
			callError:   expectedError,
		}
		
		// 创建适配器
		adapter := NewLangChainAdapter(mockTool)
		
		// 调用工具
		result, err := adapter.Call(context.Background(), "测试输入")
		
		// 验证结果
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, result)
	})
	
	// 测试空输入
	t.Run("处理空输入", func(t *testing.T) {
		// 创建模拟工具
		mockTool := &mockTool{
			name:        "空输入工具",
			description: "处理空输入的工具",
			callResult:  "处理空输入的结果",
			callError:   nil,
		}
		
		// 创建适配器
		adapter := NewLangChainAdapter(mockTool)
		
		// 调用工具，使用空输入
		result, err := adapter.Call(context.Background(), "")
		
		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "处理空输入的结果", result)
	})
}
