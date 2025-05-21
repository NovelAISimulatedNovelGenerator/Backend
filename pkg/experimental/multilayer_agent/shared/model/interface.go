// Package model 提供多层代理系统的模型接口层实现
// 本包基于LangChain Go库，为上层代理提供统一、灵活的大语言模型访问能力
package model

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// ModelType 定义了支持的模型类型
type ModelType string

const (
	// ModelTypeOllama 表示Ollama本地模型
	ModelTypeOllama ModelType = "ollama"
	// ModelTypeDeepSeek 表示DeepSeek API云端模型
	ModelTypeDeepSeek ModelType = "deepseek"
	// ModelTypeOpenAI 表示OpenAI API模型
	ModelTypeOpenAI ModelType = "openai"
)

// Model 扩展了LangChain Go的llms.Model接口
// 提供了多层代理系统特定的功能和错误处理机制
type Model interface {
	// 继承LangChain Go的基本模型接口
	llms.Model

	// ModelType 返回模型的类型
	ModelType() ModelType

	// ModelName 返回具体的模型名称
	ModelName() string

	// GetTokenLimit 返回当前模型的最大token限制
	GetTokenLimit() int

	// EstimateTokens 估算输入文本的token数量
	EstimateTokens(text string) (int, error)

	// SupportsJSON 检查模型是否支持JSON输出格式
	SupportsJSON() bool

	// SupportsStreaming 检查模型是否支持流式输出
	SupportsStreaming() bool

	// SupportsVision 检查模型是否支持图像输入
	SupportsVision() bool
}

// ModelOptions 定义了创建模型时的选项
type ModelOptions struct {
	// 模型名称
	ModelName string

	// 模型端点URL
	BaseURL string

	// API令牌
	APIToken string

	// 默认生成参数
	DefaultTemperature float64
	DefaultMaxTokens   int
	DefaultTopP        float64
	DefaultTopK        int

	// 调试模式
	Debug bool
}

// ModelFactory 提供创建模型实例的工厂接口
type ModelFactory interface {
	// CreateModel 创建指定类型和配置的模型实例
	CreateModel(modelType ModelType, options ModelOptions) (Model, error)
}

// DefaultModelFactory 是ModelFactory的默认实现
type DefaultModelFactory struct{}

// NewModelFactory 创建一个新的模型工厂实例
func NewModelFactory() ModelFactory {
	return &DefaultModelFactory{}
}

// CreateModel 创建指定类型和配置的模型实例
func (f *DefaultModelFactory) CreateModel(modelType ModelType, options ModelOptions) (Model, error) {
	switch modelType {
	case ModelTypeOllama:
		return NewOllamaModel(options)
	case ModelTypeDeepSeek:
		return NewDeepSeekModel(options)
	case ModelTypeOpenAI:
		// 尚未实现
		return nil, fmt.Errorf("模型类型 %s 尚未实现", modelType)
	default:
		return nil, fmt.Errorf("未知的模型类型: %s", modelType)
	}
}

// ModelWrapper 提供了对llms.Model的基本包装
// 实现了Model接口中与llms.Model无关的通用方法
type ModelWrapper struct {
	// 被包装的LangChain Go模型
	BaseModel llms.Model

	// 模型类型
	Type ModelType

	// 模型名称
	Name string

	// 模型参数
	TokenLimit int
	JSONSupport bool
	StreamingSupport bool
	VisionSupport bool
}

// ModelType 实现Model接口
func (m *ModelWrapper) ModelType() ModelType {
	return m.Type
}

// ModelName 实现Model接口
func (m *ModelWrapper) ModelName() string {
	return m.Name
}

// GetTokenLimit 实现Model接口
func (m *ModelWrapper) GetTokenLimit() int {
	return m.TokenLimit
}

// EstimateTokens 实现Model接口
// 使用简单的字符数近似估算token数
func (m *ModelWrapper) EstimateTokens(text string) (int, error) {
	// 简单估算: 平均每个token约为4个字符
	// 实际应用中应使用更精确的tokenizer
	return len(text) / 4, nil
}

// SupportsJSON 实现Model接口
func (m *ModelWrapper) SupportsJSON() bool {
	return m.JSONSupport
}

// SupportsStreaming 实现Model接口
func (m *ModelWrapper) SupportsStreaming() bool {
	return m.StreamingSupport
}

// SupportsVision 实现Model接口
func (m *ModelWrapper) SupportsVision() bool {
	return m.VisionSupport
}

// Call 代理到基础模型的Call方法
func (m *ModelWrapper) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return m.BaseModel.Call(ctx, prompt, options...)
}

// GenerateContent 代理到基础模型的GenerateContent方法
func (m *ModelWrapper) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.BaseModel.GenerateContent(ctx, messages, options...)
}
