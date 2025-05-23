package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/llms"

	"novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// GenericAdvancedAgent 通用高级智能体
// 这是一个示例实现，展示如何创建同时支持工具和记忆的智能体
type GenericAdvancedAgent struct {
	*BaseAdvancedAgent           // 嵌入基础高级智能体实现
	prompt             string    // 提示模板
	maxRetries         int       // 最大重试次数
	lastProcessTime    time.Time // 上次处理时间
}

// NewGenericAdvancedAgent 创建新的通用高级智能体
func NewGenericAdvancedAgent(id string, agentType AgentType, prompt string) *GenericAdvancedAgent {
	agent := &GenericAdvancedAgent{
		BaseAdvancedAgent: NewBaseAdvancedAgent(id, agentType),
		prompt:            prompt,
		maxRetries:        3,
	}
	return agent
}

// Initialize 实现Agent接口，进行初始化
func (a *GenericAdvancedAgent) Initialize(ctx context.Context) error {
	hlog.CtxInfof(ctx, "初始化通用高级智能体: ID=%s, Type=%s", a.GetID(), a.GetType())

	// 调用基础初始化
	if err := a.BaseAdvancedAgent.Initialize(ctx); err != nil {
		return err
	}

	// 加载之前的状态（如果有）
	if a.GetMemoryManager() != nil {
		stateKey := memory.CreateTaggedKey(a.GetID(), "state", "last_process_time")
		if val, err := a.LoadMemory(ctx, stateKey); err == nil && val != nil {
			if timeStr, ok := val.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
					a.lastProcessTime = parsedTime
					hlog.CtxInfof(ctx, "从记忆加载上次处理时间: %v", a.lastProcessTime)
				}
			}
		}
	}

	return nil
}

