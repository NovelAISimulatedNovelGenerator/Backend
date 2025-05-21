// Package model 提供多层代理系统的模型接口层实现
package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// DeepSeekModel 实现了基于DeepSeek API的Model接口
// 提供云端高性能模型服务，支持结构化输出和高级推理能力
type DeepSeekModel struct {
	*ModelWrapper
	apiKey     string
	baseURL    string
	httpClient *http.Client
	options    ModelOptions
}

// DeepSeekMessage 定义了DeepSeek API的消息格式
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequestBody 定义了DeepSeek API的请求体
type DeepSeekRequestBody struct {
	Model       string           `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Tools       json.RawMessage  `json:"tools,omitempty"`
	ToolChoice  interface{}      `json:"tool_choice,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type,omitempty"`
	} `json:"response_format,omitempty"`
}

// DeepSeekResponse 定义了DeepSeek API的响应格式
type DeepSeekResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Created   int64  `json:"created"`
	Model     string `json:"model"`
	Choices   []struct {
		Index        int `json:"index"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewDeepSeekModel 创建新的DeepSeek API模型实例
func NewDeepSeekModel(options ModelOptions) (Model, error) {
	// 验证必要参数
	if options.APIToken == "" {
		return nil, fmt.Errorf("缺少DeepSeek API令牌")
	}

	// 设置默认模型和URL
	if options.ModelName == "" {
		options.ModelName = "deepseek-chat" // 默认模型
		fmt.Printf("未指定DeepSeek模型名称，使用默认模型: %s\n", options.ModelName)
	}

	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1" // 默认API端点
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 120 * time.Second, // 设置较长的超时时间，适用于复杂生成任务
	}

	// 确定模型特性和限制
	tokenLimit := 8192 // 默认值
	visionSupport := false
	jsonSupport := true

	// 根据模型名称设置正确的模型参数
	switch {
	case strings.Contains(options.ModelName, "deepseek-coder"):
		tokenLimit = 16384 
	case strings.Contains(options.ModelName, "deepseek-llm-67b"):
		tokenLimit = 4096
	case strings.Contains(options.ModelName, "deepseek-vl"):
		tokenLimit = 8192
		visionSupport = true
	case strings.Contains(options.ModelName, "deepseek-chat"):
		tokenLimit = 8192
	}

	// 创建基础ModelWrapper
	wrapper := &ModelWrapper{
		BaseModel:        nil, // DeepSeek模型不直接使用LangChain Go的基础模型
		Type:             ModelTypeDeepSeek,
		Name:             options.ModelName,
		TokenLimit:       tokenLimit,
		JSONSupport:      jsonSupport,
		StreamingSupport: true,
		VisionSupport:    visionSupport,
	}

	// 创建DeepSeekModel实例
	model := &DeepSeekModel{
		ModelWrapper: wrapper,
		apiKey:       options.APIToken,
		baseURL:      baseURL,
		httpClient:   httpClient,
		options:      options,
	}

	fmt.Printf("成功创建DeepSeek模型: %s (token限制: %d)\n", options.ModelName, tokenLimit)
	return model, nil
}

// Call 实现Model接口的Call方法
func (m *DeepSeekModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	if m.options.Debug {
		fmt.Printf("[DeepSeek调用] 模型: %s, 提示词长度: %d字符\n", m.Name, len(prompt))
	}

	// 处理调用选项
	callOptions := &llms.CallOptions{}
	for _, opt := range options {
		opt(callOptions)
	}

	// 准备DeepSeek API请求消息
	messages := []DeepSeekMessage{
		{Role: "user", Content: prompt},
	}

	// 发送请求
	response, err := m.sendRequest(ctx, messages, callOptions)
	if err != nil {
		fmt.Printf("DeepSeek API调用失败: %v\n", err)
		return "", fmt.Errorf("DeepSeek API调用失败: %w", err)
	}

	// 提取响应内容
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek API返回空响应")
	}

	result := response.Choices[0].Message.Content
	if m.options.Debug {
		fmt.Printf("[DeepSeek响应] 长度: %d字符, 使用tokens: %d\n", len(result), response.Usage.TotalTokens)
	}

	return result, nil
}

// GenerateContent 实现Model接口的GenerateContent方法
func (m *DeepSeekModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if m.options.Debug {
		fmt.Printf("[DeepSeek生成内容] 模型: %s, 消息数: %d\n", m.Name, len(messages))
	}

	// 处理调用选项
	callOptions := &llms.CallOptions{}
	for _, opt := range options {
		opt(callOptions)
	}

	// 将LangChain消息转换为DeepSeek API消息格式
	deepseekMessages := []DeepSeekMessage{}
	for _, msg := range messages {
		content := ""
		
		// 处理多模态内容
		for _, part := range msg.Parts {
			switch v := part.(type) {
			case llms.TextContent:
				content += v.Text
			// 如果需要支持图像等其他内容，这里可以扩展
			default:
				if m.options.Debug {
					fmt.Printf("不支持的内容类型: %T\n", part)
				}
			}
		}

		role := "user"
		switch msg.Role {
		case llms.ChatMessageTypeAI:
			role = "assistant"
		case llms.ChatMessageTypeSystem:
			role = "system"
		case llms.ChatMessageTypeHuman:
			role = "user"
		}

		deepseekMessages = append(deepseekMessages, DeepSeekMessage{
			Role:    role,
			Content: content,
		})
	}

	// 发送请求
	response, err := m.sendRequest(ctx, deepseekMessages, callOptions)
	if err != nil {
		fmt.Printf("DeepSeek内容生成失败: %v\n", err)
		return nil, fmt.Errorf("DeepSeek内容生成失败: %w", err)
	}

	// 转换为LangChain ContentResponse格式
	contentResponse := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{},
	}

	for _, choice := range response.Choices {
		contentResponse.Choices = append(contentResponse.Choices, &llms.ContentChoice{
			Content:     choice.Message.Content,
			StopReason:  choice.FinishReason,
			GenerationInfo: map[string]any{
				"prompt_tokens":     response.Usage.PromptTokens,
				"completion_tokens": response.Usage.CompletionTokens,
				"total_tokens":      response.Usage.TotalTokens,
			},
		})
	}

	if m.options.Debug && len(contentResponse.Choices) > 0 {
		fmt.Printf("[DeepSeek内容响应] 选项数: %d, 总tokens: %d\n", 
			len(contentResponse.Choices), 
			response.Usage.TotalTokens)
	}

	return contentResponse, nil
}

// sendRequest 发送请求到DeepSeek API并解析响应
func (m *DeepSeekModel) sendRequest(ctx context.Context, messages []DeepSeekMessage, callOptions *llms.CallOptions) (*DeepSeekResponse, error) {
	// 暂时使用模拟响应，实际实现需要发送HTTP请求到DeepSeek API
	// 这里仅为示例，实际使用时应移除此代码
	mockResponse := &DeepSeekResponse{
		ID:        "mock-response-id",
		Object:    "chat.completion",
		Created:   time.Now().Unix(),
		Model:     m.Name,
		Choices: []struct {
			Index        int `json:"index"`
			Message      struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "这是一个模拟的DeepSeek API响应。在实际实现中，这里应该是模型生成的真实内容。",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}

	// TODO: 实现实际的DeepSeek API请求逻辑
	// 1. 构建请求体
	// 2. 发送HTTP请求
	// 3. 解析响应
	// 4. 处理错误

	return mockResponse, nil
}
