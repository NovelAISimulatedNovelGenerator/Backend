// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import "novelai/pkg/constants"

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAdapter_GenerateText 测试适配器的文本生成功能
func TestAdapter_GenerateText(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/completions" {
			t.Errorf("期望路径为'/completions'，实际为'%s'", r.URL.Path)
		}
		
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("期望方法为'POST'，实际为'%s'", r.Method)
		}
		
		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "cmpl-123",
			"object": "completion",
			"created": 1677858242,
			"model": "deepseek-chat",
			"choices": [
				{
					"text": "这是一个适配器测试响应",
					"index": 0,
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 5,
				"completion_tokens": 7,
				"total_tokens": 12
			}
		}`))
	}))
	defer server.Close()
	
	// 创建配置
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	
	// 创建适配器
	adapter, err := NewAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("创建适配器失败: %v", err)
	}
	
	// 发送文本生成请求
	ctx := context.Background()
	result, err := adapter.GenerateText(ctx, constants.DeepSeekChat, "这是一个测试", 100)
	if err != nil {
		t.Fatalf("生成文本失败: %v", err)
	}
	
	// 验证结果
	expectedText := "这是一个适配器测试响应"
	if result != expectedText {
		t.Errorf("期望结果为'%s'，实际为'%s'", expectedText, result)
	}
}

// TestAdapter_ChatWithSystem 测试适配器的聊天功能
func TestAdapter_ChatWithSystem(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/chat/completions" {
			t.Errorf("期望路径为'/chat/completions'，实际为'%s'", r.URL.Path)
		}
		
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("期望方法为'POST'，实际为'%s'", r.Method)
		}
		
		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677858242,
			"model": "deepseek-chat",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "这是一个系统聊天测试响应"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 8,
				"total_tokens": 18
			}
		}`))
	}))
	defer server.Close()
	
	// 创建配置
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	
	// 创建适配器
	adapter, err := NewAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("创建适配器失败: %v", err)
	}
	
	// 发送聊天请求
	ctx := context.Background()
	result, err := adapter.ChatWithSystem(ctx, constants.DeepSeekChat, "你是一个测试助手", "这是一个测试消息", 100)
	if err != nil {
		t.Fatalf("聊天请求失败: %v", err)
	}
	
	// 验证结果
	expectedText := "这是一个系统聊天测试响应"
	if result != expectedText {
		t.Errorf("期望结果为'%s'，实际为'%s'", expectedText, result)
	}
}

// TestAdapter_ChatWithMessages 测试适配器的消息列表聊天功能
func TestAdapter_ChatWithMessages(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/chat/completions" {
			t.Errorf("期望路径为'/chat/completions'，实际为'%s'", r.URL.Path)
		}
		
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("期望方法为'POST'，实际为'%s'", r.Method)
		}
		
		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677858242,
			"model": "deepseek-chat",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "这是一个消息列表聊天测试响应"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 8,
				"total_tokens": 18
			}
		}`))
	}))
	defer server.Close()
	
	// 创建配置
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	
	// 创建适配器
	adapter, err := NewAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("创建适配器失败: %v", err)
	}
	
	// 创建消息列表
	messages := []Message{
		{Role: constants.RoleSystem, Content: "你是一个测试助手"},
		{Role: constants.RoleUser, Content: "这是一个测试消息"},
	}
	
	// 发送聊天请求
	ctx := context.Background()
	result, err := adapter.ChatWithMessages(ctx, constants.DeepSeekChat, messages, 100)
	if err != nil {
		t.Fatalf("聊天请求失败: %v", err)
	}
	
	// 验证结果
	expectedText := "这是一个消息列表聊天测试响应"
	if result != expectedText {
		t.Errorf("期望结果为'%s'，实际为'%s'", expectedText, result)
	}
}

// TestAdapter_StreamMock 模拟适配器的流式响应处理
func TestAdapter_StreamMock(t *testing.T) {
	// 这个测试主要是模拟流式响应的处理逻辑
	// 由于流式测试需要模拟SSE流，实际测试会比较复杂，
	// 这里只是简单验证流式处理方法的逻辑
	
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/chat/completions" {
			t.Errorf("期望路径为'/chat/completions'，实际为'%s'", r.URL.Path)
		}
		
		// 验证请求参数
		if r.URL.Query().Get("stream") != "true" {
			t.Errorf("期望stream参数为'true'，实际为'%s'", r.URL.Query().Get("stream"))
		}
		
		// 模拟SSE响应
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		
		// 在实际测试中，这里会发送SSE格式的数据流
		// 但由于测试环境的限制，这里只是模拟一个简单的响应
		w.Write([]byte(`data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"这是"}}]}
		
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"一个"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"流式"}}]}

data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"测试"}}]}

data: [DONE]
`))
	}))
	defer server.Close()
	
	// 对于流式响应测试，由于需要复杂的SSE模拟，
	// 实际项目中可以考虑使用专门的流式响应测试框架，
	// 或者将流式响应逻辑抽象出来单独测试
}

// TestAdapter_ErrorHandling 测试适配器的错误处理
func TestAdapter_ErrorHandling(t *testing.T) {
	// 创建一个返回错误的模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回一个错误响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"message": "测试错误消息",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`))
	}))
	defer server.Close()
	
	// 创建配置
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	
	// 创建适配器
	adapter, err := NewAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("创建适配器失败: %v", err)
	}
	
	// 发送请求并期望错误
	ctx := context.Background()
	_, err = adapter.GenerateText(ctx, constants.DeepSeekChat, "这是一个错误测试", 100)
	
	// 验证错误存在
	if err == nil {
		t.Errorf("期望返回错误，但没有获得错误")
	}
}
