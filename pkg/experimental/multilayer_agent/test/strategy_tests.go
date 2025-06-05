// package test 实现策略智能体相关的测试用例
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/tmc/langchaingo/llms"
	"novelai/pkg/experimental/multilayer_agent/core"
	"novelai/pkg/experimental/multilayer_agent/decision"
	"novelai/pkg/experimental/multilayer_agent/shared/memory"
	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"
)

// StrategyAgentInitTest 测试策略智能体初始化功能
type StrategyAgentInitTest struct {
	model          model.Model
	agent          *decision.StrategyAgent
	memoryManager  *TestMemoryManager
	initialStrategy string
}

// NewStrategyAgentInitTest 创建策略智能体初始化测试用例
func NewStrategyAgentInitTest() *StrategyAgentInitTest {
	return &StrategyAgentInitTest{
		initialStrategy: "测试初始策略",
	}
}

// Name 返回测试名称
func (t *StrategyAgentInitTest) Name() string {
	return "策略智能体初始化测试"
}

// Description 返回测试描述
func (t *StrategyAgentInitTest) Description() string {
	return "测试策略智能体的初始化功能，包括状态和历史恢复"
}

// Setup 设置测试环境
func (t *StrategyAgentInitTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model
	
	// 创建内存管理器
	t.memoryManager = NewTestMemoryManager()
	
	// 创建策略智能体
	t.agent = decision.NewStrategyAgent("strategy_test_agent", "")
	t.agent.SetModel(model)
	t.agent.SetMemoryManager(t.memoryManager)
	
	// 预先存储一些状态到内存
	lastProcessTime := time.Now().Add(-24 * time.Hour)
	stateKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyLastProcessTime)
	t.memoryManager.Save(ctx, stateKey, lastProcessTime.Format(time.RFC3339))
	
	strategyKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy)
	t.memoryManager.Save(ctx, strategyKey, t.initialStrategy)
	
	historyKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory)
	t.memoryManager.Save(ctx, historyKey, []string{"历史策略1", "历史策略2", t.initialStrategy})
	
	return nil
}

// Execute 执行测试
func (t *StrategyAgentInitTest) Execute() error {
	ctx := context.Background()
	
	// 初始化智能体
	err := t.agent.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("策略智能体初始化失败: %w", err)
	}
	
	hlog.Infof("策略智能体初始化完成")
	return nil
}

// Verify 验证测试结果
func (t *StrategyAgentInitTest) Verify() (bool, string) {
	// 验证当前策略是否被正确加载
	currentStrategy := t.agent.GetCurrentStrategy()
	if currentStrategy != t.initialStrategy {
		return false, fmt.Sprintf("当前策略加载错误: 期望 '%s', 实际 '%s'", 
			t.initialStrategy, currentStrategy)
	}
	
	// 验证策略历史是否被正确加载
	ctx := context.Background()
	strategyHistory := t.agent.UseStrategyHistory(ctx)
	if len(strategyHistory) != 3 {
		return false, fmt.Sprintf("策略历史记录数量错误: 期望 3, 实际 %d", 
			len(strategyHistory))
	}
	
	return true, "策略智能体初始化测试通过"
}

// Cleanup 清理测试资源
func (t *StrategyAgentInitTest) Cleanup() {
	t.agent = nil
	t.memoryManager = nil
}

// StrategyAgentProcessTest 测试策略智能体消息处理功能
type StrategyAgentProcessTest struct {
	model    model.Model
	agent    *decision.StrategyAgent
	testModel *TestModel
	messages []*core.Message
	responses []*core.Message
}

// NewStrategyAgentProcessTest 创建策略智能体消息处理测试用例
func NewStrategyAgentProcessTest() *StrategyAgentProcessTest {
	return &StrategyAgentProcessTest{
		messages: make([]*core.Message, 0),
		responses: make([]*core.Message, 0),
	}
}

