// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import (
	"context"
	"fmt"
	"io"
	"novelai/pkg/constants"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 示例: 如何使用非流式文本生成
func ExampleCompletion() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建客户端，使用环境变量中的API密钥
	client, err := NewClient(apiKey)
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	// 创建请求
	req := &CompletionRequest{
		Model:       constants.DeepSeekChat,
		Prompt:      "给我讲一个关于人工智能的故事",
		MaxTokens:   500,
		Temperature: 0.7,
	}

	// 发送请求
	ctx := context.Background()
	resp, err := client.Completion(ctx, req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 处理响应
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if text, ok := choice["text"].(string); ok {
				fmt.Printf("生成的文本: %s\n", text)
			}
		}
	}
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if total, ok := usage["total_tokens"].(float64); ok {
			fmt.Printf("使用令牌: %d\n", int(total))
		}
	}
}

// 示例: 如何使用非流式聊天
func ExampleChatCompletion() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建客户端，使用环境变量中的API密钥
	client, err := NewClient(apiKey)
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	// 使用MessageBuilder创建聊天消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个小说创作助手，擅长奇幻故事。")
	msgBuilder.AddUserMessage("请用300字左右创作一个奇幻世界的短篇故事，主角是一个有魔法能力的年轻人。")

	// 创建请求
	req := &ChatRequest{
		Model:       constants.DeepSeekChat,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   500,
		Temperature: 0.7,
	}

	// 发送请求
	ctx := context.Background()
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 处理响应
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					fmt.Printf("生成的回复: %s\n", content)
				}
			}
		}
	}
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if total, ok := usage["total_tokens"].(float64); ok {
			fmt.Printf("使用令牌: %d\n", int(total))
		}
	}
}

// 示例: 如何使用流式文本生成
func ExampleCompletionStream() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建客户端，使用环境变量中的API密钥
	client, err := NewClient(apiKey)
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	// 使用MessageBuilder创建聊天消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个专业的小说创作助手，擅长科幻题材。")
	msgBuilder.AddUserMessage("请创作一个关于未来人类与人工智能共存的短篇故事，500字左右。")

	// 创建请求
	req := &ChatRequest{
		Model:       constants.DeepSeekChat,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   500,
		Temperature: 0.7,
		Stream:      true,
	}

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	streamResp, err := client.ChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 设置信号处理，以便可以通过Ctrl+C取消
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建一个上下文取消通道
	done := make(chan struct{})

	// 在goroutine中处理响应
	go func() {
		defer close(done)

		// 全文缓冲
		var fullText string

		for {
			chunk, err := streamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("流接收错误: %v\n", err)
				break
			}

			// 处理文本块
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if message, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := message["content"].(string); ok {
							fullText += content
							fmt.Print(content)
						}
					}
				}
			}
		}

		fmt.Printf("\n\n完整生成文本:\n%s\n", fullText)
	}()

	// 等待完成或取消
	select {
	case <-sigChan:
		cancel()
		fmt.Println("\n生成已取消")
	case <-done:
		fmt.Println("\n生成已完成")
	}
}

// 示例: 如何使用流式聊天
func ExampleChatCompletionStream() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建客户端，使用环境变量中的API密钥
	client, err := NewClient(apiKey)
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	// 使用MessageBuilder创建聊天消息
	msgBuilder := NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个小说创作助手，擅长奇幻故事。")
	msgBuilder.AddUserMessage("请用300字左右创作一个奇幻世界的短篇故事，主角是一个有魔法能力的年轻人。")

	// 创建请求
	req := &ChatRequest{
		Model:       constants.DeepSeekChat,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   500,
		Temperature: 0.7,
		Stream:      true,
	}

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	streamResp, err := client.ChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 设置信号处理，以便可以通过Ctrl+C取消
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建一个上下文取消通道
	done := make(chan struct{})

	// 在goroutine中处理响应
	go func() {
		defer close(done)

		// 全文缓冲
		var fullText string

		for {
			chunk, err := streamResp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("流接收错误: %v\n", err)
				break
			}

			// 处理文本块
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fullText += content
							fmt.Print(content)
						}
					}
				}
			}
		}

		fmt.Printf("\n\n完整生成文本:\n%s\n", fullText)
	}()

	// 等待完成或取消
	select {
	case <-sigChan:
		cancel()
		fmt.Println("\n生成已取消")
	case <-done:
		fmt.Println("\n生成已完成")
	}
}

// 示例: 如何使用简化的适配器接口
func ExampleAdapter() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建适配器，使用环境变量中的API密钥
	adapter, err := NewAdapter(apiKey)
	if err != nil {
		fmt.Printf("创建适配器错误: %v\n", err)
		return
	}

	// 简单的文本生成
	ctx := context.Background()
	result, err := adapter.GenerateText(ctx, constants.DeepSeekChat, "讲一个有趣的故事", 500)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("生成的文本: %s\n", result)

	// 简单的聊天
	systemPrompt := "你是一个友好的助手。"
	userPrompt := "请介绍一下自己。"
	chatResult, err := adapter.ChatWithSystem(ctx, constants.DeepSeekChat, systemPrompt, userPrompt, 300)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("聊天回复: %s\n", chatResult)

	// 使用消息列表发送聊天请求
	messages := []Message{
		{Role: constants.RoleSystem, Content: "你是DeepSeek AI的文档助手。"},
		{Role: constants.RoleUser, Content: "什么是DeepSeek?"},
	}
	messagesResult, err := adapter.ChatWithMessages(ctx, constants.DeepSeekChat, messages, 300)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("消息列表聊天回复: %s\n", messagesResult)
}

// 示例: 如何使用适配器的流式API
func ExampleAdapterStream() {
	// 从环境变量读取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("环境变量 DEEPSEEK_API_KEY 未设置")
		return
	}
	// 创建适配器，使用环境变量中的API密钥
	adapter, err := NewAdapter(apiKey)
	if err != nil {
		fmt.Printf("创建适配器错误: %v\n", err)
		return
	}

	// 设置上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建一个上下文取消通道
	done := make(chan struct{})

	// 在goroutine中处理流式文本生成
	go func() {
		defer close(done)

		// 发送流式聊天请求
		result, err := adapter.ChatWithSystemStream(ctx, constants.DeepSeekChat,
			"你是一个创意写作助手。",
			"请用200字创作一个关于宇宙探索的短文。",
			500)

		if err != nil {
			fmt.Printf("流式聊天错误: %v\n", err)
			return
		}

		fmt.Printf("\n\n完整生成文本:\n%s\n", result)
	}()

	// 等待完成或取消
	select {
	case <-sigChan:
		fmt.Println("\n取消生成")
	case <-done:
		// 完成
	}
}
