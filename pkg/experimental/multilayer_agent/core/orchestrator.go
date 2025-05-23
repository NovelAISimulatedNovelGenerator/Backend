package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"novelai/pkg/experimental/multilayer_agent/shared/model"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// OrchestratorConfig 编排器配置
type OrchestratorConfig struct {
	MaxConcurrentAgents int             // 最大并发智能体数
	MessageQueueSize    int             // 消息队列大小
	ProcessTimeout      time.Duration   // 处理超时时间
	EnableMetrics       bool            // 是否启用指标收集
	DefaultModelType    model.ModelType // 默认模型类型
	DefaultModelName    string          // 默认模型名称
}

// DefaultOrchestratorConfig 返回默认配置
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		MaxConcurrentAgents: 10,
		MessageQueueSize:    1000,
		ProcessTimeout:      30 * time.Second,
		EnableMetrics:       true,
		DefaultModelType:    model.ModelTypeOllama,
		DefaultModelName:    "mistral",
	}
}

// Orchestrator 智能体编排器
// 负责管理和协调多个智能体的工作
type Orchestrator struct {
	config       *OrchestratorConfig    // 配置
	agents       map[string]Agent       // 注册的智能体
	agentMutex   sync.RWMutex           // 智能体映射的读写锁
	messageQueue chan *MessageEnvelope  // 消息队列
	routingTable map[AgentType][]string // 路由表：智能体类型到ID的映射
	routingMutex sync.RWMutex           // 路由表的读写锁
	ctx          context.Context        // 上下文
	cancel       context.CancelFunc     // 取消函数
	wg           sync.WaitGroup         // 等待组
	running      bool                   // 运行状态
	runningMutex sync.RWMutex           // 运行状态的读写锁
	modelFactory model.ModelFactory     // 模型工厂
}

// MessageEnvelope 消息信封
// 包装消息和相关的处理信息
type MessageEnvelope struct {
	Message    *Message                   // 消息本体
	ResponseCh chan *MessageProcessResult // 响应通道
}

// MessageProcessResult 消息处理结果
type MessageProcessResult struct {
	Message *Message // 响应消息
	Error   error    // 处理错误
}

// NewOrchestrator 创建新的编排器
func NewOrchestrator(config *OrchestratorConfig) *Orchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	orchestrator := &Orchestrator{
		config:       config,
		agents:       make(map[string]Agent),
		messageQueue: make(chan *MessageEnvelope, config.MessageQueueSize),
		routingTable: make(map[AgentType][]string),
		ctx:          ctx,
		cancel:       cancel,
		running:      false,
		modelFactory: model.NewModelFactory(),
	}

	return orchestrator
}

// RegisterAgent 注册智能体
func (o *Orchestrator) RegisterAgent(agent Agent) error {
	if agent == nil {
		return errors.New("智能体不能为空")
	}

	agentID := agent.GetID()
	agentType := agent.GetType()

	// 如果智能体没有设置模型，使用默认模型
	if agent.GetModel() == nil {
		hlog.Infof("为智能体 %s 设置默认模型 %s:%s", agentID, o.config.DefaultModelType, o.config.DefaultModelName)
		defaultModel, err := o.modelFactory.CreateModel(
			o.config.DefaultModelType,
			model.ModelOptions{ModelName: o.config.DefaultModelName},
		)
		if err != nil {
			return fmt.Errorf("创建默认模型失败: %w", err)
		}
		agent.SetModel(defaultModel)
	}

	o.agentMutex.Lock()
	defer o.agentMutex.Unlock()

	// 检查是否已存在相同ID的智能体
	if _, exists := o.agents[agentID]; exists {
		return fmt.Errorf("已存在ID为 %s 的智能体", agentID)
	}

	// 注册智能体
	o.agents[agentID] = agent

	// 更新路由表
	o.routingMutex.Lock()
	defer o.routingMutex.Unlock()

	if _, exists := o.routingTable[agentType]; !exists {
		o.routingTable[agentType] = []string{}
	}

	o.routingTable[agentType] = append(o.routingTable[agentType], agentID)

	hlog.Infof("已注册智能体: ID=%s, 类型=%s", agentID, agentType)

	return nil
}

