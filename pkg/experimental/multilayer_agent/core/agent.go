package core

import (
	"context"

	"novelai/pkg/experimental/multilayer_agent/shared/memory"
	"novelai/pkg/experimental/multilayer_agent/shared/model"

	"github.com/tmc/langchaingo/tools"
)

// AgentType 定义智能体类型
type AgentType string

const (
	// 决策层智能体类型
	AgentTypeStrategy  AgentType = "strategy"
	AgentTypePlanner   AgentType = "planner"
	AgentTypeEvaluator AgentType = "evaluator"

	// 执行层智能体类型
	AgentTypeWorldview  AgentType = "worldview"
	AgentTypeCharacter  AgentType = "character"
	AgentTypePlot       AgentType = "plot"
	AgentTypeDialogue   AgentType = "dialogue"
	AgentTypeBackground AgentType = "background"
	AgentTypeFormatter  AgentType = "formatter"
)

// AgentStatus 定义智能体状态
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"    // 空闲状态
	AgentStatusWorking AgentStatus = "working" // 工作状态
	AgentStatusError   AgentStatus = "error"   // 错误状态
)

// Agent 定义所有智能体必须实现的基本接口
// 这是整个多层智能体系统的核心接口
type Agent interface {
	// GetID 获取智能体唯一标识
	GetID() string

	// GetType 获取智能体类型
	GetType() AgentType

	// GetStatus 获取智能体当前状态
	GetStatus() AgentStatus

	// Process 处理消息并返回响应
	// ctx: 上下文，用于控制超时和取消
	// msg: 输入消息
	// 返回：响应消息或错误
	Process(ctx context.Context, msg *Message) (*Message, error)

	// Initialize 初始化智能体
	// 在智能体开始工作前调用
	Initialize(ctx context.Context) error

	// Shutdown 关闭智能体
	// 在智能体停止工作前调用
	Shutdown(ctx context.Context) error

	// GetModel 获取智能体使用的语言模型
	GetModel() model.Model

	// SetModel 设置智能体使用的语言模型
	SetModel(m model.Model)
}

// ToolEnabledAgent 支持工具调用的智能体接口
// 扩展了基础Agent接口，增加工具调用能力
type ToolEnabledAgent interface {
	Agent

	// SetToolCaller 设置工具调用器
	SetToolCaller(caller ToolCaller)

	// GetAvailableTools 获取智能体可用的工具列表
	GetAvailableTools() []tools.Tool
}

// MemoryEnabledAgent 支持记忆功能的智能体接口
// 扩展了基础Agent接口，增加记忆管理能力
type MemoryEnabledAgent interface {
	Agent

	// SetMemoryManager 设置记忆管理器
	SetMemoryManager(manager memory.Manager)

	// SaveMemory 保存记忆
	SaveMemory(ctx context.Context, key string, value interface{}) error

	// LoadMemory 加载记忆
	LoadMemory(ctx context.Context, key string) (interface{}, error)
}

// AdvancedAgent 高级智能体接口
// 同时支持工具调用和记忆功能
type AdvancedAgent interface {
	Agent

	// 工具相关功能
	SetToolCaller(caller ToolCaller)
	GetAvailableTools() []tools.Tool

	// 记忆相关功能
	SetMemoryManager(manager memory.Manager)
	SaveMemory(ctx context.Context, key string, value interface{}) error
	LoadMemory(ctx context.Context, key string) (interface{}, error)
}

// BaseAgent 智能体基础实现
// 提供通用功能的默认实现
type BaseAgent struct {
	id            string         // 智能体唯一标识
	agentType     AgentType      // 智能体类型
	status        AgentStatus    // 当前状态
	toolCaller    ToolCaller     // 工具调用器（可选）
	memoryManager memory.Manager // 记忆管理器（可选）
	llmModel      model.Model    // 语言模型
}

// NewBaseAgent 创建基础智能体
func NewBaseAgent(id string, agentType AgentType) *BaseAgent {
	return &BaseAgent{
		id:        id,
		agentType: agentType,
		status:    AgentStatusIdle,
	}
}

