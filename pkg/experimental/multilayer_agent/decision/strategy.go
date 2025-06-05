package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/llms"

	"novelai/pkg/experimental/multilayer_agent/core"
	"novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// StrategyAgent 策略智能体
// 负责制定高层次策略决策，协调其他智能体的工作
type StrategyAgent struct {
	*core.BaseAdvancedAgent           // 嵌入基础高级智能体实现
	prompt                  string    // 策略提示模板
	maxRetries              int       // 最大重试次数
	lastProcessTime         time.Time // 上次处理时间
	strategyHistory         []string  // 策略历史记录
	currentStrategy         string    // 当前活跃策略
}

// NewStrategyAgent 创建新的策略智能体
func NewStrategyAgent(id string, prompt string) *StrategyAgent {
	agent := &StrategyAgent{
		BaseAdvancedAgent: core.NewBaseAdvancedAgent(id, core.AgentTypeStrategy),
		prompt:            prompt,
		maxRetries:        3,
		strategyHistory:   make([]string, 0),
	}
	return agent
}

// Initialize 实现Agent接口，进行初始化
func (a *StrategyAgent) Initialize(ctx context.Context) error {
	hlog.CtxInfof(ctx, "初始化策略智能体: ID=%s", a.GetID())

	// 调用基础初始化
	if err := a.BaseAdvancedAgent.Initialize(ctx); err != nil {
		return err
	}

	// 加载之前的状态（如果有）
	if a.GetMemoryManager() != nil {
		// 尝试加载上次处理时间
		stateKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyLastProcessTime)
		if val, err := a.LoadMemory(ctx, stateKey); err == nil && val != nil {
			if timeStr, ok := val.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
					a.lastProcessTime = parsedTime
					hlog.CtxInfof(ctx, "从记忆加载上次处理时间: %v", a.lastProcessTime)
				}
			}
		}

		// 尝试加载当前策略
		strategyKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy)
		if val, err := a.LoadMemory(ctx, strategyKey); err == nil && val != nil {
			if strategy, ok := val.(string); ok {
				a.currentStrategy = strategy
				hlog.CtxInfof(ctx, "从记忆加载当前策略: %s", a.currentStrategy)
			}
		}

		// 尝试加载策略历史
		historyKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory)
		if val, err := a.LoadMemory(ctx, historyKey); err == nil && val != nil {
			if history, ok := val.([]string); ok {
				a.strategyHistory = history
				hlog.CtxInfof(ctx, "从记忆加载策略历史: %d 条记录", len(a.strategyHistory))
			}
		}
	}

	return nil
}

