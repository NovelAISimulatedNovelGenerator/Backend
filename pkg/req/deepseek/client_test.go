// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import "novelai/pkg/constants"

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestConfigCreation 测试配置创建功能
func TestConfigCreation(t *testing.T) {
	// 创建一个基本配置
	config := DefaultConfig("test-api-key")
	
	// 验证默认值
	if config.APIKey != "test-api-key" {
		t.Errorf("期望APIKey为'test-api-key'，实际为'%s'", config.APIKey)
	}
	
	if config.BaseURL != DefaultDeepSeekBaseURL {
		t.Errorf("期望BaseURL为'%s'，实际为'%s'", DefaultDeepSeekBaseURL, config.BaseURL)
	}
	
	// 测试链式配置方法
	customURL := "https://custom-api.deepseek.com/v1"
	customTimeout := 60 * time.Second
	customOrgID := "org-123456"
	customUserAgent := "DeepSeek-Go-Client/1.0"
	
	config.WithBaseURL(customURL).
		WithTimeout(customTimeout).
		WithOrgID(customOrgID).
		WithUserAgent(customUserAgent)
	
	// 验证自定义值
	if config.BaseURL != customURL {
		t.Errorf("期望BaseURL为'%s'，实际为'%s'", customURL, config.BaseURL)
	}
	
	if config.Timeout != customTimeout {
		t.Errorf("期望Timeout为%v，实际为%v", customTimeout, config.Timeout)
	}
	
	if config.OrgID != customOrgID {
		t.Errorf("期望OrgID为'%s'，实际为'%s'", customOrgID, config.OrgID)
	}
	
	if config.UserAgent != customUserAgent {
		t.Errorf("期望UserAgent为'%s'，实际为'%s'", customUserAgent, config.UserAgent)
	}
}

// TestClientCreation 测试客户端创建功能
func TestClientCreation(t *testing.T) {
	// 创建客户端
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	
	// 验证客户端属性
	if client.config.APIKey != "test-api-key" {
		t.Errorf("期望APIKey为'test-api-key'，实际为'%s'", client.config.APIKey)
	}
	
	// 测试使用自定义配置创建客户端
	config := DefaultConfig("custom-api-key").WithBaseURL("https://custom-url.com/v1")
	customClient, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("使用自定义配置创建客户端失败: %v", err)
	}
	
	if customClient.config.BaseURL != "https://custom-url.com/v1" {
		t.Errorf("期望BaseURL为'https://custom-url.com/v1'，实际为'%s'", customClient.config.BaseURL)
	}
}

// TestAdapterCreation 测试适配器创建功能
func TestAdapterCreation(t *testing.T) {
	// 创建适配器
	adapter, err := NewAdapter("test-api-key")
	if err != nil {
		t.Fatalf("创建适配器失败: %v", err)
	}
	
	// 验证适配器属性
	if adapter.client.config.APIKey != "test-api-key" {
		t.Errorf("期望APIKey为'test-api-key'，实际为'%s'", adapter.client.config.APIKey)
	}
	
	// 测试使用自定义配置创建适配器
	config := DefaultConfig("custom-api-key").WithBaseURL("https://custom-url.com/v1")
	customAdapter, err := NewAdapterWithConfig(config)
	if err != nil {
		t.Fatalf("使用自定义配置创建适配器失败: %v", err)
	}
	
	if customAdapter.client.config.BaseURL != "https://custom-url.com/v1" {
		t.Errorf("期望BaseURL为'https://custom-url.com/v1'，实际为'%s'", customAdapter.client.config.BaseURL)
	}
}