// Process 实现Agent接口，处理消息
func (a *GenericAdvancedAgent) Process(ctx context.Context, msg *Message) (*Message, error) {
	a.SetStatus(AgentStatusWorking)
	defer a.SetStatus(AgentStatusIdle)

	hlog.CtxInfof(ctx, "处理消息: ID=%s, Type=%s, Subject=%s",
		msg.ID, msg.Type, msg.Subject)

	// 记录处理时间
	now := time.Now()
	a.lastProcessTime = now

	// 记录到记忆
	if a.GetMemoryManager() != nil {
		stateKey := memory.CreateTaggedKey(a.GetID(), "state", "last_process_time")
		if err := a.SaveMemory(ctx, stateKey, now.Format(time.RFC3339)); err != nil {
			hlog.CtxWarnf(ctx, "保存处理时间到记忆失败: %v", err)
		}

		// 保存消息历史
		historyKey := memory.CreateTaggedKey(a.GetID(), "history", msg.ID)
		if err := a.SaveMemory(ctx, historyKey, msg); err != nil {
			hlog.CtxWarnf(ctx, "保存消息历史到记忆失败: %v", err)
		}
	}

	// 注意：不再直接拦截工具调用消息
	// 所有消息类型都交由模型处理，让模型决定是否需要调用工具或发送消息给其他模型

	// 检查模型是否已设置
	if a.GetModel() == nil {
		hlog.CtxErrorf(ctx, "未设置语言模型，智能体无法处理消息")
		return nil, fmt.Errorf("未设置语言模型，智能体无法处理消息")
	}

	// 构建模型输入
	promptTemplate := a.prompt
	if promptTemplate == "" {
		// 默认提示模板包含智能体角色、工具能力和通信能力的说明
		promptTemplate = `你是一个智能体，类型为%s。你可以：
1. 处理用户消息并直接回复
2. 调用可用工具完成任务
3. 向其他智能体发送消息

消息类型: %s
消息来源: %s
消息主题: %s

消息内容:
%s

如需调用工具，请使用格式：
{"tool":"工具名称","input":"参数"}

如需发送消息给其他智能体，请使用格式：
{"send_to":"目标智能体ID","message":"消息内容"}`
	}
	
	// 生成提示，包含完整的消息上下文
	prompt := fmt.Sprintf(promptTemplate, 
		a.GetType(), 
		string(msg.Type),
		msg.From,
		msg.Subject,
		msg.Content)

	// 调用模型生成内容
	hlog.CtxInfof(ctx, "调用模型处理消息：%s (模型：%s)", msg.Subject, a.GetModel().ModelName())

	var modelResponse string
	var err error

	if a.GetModel().SupportsJSON() {
		// 使用JSON模式
		messages := []llms.MessageContent{
			{
				Role: "system",
				Parts: []llms.ContentPart{
					llms.TextPart(fmt.Sprintf("你是一个智能体，类型为%s。请以JSON格式回复。", a.GetType())),
				},
			},
			{
				Role: "user",
				Parts: []llms.ContentPart{
					llms.TextPart(msg.Content),
				},
			},
		}

		// 使用GenerateContent方法
		contentResponse, err := a.GetModel().GenerateContent(ctx, messages)
		if err != nil {
			hlog.CtxErrorf(ctx, "模型生成内容失败: %v", err)
			return nil, fmt.Errorf("模型生成内容失败: %w", err)
		}

		if len(contentResponse.Choices) > 0 {
			modelResponse = contentResponse.Choices[0].Content
		}
	} else {
		// 使用普通文本模式
		modelResponse, err = a.GetModel().Call(ctx, prompt)
		if err != nil {
			hlog.CtxErrorf(ctx, "模型调用失败: %v", err)
			return nil, fmt.Errorf("模型调用失败: %w", err)
		}
	}

	// 解析模型响应，支持工具调用和模型间通信
	var toolCall struct {
		Tool  string `json:"tool"`
		Input string `json:"input"`
	}

	var sendMessage struct {
		SendTo  string `json:"send_to"`
		Message string `json:"message"`
	}

	// 尝试解析工具调用
	if err := json.Unmarshal([]byte(modelResponse), &toolCall); err == nil && toolCall.Tool != "" {
		hlog.CtxInfof(ctx, "智能体请求调用工具: %s, 输入: %s", toolCall.Tool, toolCall.Input)

		// 检查工具调用器
		if a.GetToolCaller() == nil {
			return nil, fmt.Errorf("智能体没有工具调用器，无法执行工具调用请求")
		}

		// 执行工具调用
		toolResult, err := a.CallTool(ctx, toolCall.Tool, toolCall.Input)
		if err != nil {
			hlog.CtxErrorf(ctx, "工具调用失败: %v", err)
			return nil, fmt.Errorf("工具调用失败: %w", err)
		}

		// 创建工具调用结果消息
		response := NewMessage(MessageTypeToolResult, a.GetID(), msg.From)
		response.Subject = "工具调用结果: " + toolCall.Tool
		response.Content = toolResult
		response.ReplyTo = msg.ID
		response.SetMetadata("tool_name", toolCall.Tool)
		response.SetMetadata("process_time", time.Since(now).String())
		response.SetMetadata("agent_type", string(a.GetType()))

		return response, nil
	}

	// 尝试解析模型间通信
	if err := json.Unmarshal([]byte(modelResponse), &sendMessage); err == nil && sendMessage.SendTo != "" {
		hlog.CtxInfof(ctx, "智能体请求发送消息给: %s", sendMessage.SendTo)

		// 创建发送给其他智能体的消息
		response := NewMessage(MessageTypeRequest, a.GetID(), sendMessage.SendTo)
		response.Subject = fmt.Sprintf("来自%s的消息", a.GetType())
		response.Content = sendMessage.Message
		response.ReplyTo = msg.ID
		response.SetMetadata("original_from", msg.From)
		response.SetMetadata("process_time", time.Since(now).String())
		response.SetMetadata("agent_type", string(a.GetType()))

		return response, nil
	}

	// 如果不是工具调用或模型间通信，则创建普通响应消息
	response := NewMessage(MessageTypeResponse, a.GetID(), msg.From)
	response.Subject = "处理结果: " + msg.Subject
	response.Content = modelResponse
	response.ReplyTo = msg.ID

	// 添加处理元数据
	response.SetMetadata("process_time", time.Since(now).String())
	response.SetMetadata("agent_type", string(a.GetType()))
	response.SetMetadata("model_name", a.GetModel().ModelName())
	response.SetMetadata("model_type", string(a.GetModel().ModelType()))

	return response, nil
}