// UnregisterAgent 注销智能体
func (o *Orchestrator) UnregisterAgent(agentID string) error {
	o.agentMutex.Lock()
	defer o.agentMutex.Unlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return fmt.Errorf("智能体不存在: %s", agentID)
	}

	// 从路由表中移除
	o.routingMutex.Lock()
	agentType := agent.GetType()
	agentIDs := o.routingTable[agentType]
	for i, id := range agentIDs {
		if id == agentID {
			o.routingTable[agentType] = append(agentIDs[:i], agentIDs[i+1:]...)
			break
		}
	}
	o.routingMutex.Unlock()

	// 移除智能体
	delete(o.agents, agentID)

	hlog.Infof("注销智能体成功: ID=%s", agentID)
	return nil
}

// Start 启动编排器
func (o *Orchestrator) Start() error {
	o.runningMutex.Lock()
	if o.running {
		o.runningMutex.Unlock()
		return errors.New("编排器已经在运行")
	}
	o.running = true
	o.runningMutex.Unlock()

	// 初始化所有智能体
	o.agentMutex.RLock()
	agents := make([]Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	o.agentMutex.RUnlock()

	for _, agent := range agents {
		if err := agent.Initialize(o.ctx); err != nil {
			return fmt.Errorf("初始化智能体失败 %s: %w", agent.GetID(), err)
		}
	}

	// 启动消息处理工作池
	for i := 0; i < o.config.MaxConcurrentAgents; i++ {
		o.wg.Add(1)
		go o.messageProcessor(i)
	}

	hlog.Info("编排器启动成功")
	return nil
}

// Stop 停止编排器
func (o *Orchestrator) Stop() error {
	o.runningMutex.Lock()
	if !o.running {
		o.runningMutex.Unlock()
		return errors.New("编排器未在运行")
	}
	o.running = false
	o.runningMutex.Unlock()

	// 发送取消信号
	o.cancel()

	// 关闭消息队列
	close(o.messageQueue)

	// 等待所有工作协程结束
	o.wg.Wait()

	// 关闭所有智能体
	o.agentMutex.RLock()
	agents := make([]Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	o.agentMutex.RUnlock()

	for _, agent := range agents {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := agent.Shutdown(shutdownCtx); err != nil {
			hlog.Errorf("关闭智能体失败 %s: %v", agent.GetID(), err)
		}
		cancel()
	}

	hlog.Info("编排器停止成功")
	return nil
}

// SendMessage 发送消息到指定智能体
func (o *Orchestrator) SendMessage(ctx context.Context, msg *Message) (*Message, error) {
	o.runningMutex.RLock()
	if !o.running {
		o.runningMutex.RUnlock()
		return nil, errors.New("编排器未在运行")
	}
	o.runningMutex.RUnlock()

	// 创建消息信封
	envelope := &MessageEnvelope{
		Message:    msg,
		ResponseCh: make(chan *MessageProcessResult, 1),
	}

	// 发送到消息队列
	select {
	case o.messageQueue <- envelope:
		// 等待响应
		select {
		case result := <-envelope.ResponseCh:
			if result.Error != nil {
				return nil, result.Error
			}
			return result.Message, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// BroadcastMessage 广播消息到指定类型的所有智能体
func (o *Orchestrator) BroadcastMessage(ctx context.Context, agentType AgentType, msg *Message) ([]*Message, error) {
	o.routingMutex.RLock()
	agentIDs := o.routingTable[agentType]
	o.routingMutex.RUnlock()

	if len(agentIDs) == 0 {
		return nil, fmt.Errorf("没有找到类型为 %s 的智能体", agentType)
	}

	// 并发发送消息
	var wg sync.WaitGroup
	responses := make([]*Message, 0, len(agentIDs))
	responseMutex := sync.Mutex{}
	errors := make([]error, 0)

	for _, agentID := range agentIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// 克隆消息并设置接收方
			msgCopy := msg.Clone()
			msgCopy.To = id

			resp, err := o.SendMessage(ctx, msgCopy)

			responseMutex.Lock()
			if err != nil {
				errors = append(errors, err)
			} else {
				responses = append(responses, resp)
			}
			responseMutex.Unlock()
		}(agentID)
	}

	wg.Wait()

	if len(errors) > 0 {
		return responses, fmt.Errorf("部分消息发送失败: %v", errors)
	}

	return responses, nil
}

// messageProcessor 消息处理器
func (o *Orchestrator) messageProcessor(id int) {
	defer o.wg.Done()

	hlog.Infof("消息处理器 %d 启动", id)

	for envelope := range o.messageQueue {
		o.processMessage(envelope)
	}

	hlog.Infof("消息处理器 %d 停止", id)
}

// processMessage 处理单个消息
func (o *Orchestrator) processMessage(envelope *MessageEnvelope) {
	msg := envelope.Message

	// 查找目标智能体
	o.agentMutex.RLock()
	agent, exists := o.agents[msg.To]
	o.agentMutex.RUnlock()

	if !exists {
		envelope.ResponseCh <- &MessageProcessResult{
			Error: fmt.Errorf("目标智能体不存在: %s", msg.To),
		}
		return
	}

	// 创建处理上下文
	processCtx, cancel := context.WithTimeout(o.ctx, o.config.ProcessTimeout)
	defer cancel()

	// 记录处理开始
	startTime := time.Now()
	hlog.Infof("开始处理消息: ID=%s, From=%s, To=%s, Type=%s",
		msg.ID, msg.From, msg.To, msg.Type)

	// 调用智能体处理消息
	response, err := agent.Process(processCtx, msg)

	// 记录处理结果
	duration := time.Since(startTime)
	if err != nil {
		hlog.Errorf("处理消息失败: ID=%s, Error=%v, Duration=%v",
			msg.ID, err, duration)
		envelope.ResponseCh <- &MessageProcessResult{
			Error: err,
		}
	} else {
		hlog.Infof("处理消息成功: ID=%s, Duration=%v", msg.ID, duration)
		envelope.ResponseCh <- &MessageProcessResult{
			Message: response,
		}
	}
}

// GetAgent 获取指定ID的智能体
func (o *Orchestrator) GetAgent(agentID string) (Agent, bool) {
	o.agentMutex.RLock()
	defer o.agentMutex.RUnlock()

	agent, exists := o.agents[agentID]
	return agent, exists
}

// GetAgentsByType 获取指定类型的所有智能体
func (o *Orchestrator) GetAgentsByType(agentType AgentType) []Agent {
	o.routingMutex.RLock()
	agentIDs := o.routingTable[agentType]
	o.routingMutex.RUnlock()

	o.agentMutex.RLock()
	defer o.agentMutex.RUnlock()

	agents := make([]Agent, 0, len(agentIDs))
	for _, id := range agentIDs {
		if agent, exists := o.agents[id]; exists {
			agents = append(agents, agent)
		}
	}

	return agents
}

// GetStatus 获取编排器状态
func (o *Orchestrator) GetStatus() map[string]interface{} {
	o.runningMutex.RLock()
	running := o.running
	o.runningMutex.RUnlock()

	o.agentMutex.RLock()
	agentCount := len(o.agents)
	o.agentMutex.RUnlock()

	status := map[string]interface{}{
		"running":        running,
		"agent_count":    agentCount,
		"queue_size":     len(o.messageQueue),
		"queue_capacity": o.config.MessageQueueSize,
	}

	// 统计各类型智能体数量
	o.routingMutex.RLock()
	agentTypeCount := make(map[string]int)
	for agentType, ids := range o.routingTable {
		agentTypeCount[string(agentType)] = len(ids)
	}
	o.routingMutex.RUnlock()

	status["agent_types"] = agentTypeCount

	return status
}