// GetID 实现Agent接口
func (a *BaseAgent) GetID() string {
	return a.id
}

// GetType 实现Agent接口
func (a *BaseAgent) GetType() AgentType {
	return a.agentType
}

// GetStatus 实现Agent接口
func (a *BaseAgent) GetStatus() AgentStatus {
	return a.status
}

// SetStatus 设置智能体状态
func (a *BaseAgent) SetStatus(status AgentStatus) {
	a.status = status
}

// SetToolCaller 实现ToolEnabledAgent接口
func (a *BaseAgent) SetToolCaller(caller ToolCaller) {
	a.toolCaller = caller
}

// GetToolCaller 获取工具调用器
func (a *BaseAgent) GetToolCaller() ToolCaller {
	return a.toolCaller
}

// SetMemoryManager 实现MemoryEnabledAgent接口
func (a *BaseAgent) SetMemoryManager(manager memory.Manager) {
	a.memoryManager = manager
}

// GetMemoryManager 获取记忆管理器
func (a *BaseAgent) GetMemoryManager() memory.Manager {
	return a.memoryManager
}

// SaveMemory 实现MemoryEnabledAgent接口
func (a *BaseAgent) SaveMemory(ctx context.Context, key string, value interface{}) error {
	if a.memoryManager == nil {
		return nil // 没有记忆管理器时静默忽略
	}
	return a.memoryManager.Save(ctx, key, value)
}

// LoadMemory 实现MemoryEnabledAgent接口
func (a *BaseAgent) LoadMemory(ctx context.Context, key string) (interface{}, error) {
	if a.memoryManager == nil {
		return nil, nil // 没有记忆管理器时返回nil
	}
	return a.memoryManager.Load(ctx, key)
}

// GetModel 获取智能体使用的语言模型
func (a *BaseAgent) GetModel() model.Model {
	return a.llmModel
}

// SetModel 设置智能体使用的语言模型
func (a *BaseAgent) SetModel(m model.Model) {
	a.llmModel = m
}

// Initialize 默认初始化实现
func (a *BaseAgent) Initialize(ctx context.Context) error {
	a.status = AgentStatusIdle
	return nil
}

// Shutdown 默认关闭实现
func (a *BaseAgent) Shutdown(ctx context.Context) error {
	a.status = AgentStatusIdle
	return nil
}

// ToolCaller 工具调用器接口
// 用于处理工具调用请求
type ToolCaller interface {
	// Call 调用工具
	// toolName: 工具名称
	// input: 工具输入参数
	// 返回：工具输出结果或错误
	Call(ctx context.Context, toolName string, input string) (string, error)

	// GetAvailableTools 获取所有可用工具
	GetAvailableTools() []tools.Tool
}

// BaseAdvancedAgent 高级智能体基础实现
// 同时支持工具调用和记忆功能
type BaseAdvancedAgent struct {
	BaseAgent                   // 嵌入基础智能体实现
	availableTools []tools.Tool // 可用工具列表
}

// NewBaseAdvancedAgent 创建基础高级智能体
func NewBaseAdvancedAgent(id string, agentType AgentType) *BaseAdvancedAgent {
	return &BaseAdvancedAgent{
		BaseAgent: *NewBaseAgent(id, agentType),
	}
}

// GetAvailableTools 实现ToolEnabledAgent接口
func (a *BaseAdvancedAgent) GetAvailableTools() []tools.Tool {
	if a.availableTools != nil {
		return a.availableTools
	}

	if a.toolCaller != nil {
		return a.toolCaller.GetAvailableTools()
	}

	return []tools.Tool{}
}

// SetAvailableTools 设置可用工具列表
func (a *BaseAdvancedAgent) SetAvailableTools(tools []tools.Tool) {
	a.availableTools = tools
}

// CallTool 调用工具的辅助方法
func (a *BaseAdvancedAgent) CallTool(ctx context.Context, toolName string, input string) (string, error) {
	if a.toolCaller == nil {
		return "", nil // 没有工具调用器时返回空字符串
	}
	return a.toolCaller.Call(ctx, toolName, input)
}
