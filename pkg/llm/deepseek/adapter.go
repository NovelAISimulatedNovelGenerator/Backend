// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Adapter 提供了简化的DeepSeek API接口
type Adapter struct {
	client *Client
}

// NewAdapter 创建一个新的DeepSeek API适配器
func NewAdapter(apiKey string) (*Adapter, error) {
	client, err := NewClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &Adapter{client: client}, nil
}

// NewAdapterWithConfig 使用自定义配置创建适配器
func NewAdapterWithConfig(config *Config) (*Adapter, error) {
	client, err := NewClientWithConfig(config)
	if err != nil {
		return nil, err
	}

	return &Adapter{client: client}, nil
}

// Client 返回底层客户端
func (a *Adapter) Client() *Client {
	return a.client
}

// GetModels 获取可用模型列表
func (a *Adapter) GetModels(ctx context.Context) ([]string, error) {
	return a.client.Models(ctx)
}

// GenerateText 生成文本（非流式）
func (a *Adapter) GenerateText(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
	// 创建请求
	req := &CompletionRequest{
		Model:     model,
		Prompt:    prompt,
		MaxTokens: maxTokens,
	}

	// 发送请求
	resp, err := a.client.Completion(ctx, req)
	if err != nil {
		return "", err
	}

	// 提取生成的文本
	text := ""
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if textVal, ok := choice["text"].(string); ok {
				text = textVal
			}
		}
	}

	return text, nil
}

// GenerateTextStream 流式生成文本
func (a *Adapter) GenerateTextStream(ctx context.Context, model, prompt string, maxTokens int) (string, error) {
	// 创建请求
	req := &CompletionRequest{
		Model:     model,
		Prompt:    prompt,
		MaxTokens: maxTokens,
		Stream:    true,
	}

	// 发送流式请求
	stream, err := a.client.CompletionStream(ctx, req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	// 读取流式响应
	return a.readCompletionStream(stream)
}

// ChatWithSystem 使用系统提示进行聊天（非流式）
func (a *Adapter) ChatWithSystem(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	// 构建消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage(systemPrompt)
	msgBuilder.AddUserMessage(userPrompt)

	// 创建请求
	req := msgBuilder.CreateChatRequest(model, maxTokens)

	// 发送请求
	resp, err := a.client.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	// 提取生成的文本
	text := ""
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					text = content
				}
			}
		}
	}

	return text, nil
}

// ChatWithMessages 使用消息列表进行聊天（非流式）
func (a *Adapter) ChatWithMessages(ctx context.Context, model string, messages []Message, maxTokens int) (string, error) {
	// 创建请求
	req := &ChatRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: maxTokens,
	}

	// 发送请求
	resp, err := a.client.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	// 提取生成的文本
	text := ""
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					text = content
				}
			}
		}
	}

	return text, nil
}

// ChatWithSystemStream 使用系统提示进行流式聊天
func (a *Adapter) ChatWithSystemStream(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	// 构建消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage(systemPrompt)
	msgBuilder.AddUserMessage(userPrompt)

	// 创建请求
	req := msgBuilder.CreateChatRequest(model, maxTokens)
	req.Stream = true

	// 发送流式请求
	stream, err := a.client.ChatCompletionStream(ctx, req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	// 读取流式响应
	return a.readChatCompletionStream(stream)
}

// ChatWithMessagesStream 使用消息列表进行流式聊天
func (a *Adapter) ChatWithMessagesStream(ctx context.Context, model string, messages []Message, maxTokens int) (string, error) {
	// 创建请求
	req := &ChatRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    true,
	}

	// 发送流式请求
	stream, err := a.client.ChatCompletionStream(ctx, req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	// 读取流式响应
	return a.readChatCompletionStream(stream)
}

// readCompletionStream 从流式响应中读取文本完成内容
func (a *Adapter) readCompletionStream(stream *StreamReader) (string, error) {
	var fullText strings.Builder

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fullText.String(), fmt.Errorf("读取流失败: %w", err)
		}

		// 提取文本
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if text, ok := choice["text"].(string); ok {
					fullText.WriteString(text)
				}
			}
		}
	}

	return fullText.String(), nil
}

// readChatCompletionStream 从流式响应中读取聊天完成内容
func (a *Adapter) readChatCompletionStream(stream *StreamReader) (string, error) {
	var fullText strings.Builder

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fullText.String(), fmt.Errorf("读取流失败: %w", err)
		}

		// 提取增量内容
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						fullText.WriteString(content)
					}
				}
			}
		}
	}

	return fullText.String(), nil
}