// handleToolCallMessage 处理工具调用消息
func (a *GenericAdvancedAgent) handleToolCallMessage(ctx context.Context, msg *Message) (*Message, error) {
	// 获取工具名称和输入
	toolName, ok := msg.GetData("tool_name")
	if !ok {
		return CreateErrorMessage(a.GetID(), fmt.Errorf("缺少工具名称"), msg.ID), nil
	}

	input, ok := msg.GetData("input")
	if !ok {
		return CreateErrorMessage(a.GetID(), fmt.Errorf("缺少工具输入"), msg.ID), nil
	}

	// 转换为字符串
	toolNameStr, ok := toolName.(string)
	if !ok {
		return CreateErrorMessage(a.GetID(), fmt.Errorf("工具名称必须是字符串"), msg.ID), nil
	}

	inputStr := fmt.Sprintf("%v", input)

	// 调用工具
	hlog.CtxInfof(ctx, "调用工具: %s, 输入: %s", toolNameStr, inputStr)
	if a.GetToolCaller() == nil {
		return CreateErrorMessage(a.GetID(), fmt.Errorf("工具调用器未设置"), msg.ID), nil
	}

	result, err := a.CallTool(ctx, toolNameStr, inputStr)
	if err != nil {
		hlog.CtxErrorf(ctx, "工具调用失败: %v", err)
		return CreateErrorMessage(a.GetID(), err, msg.ID), nil
	}

	// 创建工具结果消息
	response := CreateToolResultMessage(a.GetID(), result, msg.ID)
	response.SetMetadata("tool_name", toolNameStr)

	// 记录工具调用结果到记忆
	if a.GetMemoryManager() != nil {
		toolResultKey := memory.CreateTaggedKey(a.GetID(), "tool_results", msg.ID)
		if err := a.SaveMemory(ctx, toolResultKey, result); err != nil {
			hlog.CtxWarnf(ctx, "保存工具调用结果到记忆失败: %v", err)
		}
	}

	return response, nil
}

// Shutdown 实现Agent接口，关闭智能体
func (a *GenericAdvancedAgent) Shutdown(ctx context.Context) error {
	hlog.CtxInfof(ctx, "关闭通用高级智能体: ID=%s, Type=%s", a.GetID(), a.GetType())

	// 保存最终状态到记忆
	if a.GetMemoryManager() != nil {
		stateKey := memory.CreateTaggedKey(a.GetID(), "state", "shutdown_time")
		if err := a.SaveMemory(ctx, stateKey, time.Now().Format(time.RFC3339)); err != nil {
			hlog.CtxWarnf(ctx, "保存关闭时间到记忆失败: %v", err)
		}
	}

	// 调用基础关闭
	return a.BaseAdvancedAgent.Shutdown(ctx)
}

// UseMemory 从记忆中查找相关信息
// category: 记忆类别
// prefix: 键前缀
// 返回: 找到的记忆映射
func (a *GenericAdvancedAgent) UseMemory(ctx context.Context, category string, prefix string) (map[string]interface{}, error) {
	if a.GetMemoryManager() == nil {
		return nil, nil
	}

	// 创建搜索前缀
	searchPrefix := memory.CreateTaggedKey(a.GetID(), category, prefix)

	// 列出匹配的键
	keys, err := a.GetMemoryManager().List(ctx, searchPrefix)
	if err != nil {
		return nil, err
	}

	// 加载所有匹配的记忆
	memories := make(map[string]interface{})
	for _, key := range keys {
		_, _, shortKey := memory.ExtractKeyParts(key)
		value, err := a.LoadMemory(ctx, key)
		if err != nil {
			hlog.CtxWarnf(ctx, "加载记忆失败 %s: %v", key, err)
			continue
		}
		memories[shortKey] = value
	}

	return memories, nil
}
