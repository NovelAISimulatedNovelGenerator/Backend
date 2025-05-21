// Package model 提供多层代理系统的模型接口层实现
package model

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms/ollama"
)

// OllamaModel 实现了基于Ollama API的Model接口
// 直接封装LangChain Go的官方Ollama实现
type OllamaModel struct {
	*ModelWrapper
	client  *ollama.LLM
	options ModelOptions
}

// NewOllamaModel 创建新的Ollama模型实例
func NewOllamaModel(options ModelOptions) (Model, error) {
	// 设置默认模型名称
	if options.ModelName == "" {
		options.ModelName = "llama2"
		fmt.Printf("未指定Ollama模型名称，使用默认模型: %s\n", options.ModelName)
	}

	// 准备Ollama选项
	opts := []ollama.Option{
		ollama.WithModel(options.ModelName),
	}

	// 设置自定义URL（如果提供）
	if options.BaseURL != "" {
		// 注意：需要通过环境变量OLLAMA_HOST设置BaseURL
		// 这里我们只记录日志，实际设置需要在应用启动前完成
		fmt.Printf("Ollama API使用以下端点(应通过OLLAMA_HOST环境变量设置): %s\n", options.BaseURL)
		
		// 设置HTTP客户端
		httpClient := &http.Client{}
		opts = append(opts, ollama.WithHTTPClient(httpClient))
	}

	// 如果指定了格式化输出
	if options.Debug {
		// Ollama支持JSON输出
		opts = append(opts, ollama.WithFormat("json"))
	}

	// 创建Ollama客户端
	client, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建Ollama客户端失败: %w", err)
	}

	// 确定模型特性和限制
	tokenLimit := 4096 // 默认值
	switch {
	case strings.Contains(options.ModelName, "llama3"):
		tokenLimit = 8192
	case strings.Contains(options.ModelName, "llama2"):
		tokenLimit = 4096
	case strings.Contains(options.ModelName, "mistral"):
		tokenLimit = 8192
	case strings.Contains(options.ModelName, "mixtral"):
		tokenLimit = 32768
	}

	// 创建基础ModelWrapper
	wrapper := &ModelWrapper{
		BaseModel:        client,
		Type:             ModelTypeOllama,
		Name:             options.ModelName,
		TokenLimit:       tokenLimit,
		JSONSupport:      true,  // Ollama支持JSON输出
		StreamingSupport: true,  // Ollama支持流式输出
		VisionSupport:    false, // 大多数Ollama模型不支持视觉输入
	}

	// 检测是否为多模态模型
	if strings.Contains(strings.ToLower(options.ModelName), "llava") ||
		strings.Contains(strings.ToLower(options.ModelName), "bakllava") {
		wrapper.VisionSupport = true
	}

	// 创建并返回OllamaModel实例
	model := &OllamaModel{
		ModelWrapper: wrapper,
		client:       client,
		options:      options,
	}

	fmt.Printf("成功创建Ollama模型: %s (token限制: %d)\n", options.ModelName, tokenLimit)
	return model, nil
}
