// package test 实现编排器相关的测试用例
package test

import (
	"context"
	"fmt"
	"time"

	"novelai/pkg/experimental/multilayer_agent/core"
	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// OrchestratorLifecycleTest 测试编排器生命周期管理
type OrchestratorLifecycleTest struct {
	model        model.Model
	orchestrator *core.Orchestrator
}

// NewOrchestratorLifecycleTest 创建编排器生命周期测试用例
func NewOrchestratorLifecycleTest() *OrchestratorLifecycleTest {
	return &OrchestratorLifecycleTest{}
}

// Name 返回测试名称
func (t *OrchestratorLifecycleTest) Name() string {
	return "编排器生命周期测试"
}

// Description 返回测试描述
func (t *OrchestratorLifecycleTest) Description() string {
	return "测试编排器的启动、停止和状态管理功能"
}

// Setup 设置测试环境
func (t *OrchestratorLifecycleTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model

	// 创建编排器配置
	config := core.DefaultOrchestratorConfig()
	// 为测试调整配置
	config.MaxConcurrentAgents = 5
	config.MessageQueueSize = 10
	config.ProcessTimeout = 10 * time.Second

	// 创建编排器
	t.orchestrator = core.NewOrchestrator(config)

	return nil
}

// Execute 执行测试
func (t *OrchestratorLifecycleTest) Execute() error {
	hlog.Infof("开始执行编排器生命周期测试...")

	// 测试启动编排器
	hlog.Infof("启动编排器...")
	err := t.orchestrator.Start()
	if err != nil {
		return fmt.Errorf("启动编排器失败: %v", err)
	}

	// 检查编排器状态
	status := t.orchestrator.GetStatus()
	running, ok := status["running"].(bool)
	if !ok || !running {
		return fmt.Errorf("编排器启动后状态错误: 预期运行中，实际未运行")
	}

	// 测试停止编排器
	hlog.Infof("停止编排器...")
	t.orchestrator.Stop()

	// 检查编排器状态
	status = t.orchestrator.GetStatus()
	running, ok = status["running"].(bool)
	if ok && running {
		return fmt.Errorf("编排器停止后状态错误: 预期已停止，实际仍在运行")
	}

	// 测试重新启动编排器
	hlog.Infof("重新启动编排器...")
	err = t.orchestrator.Start()
	if err != nil {
		return fmt.Errorf("重新启动编排器失败: %v", err)
	}

	// 再次检查编排器状态
	status = t.orchestrator.GetStatus()
	running, ok = status["running"].(bool)
	if !ok || !running {
		return fmt.Errorf("编排器重新启动后状态错误: 预期运行中，实际未运行")
	}

	return nil
}

// Verify 验证测试结果
func (t *OrchestratorLifecycleTest) Verify() (bool, string) {
	// 最终检查编排器状态
	status := t.orchestrator.GetStatus()
	
	running, ok := status["running"].(bool)
	if !ok || !running {
		return false, "编排器未处于运行状态"
	}

	return true, ""
}

// Cleanup 清理测试资源
func (t *OrchestratorLifecycleTest) Cleanup() {
	// 仅当编排器存在时尝试清理
	if t.orchestrator != nil {
		// 使用GetStatus获取编排器运行状态
		status := t.orchestrator.GetStatus()
		isRunning, ok := status["running"].(bool)
		
		// 仅当编排器实际运行中时才调用Stop
		if ok && isRunning {
			// 尝试停止编排器，忽略可能的错误
			// 因为在测试清理阶段这些错误通常不重要
			err := t.orchestrator.Stop()
			if err != nil {
				// 仅记录错误，不中断清理过程
				hlog.Warnf("清理编排器时发生错误: %v", err)
			}
		}
	}
}

// OrchestratorAgentRegistryTest 测试编排器智能体注册管理
type OrchestratorAgentRegistryTest struct {
	model        model.Model
	orchestrator *core.Orchestrator
	agents       map[string]core.Agent
}

// NewOrchestratorAgentRegistryTest 创建编排器智能体注册测试用例
func NewOrchestratorAgentRegistryTest() *OrchestratorAgentRegistryTest {
	return &OrchestratorAgentRegistryTest{
		agents: make(map[string]core.Agent),
	}
}

// Name 返回测试名称
func (t *OrchestratorAgentRegistryTest) Name() string {
	return "编排器智能体注册测试"
}

// Description 返回测试描述
func (t *OrchestratorAgentRegistryTest) Description() string {
	return "测试编排器的智能体注册、注销和查询功能"
}

// Setup 设置测试环境
func (t *OrchestratorAgentRegistryTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model

	// 创建编排器
	config := core.DefaultOrchestratorConfig()
	t.orchestrator = core.NewOrchestrator(config)

	// 启动编排器
	err := t.orchestrator.Start()
	if err != nil {
		return fmt.Errorf("启动编排器失败: %v", err)
	}

	// 创建测试智能体
	agentTypes := []struct {
		id    string
		type_ core.AgentType
		desc  string
	}{
		{"worldview_agent", core.AgentTypeWorldview, "世界观智能体"},
		{"character_agent", core.AgentTypeCharacter, "角色智能体"},
		{"plot_agent", core.AgentTypePlot, "剧情智能体"},
		{"dialogue_agent", core.AgentTypeDialogue, "对话智能体"},
	}

	for _, at := range agentTypes {
		agent := core.NewGenericAdvancedAgent(at.id, at.type_, "你是"+at.desc)
		agent.SetModel(model)
		t.agents[at.id] = agent
	}

	return nil
}

// Execute 执行测试
func (t *OrchestratorAgentRegistryTest) Execute() error {
	hlog.Infof("开始执行编排器智能体注册测试...")

	// 1. 测试注册智能体
	for id, agent := range t.agents {
		hlog.Infof("注册智能体: %s", id)
		err := t.orchestrator.RegisterAgent(agent)
		if err != nil {
			return fmt.Errorf("注册智能体 %s 失败: %v", id, err)
		}
	}

	// 2. 测试获取智能体
	for id := range t.agents {
		hlog.Infof("通过ID获取智能体: %s", id)
		agent, exists := t.orchestrator.GetAgent(id)
		if !exists {
			return fmt.Errorf("无法找到已注册的智能体: %s", id)
		}
		if agent.GetID() != id {
			return fmt.Errorf("获取到的智能体ID不匹配: 预期 %s, 实际 %s", id, agent.GetID())
		}
	}

	// 3. 测试按类型获取智能体
	agentTypeTests := []struct {
		type_         core.AgentType
		expectedCount int
	}{
		{core.AgentTypeWorldview, 1},
		{core.AgentTypeCharacter, 1},
		{core.AgentTypePlot, 1},
		{core.AgentTypeDialogue, 1},
		{core.AgentTypeBackground, 0},
	}

	for _, att := range agentTypeTests {
		hlog.Infof("按类型获取智能体: %s", att.type_)
		agents := t.orchestrator.GetAgentsByType(att.type_)
		if len(agents) != att.expectedCount {
			return fmt.Errorf("类型 %s 的智能体数量不匹配: 预期 %d, 实际 %d",
				att.type_, att.expectedCount, len(agents))
		}
	}

	// 4. 测试注销智能体
	agentToUnregister := "dialogue_agent"
	hlog.Infof("注销智能体: %s", agentToUnregister)
	t.orchestrator.UnregisterAgent(agentToUnregister)

	// 5. 验证注销结果
	_, exists := t.orchestrator.GetAgent(agentToUnregister)
	if exists {
		return fmt.Errorf("智能体 %s 注销后仍能被找到", agentToUnregister)
	}

	return nil
}

// Verify 验证测试结果
func (t *OrchestratorAgentRegistryTest) Verify() (bool, string) {
	// 验证智能体注册状态
	registeredCount := 0
	for id := range t.agents {
		if id == "dialogue_agent" {
			// 这个应该已经被注销了
			if _, exists := t.orchestrator.GetAgent(id); exists {
				return false, fmt.Sprintf("智能体 %s 应该已被注销但仍存在", id)
			}
		} else {
			// 其他应该仍然存在
			if _, exists := t.orchestrator.GetAgent(id); exists {
				registeredCount++
			}
		}
	}

	expectedCount := len(t.agents) - 1 // 减去被注销的智能体
	if registeredCount != expectedCount {
		return false, fmt.Sprintf("注册的智能体数量不匹配: 预期 %d, 实际 %d",
			expectedCount, registeredCount)
	}

	return true, ""
}

// Cleanup 清理测试资源
func (t *OrchestratorAgentRegistryTest) Cleanup() {
	if t.orchestrator != nil {
		// 注销所有智能体
		for id := range t.agents {
			t.orchestrator.UnregisterAgent(id)
		}

		// 停止编排器
		status := t.orchestrator.GetStatus()
		if running, ok := status["running"].(bool); ok && running {
			t.orchestrator.Stop()
		}
	}
}

// OrchestratorMessageRoutingTest 测试编排器消息路由
type OrchestratorMessageRoutingTest struct {
	model        model.Model
	orchestrator *core.Orchestrator
	agents       map[string]core.Agent
	messages     []*core.Message
	responses    []*core.Message
}

// NewOrchestratorMessageRoutingTest 创建编排器消息路由测试用例
func NewOrchestratorMessageRoutingTest() *OrchestratorMessageRoutingTest {
	return &OrchestratorMessageRoutingTest{
		agents:    make(map[string]core.Agent),
		messages:  make([]*core.Message, 0),
		responses: make([]*core.Message, 0),
	}
}

// Name 返回测试名称
func (t *OrchestratorMessageRoutingTest) Name() string {
	return "编排器消息路由测试"
}

// Description 返回测试描述
func (t *OrchestratorMessageRoutingTest) Description() string {
	return "测试编排器的消息发送、广播和路由功能"
}

// Setup 设置测试环境
func (t *OrchestratorMessageRoutingTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model
	t.agents = make(map[string]core.Agent)
	t.messages = make([]*core.Message, 0)
	t.responses = make([]*core.Message, 0)

	// 创建编排器
	config := core.DefaultOrchestratorConfig()
	t.orchestrator = core.NewOrchestrator(config)

	// 启动编排器
	err := t.orchestrator.Start()
	if err != nil {
		return fmt.Errorf("启动编排器失败: %v", err)
	}

	// 创建测试智能体
	worldviewAgent := core.NewGenericAdvancedAgent(
		"worldview_agent",
		core.AgentTypeWorldview,
		"你是世界观智能体，负责生成世界设定。",
	)
	worldviewAgent.SetModel(model)

	characterAgent := core.NewGenericAdvancedAgent(
		"character_agent",
		core.AgentTypeCharacter,
		"你是角色智能体，负责生成角色信息。",
	)
	characterAgent.SetModel(model)

	t.agents["worldview_agent"] = worldviewAgent
	t.agents["character_agent"] = characterAgent

	// 注册智能体
	for _, agent := range t.agents {
		err := t.orchestrator.RegisterAgent(agent)
		if err != nil {
			return fmt.Errorf("注册智能体 %s 失败: %v", agent.GetID(), err)
		}

		// 初始化智能体
		err = agent.Initialize(ctx)
		if err != nil {
			return fmt.Errorf("初始化智能体 %s 失败: %v", agent.GetID(), err)
		}
	}

	// 准备测试消息
	directMsg := core.NewMessage(core.MessageTypeRequest, "user", "worldview_agent")
	directMsg.Subject = "直接消息测试"
	directMsg.Content = "请描述一个赛博朋克世界的基本设定。"

	broadcastMsg := core.NewMessage(core.MessageTypeNotification, "system", "")
	broadcastMsg.Subject = "广播消息测试"
	broadcastMsg.Content = "这是一条广播消息，所有智能体都应该收到。"

	t.messages = append(t.messages, directMsg, broadcastMsg)

	return nil
}

// Execute 执行测试
func (t *OrchestratorMessageRoutingTest) Execute() error {
	hlog.Infof("开始执行编排器消息路由测试...")

	// 1. 测试直接消息发送
	directMsg := t.messages[0]
	hlog.Infof("发送直接消息: [%s] %s -> %s", directMsg.Type, directMsg.From, directMsg.To)

	response, err := t.orchestrator.SendMessage(context.Background(), directMsg)
	if err != nil {
		return fmt.Errorf("发送直接消息失败: %v", err)
	}

	if response == nil {
		return fmt.Errorf("直接消息没有返回响应")
	}

	t.responses = append(t.responses, response)

	// 2. 测试广播消息
	broadcastMsg := t.messages[1]
	hlog.Infof("发送广播消息: [%s] %s", broadcastMsg.Type, broadcastMsg.Subject)

	// 广播到所有世界观智能体
	responses, err := t.orchestrator.BroadcastMessage(context.Background(), core.AgentTypeWorldview, broadcastMsg)
	if err != nil {
		return fmt.Errorf("发送广播消息失败: %v", err)
	}

	if responses == nil || len(responses) == 0 {
		return fmt.Errorf("广播消息没有返回响应")
	}

	// 检查是否收到了广播类型的智能体的响应
	// 计算指定类型的智能体数量
	worldviewAgentCount := 0
	for _, agent := range t.agents {
		if agent.GetType() == core.AgentTypeWorldview {
			worldviewAgentCount++
		}
	}

	if len(responses) != worldviewAgentCount {
		return fmt.Errorf("广播消息响应数量不匹配: 预期 %d, 实际 %d",
			worldviewAgentCount, len(responses))
	}

	t.responses = append(t.responses, responses...)

	// 3. 测试特定类型广播
	typeMsg := core.NewMessage(core.MessageTypeCommand, "system", "")
	typeMsg.Subject = "类型广播测试"
	typeMsg.Content = "这是一条发送给所有角色智能体的命令。"

	// 广播到角色智能体
	typeResponses, err := t.orchestrator.BroadcastMessage(
		context.Background(),
		core.AgentTypeCharacter,
		typeMsg,
	)
	if err != nil {
		return fmt.Errorf("发送类型广播消息失败: %v", err)
	}

	// 应该只有一个角色智能体收到
	if len(typeResponses) != 1 {
		return fmt.Errorf("类型广播消息响应数量不匹配: 预期 1, 实际 %d",
			len(typeResponses))
	}

	t.responses = append(t.responses, typeResponses...)

	return nil
}

// Verify 验证测试结果
func (t *OrchestratorMessageRoutingTest) Verify() (bool, string) {
	// 检查收到的响应
	if len(t.responses) < 3 { // 至少应有: 直接消息响应 + 两个广播响应 + 一个类型广播响应
		return false, fmt.Sprintf("响应数量不足: 预期至少3个, 实际 %d", len(t.responses))
	}

	// 验证直接消息响应
	directResponse := t.responses[0]
	if directResponse.Type != core.MessageTypeResponse {
		return false, fmt.Sprintf("直接消息响应类型错误: 预期 %s, 实际 %s",
			core.MessageTypeResponse, directResponse.Type)
	}

	if directResponse.From != "worldview_agent" {
		return false, fmt.Sprintf("直接消息响应发送者错误: 预期 worldview_agent, 实际 %s",
			directResponse.From)
	}

	return true, ""
}

// Cleanup 清理测试资源
func (t *OrchestratorMessageRoutingTest) Cleanup() {
	if t.orchestrator != nil {
		// 注销所有智能体
		for id := range t.agents {
			t.orchestrator.UnregisterAgent(id)
		}

		// 停止编排器
		status := t.orchestrator.GetStatus()
		if running, ok := status["running"].(bool); ok && running {
			t.orchestrator.Stop()
		}
	}
}