// mockServer 创建一个模拟的HTTP服务器
func mockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// TestCompletionRequest 测试文本生成请求
func TestCompletionRequest(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/completions" {
			t.Errorf("期望路径为'/completions'，实际为'%s'", r.URL.Path)
		}
		
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("期望方法为'POST'，实际为'%s'", r.Method)
		}
		
		// 验证请求头
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("期望Authorization为'Bearer test-api-key'，实际为'%s'", r.Header.Get("Authorization"))
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
					"text": "这是一个测试响应",
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
	})
	defer server.Close()
	
	// 创建使用模拟服务器的客户端
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	
	// 发送文本生成请求
	req := &CompletionRequest{
		Model:     constants.DeepSeekChat,
		Prompt:    "这是一个测试",
		MaxTokens: 100,
	}
	
	ctx := context.Background()
	resp, err := client.Completion(ctx, req)
	if err != nil {
		t.Fatalf("发送文本生成请求失败: %v", err)
	}
	
	// 验证响应
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("响应中没有choices字段或为空")
	}
	
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0]不是一个有效的JSON对象")
	}
	
	text, ok := choice["text"].(string)
	if !ok {
		t.Fatalf("choices[0].text不是一个有效的字符串")
	}
	
	if text != "这是一个测试响应" {
		t.Errorf("期望响应文本为'这是一个测试响应'，实际为'%s'", text)
	}
}

// TestChatRequest 测试聊天请求
func TestChatRequest(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(func(w http.ResponseWriter, r *http.Request) {
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
						"content": "这是一个聊天测试响应"
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
	})
	defer server.Close()
	
	// 创建使用模拟服务器的客户端
	config := DefaultConfig("test-api-key").WithBaseURL(server.URL)
	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	
	// 创建聊天消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个测试助手")
	msgBuilder.AddUserMessage("这是一个测试消息")
	
	// 发送聊天请求
	req := &ChatRequest{
		Model:     constants.DeepSeekChat,
		Messages:  msgBuilder.Messages(),
		MaxTokens: 100,
	}
	
	ctx := context.Background()
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("发送聊天请求失败: %v", err)
	}
	
	// 验证响应
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("响应中没有choices字段或为空")
	}
	
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0]不是一个有效的JSON对象")
	}
	
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0].message不是一个有效的JSON对象")
	}
	
	content, ok := message["content"].(string)
	if !ok {
		t.Fatalf("choices[0].message.content不是一个有效的字符串")
	}
	
	if content != "这是一个聊天测试响应" {
		t.Errorf("期望响应内容为'这是一个聊天测试响应'，实际为'%s'", content)
	}
}

// TestMessageBuilder 测试消息构建器
func TestMessageBuilder(t *testing.T) {
	// 创建消息构建器
	msgBuilder := NewMessageBuilder()
	
	// 添加消息
	msgBuilder.AddSystemMessage("系统消息")
	msgBuilder.AddUserMessage("用户消息")
	msgBuilder.AddAssistantMessage("助手消息")
	
	// 获取消息列表
	messages := msgBuilder.Messages()
	
	// 验证消息数量
	if len(messages) != 3 {
		t.Errorf("期望消息数量为3，实际为%d", len(messages))
	}
	
	// 验证消息内容
	if messages[0].Role != constants.RoleSystem || messages[0].Content != "系统消息" {
		t.Errorf("系统消息不匹配，期望为{%s, %s}，实际为{%s, %s}", 
			constants.RoleSystem, "系统消息", messages[0].Role, messages[0].Content)
	}
	
	if messages[1].Role != constants.RoleUser || messages[1].Content != "用户消息" {
		t.Errorf("用户消息不匹配，期望为{%s, %s}，实际为{%s, %s}", 
			constants.RoleUser, "用户消息", messages[1].Role, messages[1].Content)
	}
	
	if messages[2].Role != constants.RoleAssistant || messages[2].Content != "助手消息" {
		t.Errorf("助手消息不匹配，期望为{%s, %s}，实际为{%s, %s}", 
			constants.RoleAssistant, "助手消息", messages[2].Role, messages[2].Content)
	}
	
	// 测试创建聊天请求
	req := msgBuilder.CreateChatRequest(constants.DeepSeekChat, 100)
	
	// 验证请求属性
	if req.Model != constants.DeepSeekChat {
		t.Errorf("期望Model为'%s'，实际为'%s'", constants.DeepSeekChat, req.Model)
	}
	
	if req.MaxTokens != 100 {
		t.Errorf("期望MaxTokens为100，实际为%d", req.MaxTokens)
	}
	
	if len(req.Messages) != 3 {
		t.Errorf("期望Messages长度为3，实际为%d", len(req.Messages))
	}
}