// Name 返回测试名称
func (t *StrategyAgentProcessTest) Name() string {
	return "策略智能体消息处理测试"
}

// Description 返回测试描述
func (t *StrategyAgentProcessTest) Description() string {
	return "测试策略智能体对不同类型消息的处理能力，包括策略制定、更新和委派"
}

// Setup 设置测试环境
func (t *StrategyAgentProcessTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	// 创建测试模型
	t.testModel = NewTestModel()
	t.model = t.testModel
	
	// 创建策略智能体
	t.agent = decision.NewStrategyAgent("strategy_test_agent", "")
	t.agent.SetModel(t.testModel)
	t.agent.SetMemoryManager(NewTestMemoryManager())
	
	// 初始化智能体
	if err := t.agent.Initialize(ctx); err != nil {
		return fmt.Errorf("策略智能体初始化失败: %w", err)
	}
	
	// 准备测试消息
	// 1. 新策略请求
	newStrategyResponse := map[string]string{
		"action": "new_strategy",
		"strategy": "专注于生成赛博朋克世界观的小说",
	}
	t.testModel.SetNextResponse(mustMarshalJSON(newStrategyResponse))
	
	msg1 := core.NewMessage(core.MessageTypeRequest, "user", t.agent.GetID())
	msg1.Subject = "创建新策略"
	msg1.Content = "我需要一个赛博朋克风格的小说世界观"
	t.messages = append(t.messages, msg1)
	
	// 2. 更新策略请求
	updateStrategyResponse := map[string]string{
		"action": "update_strategy",
		"strategy": "专注于生成赛博朋克世界观并加入反乌托邦元素的小说",
	}
	t.testModel.SetNextResponse(mustMarshalJSON(updateStrategyResponse))
	
	msg2 := core.NewMessage(core.MessageTypeRequest, "user", t.agent.GetID())
	msg2.Subject = "更新当前策略"
	msg2.Content = "请在赛博朋克世界观中加入更多的反乌托邦元素"
	t.messages = append(t.messages, msg2)
	
	// 3. 委派任务请求
	delegateResponse := map[string]string{
		"action": "delegate",
		"agent_id": "worldview_agent",
		"message": "请生成一个赛博朋克反乌托邦风格的世界观",
	}
	t.testModel.SetNextResponse(mustMarshalJSON(delegateResponse))
	
	msg3 := core.NewMessage(core.MessageTypeRequest, "user", t.agent.GetID())
	msg3.Subject = "分配世界观任务"
	msg3.Content = "需要详细的世界观描述"
	t.messages = append(t.messages, msg3)
	
	// 4. 直接回复请求
	replyResponse := map[string]string{
		"action": "reply",
		"message": "我已经理解了您的需求，将立即开始工作。",
	}
	t.testModel.SetNextResponse(mustMarshalJSON(replyResponse))
	
	msg4 := core.NewMessage(core.MessageTypeRequest, "user", t.agent.GetID())
	msg4.Subject = "简单询问"
	msg4.Content = "你能理解我的需求吗？"
	t.messages = append(t.messages, msg4)
	
	return nil
}

// Execute 执行测试
func (t *StrategyAgentProcessTest) Execute() error {
	ctx := context.Background()
	
	// 处理所有测试消息
	for i, msg := range t.messages {
		hlog.Infof("处理测试消息 %d: %s", i+1, msg.Subject)
		
		response, err := t.agent.Process(ctx, msg)
		if err != nil {
			return fmt.Errorf("处理消息 %d 失败: %w", i+1, err)
		}
		
		t.responses = append(t.responses, response)
		hlog.Infof("消息 %d 处理完成，响应类型: %s", i+1, response.Type)
	}
	
	return nil
}

