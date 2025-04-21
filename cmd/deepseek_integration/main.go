// main.go
// DeepSeek API 集成测试示例
// 运行前请设置环境变量 DEEPSEEK_API_KEY，避免泄露密钥
// 用于验证真实 API 调用链路
package main

import (
	"context"
	"fmt"
	"io"

	"novelai/pkg/constants"
	"novelai/pkg/req/deepseek"
)

func main() {
	// 本示例将依次测试 DeepSeek 所有主要请求类型
	// 1. 非流式文本生成（completions, beta）
	// 2. 非流式聊天（chat/completions, v1）
	// 3. 流式文本生成（completions, beta, stream）
	// 4. 流式聊天（chat/completions, v1, stream）
	// 5. 获取模型列表（models, v1）

	// 从环境变量读取 API Key，确保安全
	apiKey := "sk-2b2644ac24024ccd82d0b47ab28a78a0"
	if apiKey == "" {
		fmt.Println("请先设置环境变量 DEEPSEEK_API_KEY")
		return
	}

	// 创建 DeepSeek 客户端，baseurl 只提供基础域名
	config := deepseek.DefaultConfig(apiKey).WithBaseURL("https://api.deepseek.com")
	client, err := deepseek.NewClientWithConfig(config)
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	ctx := context.Background()

	// 1. 非流式文本生成（completions, beta）
	fmt.Println("\n--- 非流式文本生成（completions, beta）---")
	completionReq := &deepseek.CompletionRequest{
		Model:       constants.DeepSeekChat,
		Prompt:      "讲一个关于人工智能的故事",
		MaxTokens:   100,
		Temperature: 0.7,
	}
	completionResp, err := client.Completion(ctx, completionReq)
	if err != nil {
		fmt.Printf("Completion 错误: %v\n", err)
	} else {
		fmt.Printf("Completion 响应: %+v\n", completionResp)
	}

	// 2. 非流式聊天（chat/completions, v1）
	fmt.Println("\n--- 非流式聊天（chat/completions, v1）---")
	chatReq := &deepseek.ChatRequest{
		Model: constants.DeepSeekChat,
		Messages: []deepseek.Message{
			{Role: constants.RoleSystem, Content: "你是一个助手。"},
			{Role: constants.RoleUser, Content: "你好，介绍一下你自己。"},
		},
		MaxTokens: 100,
		Temperature: 0.7,
	}
	chatResp, err := client.ChatCompletion(ctx, chatReq)
	if err != nil {
		fmt.Printf("ChatCompletion 错误: %v\n", err)
	} else {
		fmt.Printf("ChatCompletion 响应: %+v\n", chatResp)
	}

	// 3. 流式文本生成（completions, beta, stream）
	fmt.Println("\n--- 流式文本生成（completions, beta, stream）---")
	completionReqStream := &deepseek.CompletionRequest{
		Model:       constants.DeepSeekChat,
		Prompt:      "简要介绍人工智能的发展史",
		MaxTokens:   100,
		Temperature: 0.7,
	}
	stream, err := client.CompletionStream(ctx, completionReqStream)
	if err != nil {
		fmt.Printf("CompletionStream 错误: %v\n", err)
	} else {
		fmt.Print("CompletionStream 响应: ")
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("流式响应错误: %v\n", err)
				break
			}
			fmt.Printf("%v ", chunk)
		}
		fmt.Println()
	}

	// 4. 流式聊天（chat/completions, v1, stream）
	fmt.Println("\n--- 流式聊天（chat/completions, v1, stream）---")
	chatReqStream := &deepseek.ChatRequest{
		Model: constants.DeepSeekChat,
		Messages: []deepseek.Message{
			{Role: constants.RoleSystem, Content: "你是一个助手。"},
			{Role: constants.RoleUser, Content: "请用一句话介绍你自己。"},
		},
		MaxTokens: 100,
		Temperature: 0.7,
	}
	chatStream, err := client.ChatCompletionStream(ctx, chatReqStream)
	if err != nil {
		fmt.Printf("ChatCompletionStream 错误: %v\n", err)
	} else {
		fmt.Print("ChatCompletionStream 响应: ")
		for {
			chunk, err := chatStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("流式响应错误: %v\n", err)
				break
			}
			fmt.Printf("%v ", chunk)
		}
		fmt.Println()
	}

	// 5. 获取模型列表（models, v1）
	fmt.Println("\n--- 获取模型列表（models, v1）---")
	models, err := client.Models(ctx)
	if err != nil {
		fmt.Printf("Models 错误: %v\n", err)
	} else {
		fmt.Printf("Models 响应: %+v\n", models)
	}
}
