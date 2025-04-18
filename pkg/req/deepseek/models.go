// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

// DeepSeek模型常量
const (
	// DeepSeekChat 是DeepSeek的通用聊天模型
	DeepSeekChat = "deepseek-chat"
	
	// DeepSeekCoder 是DeepSeek的代码生成模型
	DeepSeekCoder = "deepseek-coder"
	
	// DeepSeekMax 是DeepSeek的大参数模型
	DeepSeekMax = "deepseek-llm-67b-max"
	
	// DeepSeek7B 是DeepSeek的7B参数模型
	DeepSeek7B = "deepseek-llm-7b-base"
)

// 角色常量，与OpenAI保持兼容
const (
	// RoleSystem 是系统角色
	RoleSystem = "system"
	
	// RoleUser 是用户角色
	RoleUser = "user"
	
	// RoleAssistant 是助手角色
	RoleAssistant = "assistant"
)

// Message 表示一个聊天消息
type Message struct {
	// Role 是消息的角色，可以是 system、user 或 assistant
	Role string `json:"role"`
	
	// Content 是消息的内容
	Content string `json:"content"`
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
		Role:    RoleSystem,
		Content: content,
	})
	return b
}

// AddUserMessage 添加一个用户消息
func (b *MessageBuilder) AddUserMessage(content string) *MessageBuilder {
	b.messages = append(b.messages, Message{
		Role:    RoleUser,
		Content: content,
	})
	return b
}

// AddAssistantMessage 添加一个助手消息
func (b *MessageBuilder) AddAssistantMessage(content string) *MessageBuilder {
	b.messages = append(b.messages, Message{
		Role:    RoleAssistant,
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