// Process 实现Agent接口，处理消息
func (a *StrategyAgent) Process(ctx context.Context, msg *core.Message) (*core.Message, error) {
	a.SetStatus(core.AgentStatusWorking)
	defer a.SetStatus(core.AgentStatusIdle)

	hlog.CtxInfof(ctx, "策略智能体处理消息: ID=%s, Type=%s, Subject=%s",
		msg.ID, msg.Type, msg.Subject)

	// 记录处理时间
	now := time.Now()
	a.lastProcessTime = now

	// 记录到记忆
	if a.GetMemoryManager() != nil {
		stateKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyLastProcessTime)
		if err := a.SaveMemory(ctx, stateKey, now.Format(time.RFC3339)); err != nil {
			hlog.CtxWarnf(ctx, "保存处理时间到记忆失败: %v", err)
		}

		// 保存消息历史
		historyKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryHistory, msg.ID)
		if err := a.SaveMemory(ctx, historyKey, msg); err != nil {
			hlog.CtxWarnf(ctx, "保存消息历史到记忆失败: %v", err)
		}
	}

	// 检查模型是否已设置
	if a.GetModel() == nil {
		hlog.CtxErrorf(ctx, "未设置语言模型，策略智能体无法处理消息")
		return nil, fmt.Errorf("未设置语言模型，策略智能体无法处理消息")
	}

	// 构建模型输入
	promptTemplate := a.prompt
	if promptTemplate == "" {
		// 默认提示模板，专注于策略制定
		promptTemplate = `
		你是策略智能体，负责在多层智能体系统中制定高层次策略决策。
		你的职责是分析用户需求和系统状态，制定清晰的整体策略方向，并协调其他智能体的工作。

		当前策略历史:
		%s

		当前活跃策略:
		%s

		消息类型: %s
		消息来源: %s
		消息主题: %s

		消息内容:
		%s

		请分析以上信息，并做出以下类型的决策之一:
		1. 制定新策略 - 格式: {"action":"new_strategy","strategy":"策略内容"}
		2. 修改当前策略 - 格式: {"action":"update_strategy","strategy":"策略内容"}
		3. 将任务分配给其他智能体 - 格式: {"action":"delegate","agent_id":"目标智能体ID","message":"任务内容"}
		4. 直接回复用户 - 格式: {"action":"reply","message":"回复内容"}`
	}

	// 生成策略历史摘要文本
	historyText := "无"
	if len(a.strategyHistory) > 0 {
		historyText = ""
		// 最多展示最近的5条策略历史
		startIdx := 0
		if len(a.strategyHistory) > 5 {
			startIdx = len(a.strategyHistory) - 5
		}
		for i := startIdx; i < len(a.strategyHistory); i++ {
			historyText += fmt.Sprintf("%d. %s\n", i+1, a.strategyHistory[i])
		}
	}

	// 当前策略文本
	currentStrategyText := "无"
	if a.currentStrategy != "" {
		currentStrategyText = a.currentStrategy
	}

	// 生成提示，包含完整的消息上下文
	prompt := fmt.Sprintf(promptTemplate,
		historyText,
		currentStrategyText,
		string(msg.Type),
		msg.From,
		msg.Subject,
		msg.Content)

	// 调用模型生成内容
	hlog.CtxInfof(ctx, "策略智能体调用模型处理消息：%s (模型：%s)", msg.Subject, a.GetModel().ModelName())

	var modelResponse string
	var err error

	// 策略智能体优先使用JSON模式，便于结构化输出
	if a.GetModel().SupportsJSON() {
		// 使用JSON模式
		messages := []llms.MessageContent{
			{
				Role: "system",
				Parts: []llms.ContentPart{
					llms.TextPart("你是策略智能体，负责制定高层次决策。请以JSON格式回复。"),
				},
			},
			{
				Role: "user",
				Parts: []llms.ContentPart{
					llms.TextPart(prompt),
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

	// 解析模型响应，处理不同类型的决策
	var strategyAction struct {
		Action   string `json:"action"`
		Strategy string `json:"strategy,omitempty"`
		AgentID  string `json:"agent_id,omitempty"`
		Message  string `json:"message,omitempty"`
	}

	if err := json.Unmarshal([]byte(modelResponse), &strategyAction); err != nil {
		hlog.CtxWarnf(ctx, "解析模型响应失败: %v，尝试作为直接回复处理", err)
		// 如果解析失败，默认作为直接回复处理
		strategyAction.Action = "reply"
		strategyAction.Message = modelResponse
	}

	// 根据不同的决策类型处理响应
	switch strategyAction.Action {
	case memory.ActionTypeNewStrategy, memory.ActionTypeUpdateStrategy:
		// 更新或创建新策略
		a.currentStrategy = strategyAction.Strategy
		a.strategyHistory = append(a.strategyHistory, strategyAction.Strategy)

		// 保存到记忆
		if a.GetMemoryManager() != nil {
			if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy), a.currentStrategy); err != nil {
				hlog.CtxWarnf(ctx, "保存当前策略到记忆失败: %v", err)
			}
			if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory), a.strategyHistory); err != nil {
				hlog.CtxWarnf(ctx, "保存策略历史到记忆失败: %v", err)
			}
		}

		// 创建策略更新消息
		response := core.NewMessage(core.MessageTypeResponse, a.GetID(), msg.From)
		response.Subject = "策略更新"
		var actionText string
		if strategyAction.Action == "new_strategy" {
			actionText = "制定新"
		} else {
			actionText = "更新"
		}
		response.Content = fmt.Sprintf("已%s策略：\n%s", actionText, strategyAction.Strategy)
		response.ReplyTo = msg.ID

		// 添加处理元数据
		response.SetMetadata(memory.MetadataKeyProcessTime, time.Since(now).String())
		response.SetMetadata(memory.MetadataKeyAgentType, string(a.GetType()))
		response.SetMetadata("strategy_action", strategyAction.Action)

		return response, nil

	case memory.ActionTypeDelegate:
		// 将任务分配给其他智能体
		if strategyAction.AgentID == "" {
			hlog.CtxErrorf(ctx, "委派任务缺少目标智能体ID")
			return core.CreateErrorMessage(a.GetID(), fmt.Errorf("委派任务缺少目标智能体ID"), msg.ID), nil
		}

		// 创建发送给其他智能体的消息
		response := core.NewMessage(core.MessageTypeRequest, a.GetID(), strategyAction.AgentID)
		response.Subject = fmt.Sprintf("策略委派任务: %s", msg.Subject)
		response.Content = strategyAction.Message
		response.ReplyTo = msg.ID
		response.SetMetadata(memory.MetadataKeyOriginalFrom, msg.From)
		response.SetMetadata(memory.MetadataKeyProcessTime, time.Since(now).String())
		response.SetMetadata("strategy_context", a.currentStrategy)

		return response, nil

	case memory.ActionTypeReply, "":
		// 直接回复用户
		response := core.NewMessage(core.MessageTypeResponse, a.GetID(), msg.From)
		response.Subject = "策略回复: " + msg.Subject
		response.Content = strategyAction.Message
		if strategyAction.Message == "" {
			response.Content = modelResponse // 如果消息为空，使用原始响应
		}
		response.ReplyTo = msg.ID

		// 添加处理元数据
		response.SetMetadata(memory.MetadataKeyProcessTime, time.Since(now).String())
		response.SetMetadata(memory.MetadataKeyAgentType, string(a.GetType()))
		response.SetMetadata(memory.MetadataKeyModelName, a.GetModel().ModelName())
		response.SetMetadata("current_strategy", a.currentStrategy)

		return response, nil

	default:
		// 未知操作类型
		hlog.CtxErrorf(ctx, "未知的策略操作类型: %s", strategyAction.Action)
		return core.CreateErrorMessage(a.GetID(), fmt.Errorf("未知的策略操作类型: %s", strategyAction.Action), msg.ID), nil
	}
}

