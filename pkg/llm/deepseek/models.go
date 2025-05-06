// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import "novelai/pkg/constants"

// DeepSeek相关常量全部迁移至pkg/constants/deepseek.go，统一维护
// 使用方式如 constants.DeepSeekChat, constants.constants.RoleSystem 等

// Message 表示一个聊天消息
type Message struct {
	// Role 是消息的角色，可以是 system、user 或 assistant
	Role string `json:"role"`

	// Content 是消息的内容
	Content string `json:"content"`
}

// ResponseFormat 指定模型输出格式（如 JSON 模式）
type ResponseFormat struct {
	// Type 指定输出类型，可选值为 "text" 或 "json_object"
	// - "text"：默认值，输出普通文本
	// - "json_object"：启用 JSON 模式，强制输出有效 JSON
	// 注意：若启用 JSON 模式，需在系统/用户消息中明确指示模型输出 JSON，否则可能生成空白字符直至超出 max_tokens，导致响应变慢或被截断。
	Type string `json:"type"` // 只能为 "text" 或 "json_object"，默认 "text"
}

// CompletionRequest 表示文本生成请求
type CompletionRequest struct {
	// Model 是使用的模型名称
	Model string `json:"model"`

	// Prompt 是提示语
	Prompt string `json:"prompt,omitempty"`

	// MaxTokens 是生成的最大token数量
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature 控制生成的随机性
	Temperature float64 `json:"temperature,omitempty"`

	// TopP 控制采样的token占比
	TopP float64 `json:"top_p,omitempty"`

	// N 生成的结果数量
	N int `json:"n,omitempty"`

	// Stream 是否使用流式响应
	Stream bool `json:"stream,omitempty"`

	// Stop 是停止生成的序列
	Stop []string `json:"stop,omitempty"`
}

// ChatRequest 表示聊天生成请求
type ChatRequest struct {
	// Model 是使用的模型名称
	Model string `json:"model"`

	// Messages 是聊天消息列表
	Messages []Message `json:"messages"`

	// MaxTokens 是生成的最大token数量
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature 控制生成的随机性
	Temperature float64 `json:"temperature,omitempty"`

	// TopP 控制采样的token占比
	TopP float64 `json:"top_p,omitempty"`

	// N 生成的结果数量
	N int `json:"n,omitempty"`

	// Stream 是否使用流式响应
	Stream bool `json:"stream,omitempty"`

	// Stop 是停止生成的序列
	Stop []string `json:"stop,omitempty"`

	// ResponseFormat 指定模型输出格式（如 JSON 模式）
	ResponseFormat ResponseFormat `json:"response_format,omitempty"`
}

// MessageBuilder 用于构建聊天消息序列
type MessageBuilder struct {
	messages []Message
}

// NewMessageBuilder 创建一个新的消息构建器
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		messages: make([]Message, 0),
	}
}

// AddSystemMessage 添加一个系统消息
func (b *MessageBuilder) AddSystemMessage(content string) *MessageBuilder {
	b.messages = append(b.messages, Message{
		Role:    constants.RoleSystem,
		Content: content,
	})
	return b
}

// AddUserMessage 添加一个用户消息
func (b *MessageBuilder) AddUserMessage(content string) *MessageBuilder {
	b.messages = append(b.messages, Message{
		Role:    constants.RoleUser,
		Content: content,
	})
	return b
}

// AddAssistantMessage 添加一个助手消息
func (b *MessageBuilder) AddAssistantMessage(content string) *MessageBuilder {
	b.messages = append(b.messages, Message{
		Role:    constants.RoleAssistant,
		Content: content,
	})
	return b
}

// Messages 返回所有消息
func (b *MessageBuilder) Messages() []Message {
	return b.messages
}

// CreateChatRequest 使用消息创建一个聊天请求
func (b *MessageBuilder) CreateChatRequest(model string, maxTokens int) *ChatRequest {
	return &ChatRequest{
		Model:     model,
		Messages:  b.messages,
		MaxTokens: maxTokens,
	}
}
