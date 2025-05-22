package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewToolCaller 测试创建工具调用器
func TestNewToolCaller(t *testing.T) {
	// 创建工具注册表
	registry := NewToolRegistry()
	
	// 创建工具调用器
	caller := NewToolCaller(registry)
	
	// 验证工具调用器
	assert.NotNil(t, caller)
	assert.Equal(t, registry, caller.registry)
}

// TestCallTool 测试调用工具功能
func TestCallTool(t *testing.T) {
	// 创建上下文
	ctx := context.Background()
	
	// 测试调用不存在的工具
	t.Run("调用不存在的工具应返回错误响应", func(t *testing.T) {
		// 创建空注册表和调用器
		registry := NewToolRegistry()
		caller := NewToolCaller(registry)
		
		// 创建调用请求
		req := ToolRequest{
			ToolName: "不存在的工具",
			Input:    json.RawMessage(`"测试输入"`),
		}
		
		// 调用工具
		resp, err := caller.CallTool(ctx, req)
		
		// 验证结果
		assert.NoError(t, err, "CallTool不应返回处理错误")
		assert.NotNil(t, resp)
		assert.Equal(t, "不存在的工具", resp.ToolName)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error, "工具不存在")
		assert.Empty(t, resp.Result)
	})
	
	// 测试调用成功的工具
	t.Run("调用成功的工具应返回成功响应", func(t *testing.T) {
		// 创建注册表和工具
		registry := NewToolRegistry()
		successTool := &mockTool{
			name:        "成功工具",
			description: "总是成功的工具",
			callResult:  "调用成功的结果",
			callError:   nil,
		}
		
		// 注册工具
		_ = registry.RegisterTool(successTool)
		
		// 创建调用器
		caller := NewToolCaller(registry)
		
		// 创建调用请求
		req := ToolRequest{
			ToolName: "成功工具",
			Input:    json.RawMessage(`"测试输入"`),
		}
		
		// 调用工具
		resp, err := caller.CallTool(ctx, req)
		
		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "成功工具", resp.ToolName)
		assert.True(t, resp.Success)
		assert.Empty(t, resp.Error)
		assert.Equal(t, "调用成功的结果", resp.Result)
	})
	
	// 测试调用失败的工具
	t.Run("调用失败的工具应返回失败响应", func(t *testing.T) {
		// 创建注册表和工具
		registry := NewToolRegistry()
		failTool := &mockTool{
			name:        "失败工具",
			description: "总是失败的工具",
			callResult:  "",
			callError:   errors.New("模拟的工具错误"),
		}
		
		// 注册工具
		_ = registry.RegisterTool(failTool)
		
		// 创建调用器
		caller := NewToolCaller(registry)
		
		// 创建调用请求
		req := ToolRequest{
			ToolName: "失败工具",
			Input:    json.RawMessage(`"测试输入"`),
		}
		
		// 调用工具
		resp, err := caller.CallTool(ctx, req)
		
		// 验证结果
		assert.NoError(t, err, "即使工具调用失败，CallTool也不应返回处理错误")
		assert.NotNil(t, resp)
		assert.Equal(t, "失败工具", resp.ToolName)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error, "工具调用失败")
		assert.Contains(t, resp.Error, "模拟的工具错误")
		assert.Empty(t, resp.Result)
	})
	
	// 测试不同类型的输入参数
	t.Run("应正确处理不同类型的输入参数", func(t *testing.T) {
		// 测试用例
		testCases := []struct {
			name  string
			input json.RawMessage
		}{
			{"字符串输入", json.RawMessage(`"字符串参数"`)},
			{"数字输入", json.RawMessage(`123`)},
			{"布尔输入", json.RawMessage(`true`)},
			{"对象输入", json.RawMessage(`{"key": "value"}`)},
			{"数组输入", json.RawMessage(`[1, 2, 3]`)},
			{"空输入", nil},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 创建注册表和工具
				registry := NewToolRegistry()
				echoTool := &mockTool{
					name:        "回显工具",
					description: "回显输入参数的工具",
					callResult:  "回显工具被调用",
					callError:   nil,
				}
				
				// 注册工具
				_ = registry.RegisterTool(echoTool)
				
				// 创建调用器
				caller := NewToolCaller(registry)
				
				// 创建调用请求
				req := ToolRequest{
					ToolName: "回显工具",
					Input:    tc.input,
				}
				
				// 调用工具
				resp, err := caller.CallTool(ctx, req)
				
				// 验证结果
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
				assert.Equal(t, "回显工具被调用", resp.Result)
			})
		}
	})
}

// TestCallToolFromJSON 测试从JSON字符串调用工具
func TestCallToolFromJSON(t *testing.T) {
	// 创建上下文
	ctx := context.Background()
	
	// 测试有效JSON
	t.Run("有效的JSON请求应被正确处理", func(t *testing.T) {
		// 创建注册表和工具
		registry := NewToolRegistry()
		successTool := &mockTool{
			name:        "JSON测试工具",
			description: "测试JSON调用的工具",
			callResult:  "JSON调用成功",
			callError:   nil,
		}
		
		// 注册工具
		_ = registry.RegisterTool(successTool)
		
		// 创建调用器
		caller := NewToolCaller(registry)
		
		// 创建JSON请求
		jsonRequest := `{
			"tool_name": "JSON测试工具",
			"input": "JSON测试输入"
		}`
		
		// 调用工具
		resp, err := caller.CallToolFromJSON(ctx, jsonRequest)
		
		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "JSON测试工具", resp.ToolName)
		assert.True(t, resp.Success)
		assert.Equal(t, "JSON调用成功", resp.Result)
	})
	
	// 测试无效JSON
	t.Run("无效的JSON请求应返回错误", func(t *testing.T) {
		// 创建调用器
		registry := NewToolRegistry()
		caller := NewToolCaller(registry)
		
		// 创建无效的JSON请求
		invalidJSON := `{这不是有效的JSON`
		
		// 调用工具
		resp, err := caller.CallToolFromJSON(ctx, invalidJSON)
		
		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "无效的工具调用JSON")
	})
	
	// 测试缺少必要字段的JSON
	t.Run("缺少必要字段的JSON应能处理", func(t *testing.T) {
		// 创建调用器
		registry := NewToolRegistry()
		caller := NewToolCaller(registry)
		
		// 创建缺少tool_name字段的JSON请求
		missingFieldJSON := `{
			"input": "测试输入"
		}`
		
		// 调用工具
		resp, err := caller.CallToolFromJSON(ctx, missingFieldJSON)
		
		// 即使缺少字段，JSON解析也应成功，但工具调用会失败
		assert.NoError(t, err, "JSON解析应成功")
		assert.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error, "工具不存在")
	})
}
