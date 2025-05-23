// package main 提供多层代理系统的临时测试入口点
// 本文件用于全面测试多层代理系统的各个组件功能
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"novelai/pkg/experimental/multilayer_agent/core"
	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"
	"novelai/pkg/experimental/multilayer_agent/shared/tools/example_tool"
)

// 测试配置结构
type TestConfig struct {
	// 模型配置
	ModelType model.ModelType `json:"model_type"`
	ModelName string          `json:"model_name"`
	ModelURL  string          `json:"model_url,omitempty"`
	APIToken  string          `json:"api_token,omitempty"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Debug     bool            `json:"debug"`

	// 测试参数
	RunModelTest       bool `json:"run_model_test"`
	RunToolTest        bool `json:"run_tool_test"`
	RunIntegrationTest bool `json:"run_integration_test"`
}

// 默认测试配置
var defaultConfig = TestConfig{
	ModelType:          model.ModelTypeOllama,
	ModelName:          "llama2",
	Debug:              true,
	RunModelTest:       true,
	RunToolTest:        true,
	RunIntegrationTest: true,
}

func main() {
	// 创建上下文
	ctx := context.Background()

	// 设置日志
	setupLogging()

	// 加载配置
	config := loadConfig()

	// 打印测试信息
	hlog.Infof("开始多层代理系统集成测试")
	hlog.Infof("使用模型: %s (%s)", config.ModelName, config.ModelType)

	// 创建模型
	testModel, err := createModel(config)
	if err != nil {
		hlog.Errorf("创建模型失败: %v", err)
		os.Exit(1)
	}

	// 创建工具注册表
	registry := agenttools.NewToolRegistry()

	// 注册示例工具
	if config.RunModelTest {
		hlog.Infof("\n===== 开始模型测试 =====")
		runModelTest(ctx, testModel)
	}

	if config.RunToolTest {
		hlog.Infof("\n===== 开始工具测试 =====")
		runToolTest(ctx, registry)
	}

	if config.RunIntegrationTest {
		hlog.Infof("\n===== 开始集成测试 =====")
		runIntegrationTest(ctx, testModel, registry)
	}

	// 新增：运行多层智能体系统测试
	hlog.Infof("\n===== 开始多层智能体系统测试 =====")
	runMultiAgentSystemTest(ctx, testModel, registry)

	hlog.Infof("多层代理系统测试完成")
}

// setupLogging 配置日志系统
func setupLogging() {
	// 这里可以添加更详细的日志配置
	hlog.SetLevel(hlog.LevelDebug)
}

// loadConfig 加载测试配置
func loadConfig() TestConfig {
	// 从环境变量或命令行参数加载配置
	// 本例中简单返回默认配置
	config := defaultConfig

	// 从环境变量覆盖模型配置
	if modelName := os.Getenv("AGENT_MODEL_NAME"); modelName != "" {
		config.ModelName = modelName
	}

	if modelURL := os.Getenv("AGENT_MODEL_URL"); modelURL != "" {
		config.ModelURL = modelURL
	}

	if apiToken := os.Getenv("AGENT_API_TOKEN"); apiToken != "" {
		config.APIToken = apiToken
	}

	// 打印当前配置
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	hlog.Infof("测试配置: %s", string(configJSON))

	return config
}

// createModel 创建并初始化模型实例
func createModel(config TestConfig) (model.Model, error) {
	// 创建模型工厂
	factory := model.NewModelFactory()

	// 准备模型选项
	options := model.ModelOptions{
		ModelName:        config.ModelName,
		BaseURL:          config.ModelURL,
		APIToken:         config.APIToken,
		Debug:            config.Debug,
		DefaultMaxTokens: config.MaxTokens,
	}

	// 创建模型实例
	return factory.CreateModel(config.ModelType, options)
}

// createExampleTool 创建示例工具
func createExampleTool() tools.Tool {
	// 创建工具配置
	config := &example_tool.ExampleToolConfig{
		Verbose:    true,
		MaxRetries: 3,
		DefaultParams: example_tool.ExampleToolParams{
			Text:    "示例文本",
			Number:  42,
			Flag:    true,
			Options: []string{"选项A", "选项B"},
		},
	}

	// 创建并返回工具
	return example_tool.NewExampleTool(config)
}

// runModelTest 测试模型功能
func runModelTest(ctx context.Context, testModel model.Model) {
	hlog.Infof("模型类型: %s, 名称: %s", testModel.ModelType(), testModel.ModelName())
	hlog.Infof("模型最大Token限制: %d", testModel.GetTokenLimit())
	hlog.Infof("支持JSON输出: %v", testModel.SupportsJSON())
	hlog.Infof("支持流式输出: %v", testModel.SupportsStreaming())
	hlog.Infof("支持视觉输入: %v", testModel.SupportsVision())

	// 测试简单文本生成
	prompt := "用一句话解释什么是人工智能"
	hlog.Infof("测试简单文本生成，提示词: %s", prompt)

	start := time.Now()
	response, err := testModel.Call(ctx, prompt)
	if err != nil {
		hlog.Errorf("模型调用失败: %v", err)
		return
	}

	elapsed := time.Since(start)
	hlog.Infof("生成耗时: %v", elapsed)
	hlog.Infof("生成结果: %s", response)

	// 测试结构化输出
	if testModel.SupportsJSON() {
		jsonPrompt := "生成JSON格式的小说角色描述，包含名称、年龄、背景故事和技能"
		hlog.Infof("测试JSON输出，提示词: %s", jsonPrompt)

		start = time.Now()
		response, err = testModel.Call(ctx, jsonPrompt, llms.WithJSONMode())
		if err != nil {
			hlog.Errorf("JSON模式调用失败: %v", err)
			return
		}

		elapsed = time.Since(start)
		hlog.Infof("JSON生成耗时: %v", elapsed)
		hlog.Infof("JSON生成结果: %s", response)
	}
}

// runToolTest 测试工具调用
func runToolTest(ctx context.Context, registry *agenttools.ToolRegistry) {
	// 获取示例工具
	exampleTool, err := registry.GetTool("example_tool")
	if err != nil {
		hlog.Errorf("获取示例工具失败: %v", err)
		return
	}

	// 准备测试输入
	testInput := `{
		"text": "测试输入",
		"number": 100,
		"flag": true,
		"options": ["测试选项1", "测试选项2", "测试选项3"],
		"nested_data": {
			"key": "测试键",
			"value": "测试值"
		}
	}`

	// 调用工具
	hlog.Infof("调用工具: %s", exampleTool.Name())
	hlog.Infof("工具输入: %s", testInput)

	start := time.Now()
	result, err := exampleTool.Call(ctx, testInput)
	if err != nil {
		hlog.Errorf("工具调用失败: %v", err)
		return
	}

	elapsed := time.Since(start)
	hlog.Infof("工具调用耗时: %v", elapsed)
	hlog.Infof("工具调用结果: %s", result)

	// 测试使用工具调用适配器
	hlog.Infof("测试工具调用适配器")
	caller := agenttools.NewToolCaller(registry)

	// 使用工具名称和参数调用
	req := agenttools.ToolRequest{
		ToolName: "example_tool",
		Input:    json.RawMessage(testInput),
	}
	resp, err := caller.CallTool(ctx, req)
	if err != nil {
		hlog.Errorf("通过适配器调用工具失败: %v", err)
		return
	}

	// 检查调用是否成功
	if !resp.Success {
		hlog.Errorf("工具调用失败: %s", resp.Error)
		return
	}

	result = resp.Result

	hlog.Infof("适配器调用结果: %s", result)
}

// runIntegrationTest 运行集成测试
func runIntegrationTest(ctx context.Context, testModel model.Model, registry *agenttools.ToolRegistry) {
	// 测试模型与工具集成
	// 这里模拟代理决策过程：模型决定使用哪个工具并提供参数

	// 模拟代理提示词
	agentPrompt := fmt.Sprintf(`你是一个能够使用工具的AI助手。可用的工具有:

工具名称: %s
工具描述: %s

请使用这个工具完成以下任务:
创建一个包含文本"集成测试"和数字99的请求。

输出格式应为JSON，包含你要调用的工具名称和工具参数。
例如:
{
  "tool": "工具名称",
  "params": {参数JSON对象}
}`, "example_tool", "这是一个示例工具")

	// 调用模型获取工具调用决策
	hlog.Infof("模拟代理决策过程")
	hlog.Infof("提示词: %s", agentPrompt)

	response, err := testModel.Call(ctx, agentPrompt, llms.WithJSONMode())
	if err != nil {
		hlog.Errorf("模型调用失败: %v", err)
		return
	}

	hlog.Infof("模型响应: %s", response)

	// 解析模型响应
	var toolCall struct {
		Tool   string          `json:"tool"`
		Params json.RawMessage `json:"params"`
	}

	err = json.Unmarshal([]byte(response), &toolCall)
	if err != nil {
		hlog.Errorf("解析模型响应失败: %v", err)
		return
	}

	// 检查工具名称
	if toolCall.Tool != "example_tool" {
		hlog.Errorf("模型选择了错误的工具: %s", toolCall.Tool)
		return
	}

	// 调用工具
	hlog.Infof("调用工具: %s", toolCall.Tool)
	hlog.Infof("工具参数: %s", string(toolCall.Params))

	// 获取工具
	tool, err := registry.GetTool(toolCall.Tool)
	if err != nil {
		hlog.Errorf("获取工具失败: %v", err)
		return
	}

	// 调用工具
	result, err := tool.Call(ctx, string(toolCall.Params))
	if err != nil {
		hlog.Errorf("工具调用失败: %v", err)
		return
	}

	hlog.Infof("工具调用结果: %s", result)

	// 将工具结果发送回模型
	finalPrompt := fmt.Sprintf(`之前你决定使用工具 %s，参数是 %s。
工具执行结果是: %s

请基于这个结果给用户一个总结。`, toolCall.Tool, string(toolCall.Params), result)

	finalResponse, err := testModel.Call(ctx, finalPrompt)
	if err != nil {
		hlog.Errorf("最终模型调用失败: %v", err)
		return
	}

	hlog.Infof("集成测试最终响应: %s", finalResponse)
}

// runMultiAgentSystemTest 测试多层智能体系统功能
func runMultiAgentSystemTest(ctx context.Context, testModel model.Model, registry *agenttools.ToolRegistry) {
	hlog.Infof("多层智能体系统测试开始...")
	
	// 测试分为两部分：
	// 1. 测试单个智能体的能力
	// 2. 测试编排器及多智能体协作
	
	// 第一部分：测试单个智能体
	hlog.Infof("=== 测试单个智能体 ===")
	testSingleAgent(ctx, testModel)
	
	// 第二部分：测试编排器
	hlog.Infof("=== 测试编排器 ===")
	testOrchestrator(ctx, testModel, registry)
	
	hlog.Infof("多层智能体系统测试完成")
}

// testSingleAgent 测试单个智能体的能力
func testSingleAgent(ctx context.Context, testModel model.Model) {
	// 1. 创建测试智能体
	hlog.Infof("创建测试智能体")
	testAgent := core.NewGenericAdvancedAgent(
		"test_agent", 
		core.AgentTypeWorldview, 
		"你是一个测试智能体，可以调用工具并生成响应。",
	)
	
	// 2. 设置模型
	hlog.Infof("为智能体设置模型")
	testAgent.SetModel(testModel)
	
	// 3. 初始化智能体
	hlog.Infof("初始化智能体")
	err := testAgent.Initialize(ctx)
	if err != nil {
		hlog.Errorf("初始化智能体失败: %v", err)
		return
	}
	
	// 4. 测试智能体处理消息的能力
	hlog.Infof("测试智能体处理消息能力")
	testMsg := core.NewMessage(core.MessageTypeRequest, "user", "test_agent")
	testMsg.Subject = "测试智能体处理能力"
	testMsg.Content = "请生成一个奇幻世界的基本设定。"
	
	// 5. 直接调用智能体的Process方法
	hlog.Infof("发送消息到智能体: %s", testMsg.Subject)
	responseMsg, err := testAgent.Process(ctx, testMsg)
	if err != nil {
		hlog.Errorf("处理消息失败: %v", err)
	} else {
		hlog.Infof("收到响应: %s", responseMsg.Subject)
		hlog.Infof("响应内容: %s", responseMsg.Content)
		hlog.Infof("元数据: %v", responseMsg.Metadata)
	}
	
	// 6. 关闭智能体
	hlog.Infof("关闭智能体")
	testAgent.Shutdown(ctx)
}

// testOrchestrator 测试编排器功能
func testOrchestrator(ctx context.Context, testModel model.Model, registry *agenttools.ToolRegistry) {
	// 1. 创建编排器
	hlog.Infof("创建编排器")
	orchConfig := core.DefaultOrchestratorConfig()
	orchestrator := core.NewOrchestrator(orchConfig)
	
	// 2. 创建多个不同类型的智能体
	hlog.Infof("创建多个不同类型的智能体")
	
	// 创建世界观智能体
	worldviewAgent := core.NewGenericAdvancedAgent(
		"worldview_agent", 
		core.AgentTypeWorldview, 
		"你是世界观智能体，负责生成世界设定。",
	)
	worldviewAgent.SetModel(testModel)
	
	// 创建角色智能体
	characterAgent := core.NewGenericAdvancedAgent(
		"character_agent", 
		core.AgentTypeCharacter, 
		"你是角色智能体，负责生成角色信息。",
	)
	characterAgent.SetModel(testModel)
	
	// 3. 注册智能体到编排器
	hlog.Infof("注册智能体到编排器")
	err := orchestrator.RegisterAgent(worldviewAgent)
	if err != nil {
		hlog.Errorf("注册世界观智能体失败: %v", err)
		return
	}
	
	err = orchestrator.RegisterAgent(characterAgent)
	if err != nil {
		hlog.Errorf("注册角色智能体失败: %v", err)
		return
	}
	
	// 4. 测试获取智能体
	hlog.Infof("测试获取智能体")
	
	// 根据ID获取智能体
	agent, exists := orchestrator.GetAgent("worldview_agent")
	if exists {
		hlog.Infof("根据ID找到智能体: %s, 类型: %s", agent.GetID(), agent.GetType())
	} else {
		hlog.Errorf("根据ID找不到智能体")
	}
	
	// 根据类型获取智能体
	agents := orchestrator.GetAgentsByType(core.AgentTypeCharacter)
	hlog.Infof("找到 %d 个角色类型智能体", len(agents))
	
	// 5. 启动编排器
	hlog.Infof("启动编排器")
	err = orchestrator.Start()
	if err != nil {
		hlog.Errorf("启动编排器失败: %v", err)
		return
	}
	
	// 6. 发送消息并测试编排器消息路由
	hlog.Infof("发送消息到编排器")
	
	// 创建消息
	msg := core.NewMessage(core.MessageTypeRequest, "user", "worldview_agent")
	msg.Subject = "请创建一个奇幻世界设定"
	msg.Content = "请描述一个充满魔法的奇幻世界的基本设定。"
	
	// 发送消息
	hlog.Infof("发送消息: %s", msg.Subject)
	response, err := orchestrator.SendMessage(ctx, msg)
	if err != nil {
		hlog.Errorf("发送消息失败: %v", err)
	} else {
		hlog.Infof("收到响应: %s", response.Subject)
		hlog.Infof("响应内容: %s", response.Content)
	}
	
	// 7. 测试广播消息
	hlog.Infof("测试广播消息")
	broadcastMsg := core.NewMessage(core.MessageTypeNotification, "user", "")
	broadcastMsg.Subject = "系统通知"
	broadcastMsg.Content = "所有智能体注意，这是一条系统通知。"
	
	// 广播给所有的智能体
	responses, err := orchestrator.BroadcastMessage(ctx, core.AgentTypeWorldview, broadcastMsg)
	if err != nil {
		hlog.Errorf("广播消息失败: %v", err)
	} else {
		hlog.Infof("收到 %d 个响应", len(responses))
		for i, resp := range responses {
			hlog.Infof("响应 %d: %s", i+1, resp.Subject)
		}
	}
	
	// 8. 获取编排器状态
	status := orchestrator.GetStatus()
	hlog.Infof("编排器状态: %v", status)
	
	// 9. 停止编排器
	hlog.Infof("停止编排器")
	err = orchestrator.Stop()
	if err != nil {
		hlog.Errorf("停止编排器失败: %v", err)
	}
}