// Verify 验证测试结果
func (t *StrategyAgentProcessTest) Verify() (bool, string) {
	if len(t.responses) != len(t.messages) {
		return false, fmt.Sprintf("响应数量不匹配: 期望 %d, 实际 %d", 
			len(t.messages), len(t.responses))
	}
	
	// 验证新策略响应
	if t.responses[0].Subject != "策略更新" {
		return false, fmt.Sprintf("消息1响应主题错误: %s", t.responses[0].Subject)
	}
	
	// 验证更新策略响应
	if t.responses[1].Subject != "策略更新" {
		return false, fmt.Sprintf("消息2响应主题错误: %s", t.responses[1].Subject)
	}
	
	// 验证策略历史记录
	ctx := context.Background()
	history := t.agent.UseStrategyHistory(ctx)
	if len(history) != 2 {
		return false, fmt.Sprintf("策略历史记录数量错误: 期望 2, 实际 %d", len(history))
	}
	
	// 验证委派任务响应
	if t.responses[2].Type != core.MessageTypeRequest || t.responses[2].To != "worldview_agent" {
		return false, "委派任务响应错误"
	}
	
	// 验证直接回复响应
	if t.responses[3].Type != core.MessageTypeResponse || t.responses[3].To != "user" {
		return false, "直接回复响应错误"
	}
	
	return true, "策略智能体消息处理测试通过"
}

// Cleanup 清理测试资源
func (t *StrategyAgentProcessTest) Cleanup() {
	t.agent = nil
	t.testModel = nil
	t.messages = nil
	t.responses = nil
}

// StrategyManagementTest 测试策略管理功能
type StrategyManagementTest struct {
	model          model.Model
	agent          *decision.StrategyAgent
	memoryManager  *TestMemoryManager
	initialStrategy string
	updatedStrategy string
}

// NewStrategyManagementTest 创建策略管理测试用例
func NewStrategyManagementTest() *StrategyManagementTest {
	return &StrategyManagementTest{
		initialStrategy: "初始测试策略",
		updatedStrategy: "更新后的测试策略",
	}
}

// Name 返回测试名称
func (t *StrategyManagementTest) Name() string {
	return "策略管理测试"
}

// Description 返回测试描述
func (t *StrategyManagementTest) Description() string {
	return "测试策略智能体的策略管理功能，包括设置、获取和历史记录"
}

// Setup 设置测试环境
func (t *StrategyManagementTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model
	
	// 创建内存管理器
	t.memoryManager = NewTestMemoryManager()
	
	// 创建策略智能体
	t.agent = decision.NewStrategyAgent("strategy_test_agent", "")
	t.agent.SetModel(model)
	t.agent.SetMemoryManager(t.memoryManager)
	
	// 初始化智能体
	if err := t.agent.Initialize(ctx); err != nil {
		return fmt.Errorf("策略智能体初始化失败: %w", err)
	}
	
	return nil
}

// Execute 执行测试
func (t *StrategyManagementTest) Execute() error {
	ctx := context.Background()
	
	// 设置初始策略
	err := t.agent.SetCurrentStrategy(ctx, t.initialStrategy)
	if err != nil {
		return fmt.Errorf("设置初始策略失败: %w", err)
	}
	
	// 获取并验证初始策略
	currentStrategy := t.agent.GetCurrentStrategy()
	if currentStrategy != t.initialStrategy {
		return fmt.Errorf("初始策略不匹配: 期望 '%s', 实际 '%s'", 
			t.initialStrategy, currentStrategy)
	}
	
	// 设置更新策略
	err = t.agent.SetCurrentStrategy(ctx, t.updatedStrategy)
	if err != nil {
		return fmt.Errorf("设置更新策略失败: %w", err)
	}
	
	return nil
}