// UseStrategyHistory 获取策略历史记录
func (a *StrategyAgent) UseStrategyHistory(ctx context.Context) []string {
	return a.strategyHistory
}

// GetCurrentStrategy 获取当前活跃策略
func (a *StrategyAgent) GetCurrentStrategy() string {
	return a.currentStrategy
}

// SetCurrentStrategy 设置当前活跃策略
func (a *StrategyAgent) SetCurrentStrategy(ctx context.Context, strategy string) error {
	a.currentStrategy = strategy

	// 添加到历史记录
	a.strategyHistory = append(a.strategyHistory, strategy)

	// 保存到记忆
	if a.GetMemoryManager() != nil {
		if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy), a.currentStrategy); err != nil {
			return err
		}
		if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory), a.strategyHistory); err != nil {
			return err
		}
	}

	return nil
}

// Shutdown 实现Agent接口，关闭智能体
func (a *StrategyAgent) Shutdown(ctx context.Context) error {
	hlog.CtxInfof(ctx, "关闭策略智能体: ID=%s", a.GetID())

	// 保存最终状态到记忆
	if a.GetMemoryManager() != nil {
		stateKey := memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, "shutdown_time")
		if err := a.SaveMemory(ctx, stateKey, time.Now().Format(time.RFC3339)); err != nil {
			hlog.CtxWarnf(ctx, "保存关闭时间到记忆失败: %v", err)
		}

		// 确保策略状态被保存
		if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy), a.currentStrategy); err != nil {
			hlog.CtxWarnf(ctx, "保存当前策略到记忆失败: %v", err)
		}
		if err := a.SaveMemory(ctx, memory.CreateTaggedKey(a.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory), a.strategyHistory); err != nil {
			hlog.CtxWarnf(ctx, "保存策略历史到记忆失败: %v", err)
		}
	}

	// 调用基础关闭
	return a.BaseAdvancedAgent.Shutdown(ctx)
}
