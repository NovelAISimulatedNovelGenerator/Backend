package core

import (
	"encoding/json"
	"time"
)

// MessageType 定义消息类型
type MessageType string

const (
	// 基础消息类型
	MessageTypeRequest      MessageType = "request"      // 请求消息
	MessageTypeResponse     MessageType = "response"     // 响应消息
	MessageTypeNotification MessageType = "notification" // 通知消息
	MessageTypeError        MessageType = "error"        // 错误消息

	// 智能体间通信消息类型
	MessageTypeCommand    MessageType = "command"    // 命令消息
	MessageTypeQuery      MessageType = "query"      // 查询消息
	MessageTypeReport     MessageType = "report"     // 报告消息
	MessageTypeToolCall   MessageType = "tool_call"  // 工具调用消息
	MessageTypeToolResult MessageType = "tool_result" // 工具结果消息
)

// MessagePriority 定义消息优先级
type MessagePriority string

const (
	MessagePriorityLow    MessagePriority = "low"    // 低优先级
	MessagePriorityNormal MessagePriority = "normal" // 普通优先级
	MessagePriorityHigh   MessagePriority = "high"   // 高优先级
	MessagePriorityUrgent MessagePriority = "urgent" // 紧急优先级
)

// Message 统一消息结构
// 用于智能体间的所有通信
type Message struct {
	// 基础字段
	ID        string          `json:"id"`         // 消息唯一标识
	Type      MessageType     `json:"type"`       // 消息类型
	Priority  MessagePriority `json:"priority"`   // 消息优先级
	Timestamp time.Time       `json:"timestamp"`  // 时间戳

	// 发送方和接收方信息
	From string `json:"from"` // 发送方ID
	To   string `json:"to"`   // 接收方ID（可选，空表示广播）

	// 消息内容
	Subject string                 `json:"subject"`      // 消息主题
	Content string                 `json:"content"`      // 消息内容（文本）
	Data    map[string]interface{} `json:"data"`        // 结构化数据
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata"` // 附加元数据

	// 关联信息
	CorrelationID string `json:"correlation_id,omitempty"` // 关联ID，用于追踪相关消息
	ReplyTo       string `json:"reply_to,omitempty"`       // 回复的消息ID
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, from, to string) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		Priority:  MessagePriorityNormal,
		Timestamp: time.Now(),
		From:      from,
		To:        to,
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
}

// generateMessageID 生成消息ID
// 使用时间戳和随机数组合
func generateMessageID() string {
	// 简化实现，实际应使用UUID或更复杂的ID生成策略
	return time.Now().Format("20060102150405.999999999")
}

// SetContent 设置消息内容
func (m *Message) SetContent(subject, content string) {
	m.Subject = subject
	m.Content = content
}

// SetData 设置结构化数据
func (m *Message) SetData(key string, value interface{}) {
	if m.Data == nil {
		m.Data = make(map[string]interface{})
	}
	m.Data[key] = value
}

// GetData 获取结构化数据
func (m *Message) GetData(key string) (interface{}, bool) {
	if m.Data == nil {
		return nil, false
	}
	value, exists := m.Data[key]
	return value, exists
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata 获取元数据
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}

// IsRequest 判断是否为请求消息
func (m *Message) IsRequest() bool {
	return m.Type == MessageTypeRequest || m.Type == MessageTypeCommand || m.Type == MessageTypeQuery
}

// IsResponse 判断是否为响应消息
func (m *Message) IsResponse() bool {
	return m.Type == MessageTypeResponse || m.Type == MessageTypeReport || m.Type == MessageTypeToolResult
}

// IsError 判断是否为错误消息
func (m *Message) IsError() bool {
	return m.Type == MessageTypeError
}

// Clone 克隆消息
// 创建消息的深拷贝
func (m *Message) Clone() *Message {
	clone := &Message{
		ID:            m.ID,
		Type:          m.Type,
		Priority:      m.Priority,
		Timestamp:     m.Timestamp,
		From:          m.From,
		To:            m.To,
		Subject:       m.Subject,
		Content:       m.Content,
		CorrelationID: m.CorrelationID,
		ReplyTo:       m.ReplyTo,
	}

	// 深拷贝Data
	if m.Data != nil {
		clone.Data = make(map[string]interface{})
		for k, v := range m.Data {
			clone.Data[k] = v
		}
	}

	// 深拷贝Metadata
	if m.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range m.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// ToJSON 将消息转换为JSON字符串
func (m *Message) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析消息
func FromJSON(jsonStr string) (*Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// CreateErrorMessage 创建错误消息
func CreateErrorMessage(from string, err error, replyTo string) *Message {
	msg := NewMessage(MessageTypeError, from, "")
	msg.Subject = "Error"
	msg.Content = err.Error()
	msg.ReplyTo = replyTo
	return msg
}

// CreateToolCallMessage 创建工具调用消息
func CreateToolCallMessage(from, toolName string, input interface{}) *Message {
	msg := NewMessage(MessageTypeToolCall, from, "tool_caller")
	msg.Subject = "Tool Call: " + toolName
	msg.SetData("tool_name", toolName)
	msg.SetData("input", input)
	return msg
}

// CreateToolResultMessage 创建工具结果消息
func CreateToolResultMessage(from string, result interface{}, replyTo string) *Message {
	msg := NewMessage(MessageTypeToolResult, from, "")
	msg.Subject = "Tool Result"
	msg.SetData("result", result)
	msg.ReplyTo = replyTo
	return msg
}