// Verify 验证测试结果
func (t *StrategyManagementTest) Verify() (bool, string) {
	// 验证当前策略是否被正确更新
	currentStrategy := t.agent.GetCurrentStrategy()
	if currentStrategy != t.updatedStrategy {
		return false, fmt.Sprintf("当前策略更新错误: 期望 '%s', 实际 '%s'", 
			t.updatedStrategy, currentStrategy)
	}
	
	// 验证策略历史是否被正确记录
	strategyHistory := t.agent.UseStrategyHistory(context.Background())
	if len(strategyHistory) != 2 {
		return false, fmt.Sprintf("策略历史记录数量错误: 期望 2, 实际 %d", 
			len(strategyHistory))
	}
	
	if strategyHistory[1] != t.updatedStrategy {
		return false, fmt.Sprintf("策略历史记录错误: 期望最新策略为 '%s', 实际为 '%s'", 
			t.updatedStrategy, strategyHistory[1])
	}
	
	// 验证是否正确存储到内存
	strategyKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy)
	value, err := t.memoryManager.Load(context.Background(), strategyKey)
	if err != nil {
		return false, fmt.Sprintf("无法从内存加载当前策略: %v", err)
	}
	
	if value != t.updatedStrategy {
		return false, fmt.Sprintf("内存中的策略错误: 期望 '%s', 实际 '%s'", 
			t.updatedStrategy, value)
	}
	
	return true, "策略管理测试通过"
}

// Cleanup 清理测试资源
func (t *StrategyManagementTest) Cleanup() {
	t.agent = nil
	t.memoryManager = nil
}

// StrategyShutdownTest 测试策略智能体关闭功能
type StrategyShutdownTest struct {
	model          model.Model
	agent          *decision.StrategyAgent
	memoryManager  *TestMemoryManager
	testStrategy   string
}

// NewStrategyShutdownTest 创建策略智能体关闭测试用例
func NewStrategyShutdownTest() *StrategyShutdownTest {
	return &StrategyShutdownTest{
		testStrategy: "关闭测试策略",
	}
}

// Name 返回测试名称
func (t *StrategyShutdownTest) Name() string {
	return "策略智能体关闭测试"
}

// Description 返回测试描述
func (t *StrategyShutdownTest) Description() string {
	return "测试策略智能体关闭时状态保存功能"
}

// Setup 设置测试环境
func (t *StrategyShutdownTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model
	
	// 创建内存管理器
	t.memoryManager = NewTestMemoryManager()
	
	// 创建策略智能体
	t.agent = decision.NewStrategyAgent("strategy_test_agent", "")
	t.agent.SetModel(model)
	t.agent.SetMemoryManager(t.memoryManager)
	
	// 初始化智能体
	if err := t.agent.Initialize(ctx); err != nil {
		return fmt.Errorf("策略智能体初始化失败: %w", err)
	}
	
	// 设置测试策略
	if err := t.agent.SetCurrentStrategy(ctx, t.testStrategy); err != nil {
		return fmt.Errorf("设置测试策略失败: %w", err)
	}
	
	return nil
}

// Execute 执行测试
func (t *StrategyShutdownTest) Execute() error {
	ctx := context.Background()
	
	// 关闭智能体
	err := t.agent.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("关闭策略智能体失败: %w", err)
	}
	
	hlog.Infof("策略智能体关闭完成")
	return nil
}

// Verify 验证测试结果
func (t *StrategyShutdownTest) Verify() (bool, string) {
	ctx := context.Background()
	
	// 验证状态是否正确保存到内存
	strategyKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyCurrentStrategy)
	value, err := t.memoryManager.Load(ctx, strategyKey)
	if err != nil {
		return false, fmt.Sprintf("无法从内存加载当前策略: %v", err)
	}
	
	if value != t.testStrategy {
		return false, fmt.Sprintf("内存中的策略错误: 期望 '%s', 实际 '%s'", 
			t.testStrategy, value)
	}
	
	// 验证历史是否正确保存
	historyKey := memory.CreateTaggedKey(t.agent.GetID(), memory.MemoryCategoryState, memory.StateKeyStrategyHistory)
	historyValue, err := t.memoryManager.Load(ctx, historyKey)
	if err != nil {
		return false, fmt.Sprintf("无法从内存加载策略历史: %v", err)
	}
	
	history, ok := historyValue.([]string)
	if !ok {
		return false, "内存中的策略历史类型错误"
	}
	
	if len(history) == 0 || history[len(history)-1] != t.testStrategy {
		return false, "内存中的策略历史错误"
	}
	
	return true, "策略智能体关闭测试通过"
}

