package memory

// 记忆系统常量定义
// 这些常量用于标准化内存键的创建和使用

// 记忆类别常量
const (
	// MemoryCategoryState 状态类别，用于存储智能体运行状态
	MemoryCategoryState = "state"

	// MemoryCategoryHistory 历史类别，用于存储消息历史
	MemoryCategoryHistory = "history"

	// MemoryCategoryContext 上下文类别，用于存储会话上下文
	MemoryCategoryContext = "context"

	// MemoryCategoryKnowledge 知识类别，用于存储智能体获取的知识
	MemoryCategoryKnowledge = "knowledge"
)

// 状态记忆键常量
const (
	// StateKeyLastProcessTime 上次处理时间的键
	StateKeyLastProcessTime = "last_process_time"

	// StateKeyCurrentStrategy 当前策略的键
	StateKeyCurrentStrategy = "current_strategy"

	// StateKeyStrategyHistory 策略历史的键
	StateKeyStrategyHistory = "strategy_history"

	// StateKeyAgentStatus 智能体状态的键
	StateKeyAgentStatus = "agent_status"
)

// 元数据键常量
const (
	// MetadataKeyAgentType 智能体类型元数据键
	MetadataKeyAgentType = "agent_type"

	// MetadataKeyProcessTime 处理时间元数据键
	MetadataKeyProcessTime = "process_time"

	// MetadataKeyModelName 模型名称元数据键
	MetadataKeyModelName = "model_name"

	// MetadataKeyModelType 模型类型元数据键
	MetadataKeyModelType = "model_type"

	// MetadataKeyOriginalFrom 原始发送者元数据键
	MetadataKeyOriginalFrom = "original_from"
)

// 工具相关常量
const (
	// ToolNameKey 工具名称键
	ToolNameKey = "tool_name"

	// ToolInputKey 工具输入键
	ToolInputKey = "input"

	// ToolResultKey 工具结果键
	ToolResultKey = "result"
)

// 消息交互常量
const (
	// ActionKeySendTo 发送消息目标键
	ActionKeySendTo = "send_to"

	// ActionKeyMessage 消息内容键
	ActionKeyMessage = "message"

	// ActionKeyStrategy 策略内容键
	ActionKeyStrategy = "strategy"

	// ActionKeyAction 动作类型键
	ActionKeyAction = "action"

	// ActionKeyAgentID 智能体ID键
	ActionKeyAgentID = "agent_id"
)

// 动作类型常量
const (
	// ActionTypeNewStrategy 新建策略动作
	ActionTypeNewStrategy = "new_strategy"

	// ActionTypeUpdateStrategy 更新策略动作
	ActionTypeUpdateStrategy = "update_strategy"

	// ActionTypeDelegate 任务委派动作
	ActionTypeDelegate = "delegate"

	// ActionTypeReply 直接回复动作
	ActionTypeReply = "reply"
)