// Cleanup 清理测试资源
func (t *StrategyShutdownTest) Cleanup() {
	t.agent = nil
	t.memoryManager = nil
}

// 辅助函数和辅助测试类型

// TestMemoryManager 测试用内存管理器
type TestMemoryManager struct {
	data map[string]interface{}
}

// NewTestMemoryManager 创建测试用内存管理器
func NewTestMemoryManager() *TestMemoryManager {
	return &TestMemoryManager{
		data: make(map[string]interface{}),
	}
}

// Save 实现memory.Manager接口的Save方法
func (m *TestMemoryManager) Save(ctx context.Context, key string, value interface{}) error {
	m.data[key] = value
	return nil
}

// Load 实现memory.Manager接口的Load方法
func (m *TestMemoryManager) Load(ctx context.Context, key string) (interface{}, error) {
	value, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("记忆不存在: %s", key)
	}
	return value, nil
}

// Delete 实现memory.Manager接口的Delete方法
func (m *TestMemoryManager) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// List 实现memory.Manager接口的List方法
func (m *TestMemoryManager) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for key := range m.data {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// Clear 实现memory.Manager接口的Clear方法
func (m *TestMemoryManager) Clear(ctx context.Context) error {
	m.data = make(map[string]interface{})
	return nil
}

// TestModel 测试用模型
type TestModel struct {
	responses []string
	index     int
}

// NewTestModel 创建测试用模型
func NewTestModel() *TestModel {
	return &TestModel{
		responses: make([]string, 0),
		index:     0,
	}
}

// SetNextResponse 设置下一个响应
func (m *TestModel) SetNextResponse(response string) {
	m.responses = append(m.responses, response)
}

// Call 实现Model接口的Call方法
func (m *TestModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	if m.index >= len(m.responses) {
		return "{\"action\":\"reply\",\"message\":\"默认测试响应\"}", nil
	}
	
	response := m.responses[m.index]
	m.index++
	return response, nil
}

// GenerateContent 实现Model接口的GenerateContent方法
func (m *TestModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	var response string
	if m.index >= len(m.responses) {
		response = "{\"action\":\"reply\",\"message\":\"默认测试响应\"}"
	} else {
		response = m.responses[m.index]
		m.index++
	}
	
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: response,
			},
		},
	}, nil
}

// ModelName 实现Model接口的ModelName方法
func (m *TestModel) ModelName() string {
	return "TestModel"
}

// SupportsJSON 实现Model接口的SupportsJSON方法
func (m *TestModel) SupportsJSON() bool {
	return true
}

// ModelType 实现Model接口的ModelType方法
func (m *TestModel) ModelType() model.ModelType {
	return "test"
}

// GetTokenLimit 实现Model接口的GetTokenLimit方法
func (m *TestModel) GetTokenLimit() int {
	return 4096
}

// EstimateTokens 实现Model接口的EstimateTokens方法
func (m *TestModel) EstimateTokens(text string) (int, error) {
	return len(text) / 4, nil
}

// SupportsStreaming 实现Model接口的SupportsStreaming方法
func (m *TestModel) SupportsStreaming() bool {
	return false
}

// SupportsVision 实现Model接口的SupportsVision方法
func (m *TestModel) SupportsVision() bool {
	return false
}

// 辅助函数
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("JSON编码失败: %v", err))
	}
	return string(data)
}

// RegisterStrategyTests 注册策略智能体相关测试
func RegisterStrategyTests(suite *TestSuite) {
	suite.AddTestCase(NewStrategyAgentInitTest())
	suite.AddTestCase(NewStrategyAgentProcessTest())
	suite.AddTestCase(NewStrategyManagementTest())
	suite.AddTestCase(NewStrategyShutdownTest())
}
