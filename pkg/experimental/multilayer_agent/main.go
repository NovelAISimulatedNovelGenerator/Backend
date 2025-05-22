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

	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"
	"novelai/pkg/experimental/multilayer_agent/shared/tools/example_tool"
)

// 测试配置结构
type TestConfig struct {
	// 模型配置
	ModelType  model.ModelType `json:"model_type"`
	ModelName  string          `json:"model_name"`
	ModelURL   string          `json:"model_url,omitempty"`
	APIToken   string          `json:"api_token,omitempty"`
	MaxTokens  int             `json:"max_tokens,omitempty"`
	Debug      bool            `json:"debug"`
	
	// 测试参数
	RunModelTest      bool `json:"run_model_test"`
	RunToolTest       bool `json:"run_tool_test"`
	RunIntegrationTest bool `json:"run_integration_test"`
}

// 默认测试配置
var defaultConfig = TestConfig{
	ModelType:         model.ModelTypeOllama,
	ModelName:         "llama2",
	Debug:             true,
	RunModelTest:      true,
	RunToolTest:       true,
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
	exampleTool := createExampleTool()
	err = registry.RegisterTool(exampleTool)
	if err != nil {
		hlog.Errorf("注册示例工具失败: %v", err)
		os.Exit(1)
	}
	
	// 打印已注册工具
	tools := registry.ListTools()
	hlog.Infof("已注册 %d 个工具:", len(tools))
	for _, tool := range tools {
		hlog.Infof("  - %s: %s", tool.Name(), tool.Description())
	}
	
	// 运行模型测试
	if config.RunModelTest {
		hlog.Infof("=== 开始模型测试 ===")
		runModelTest(ctx, testModel)
	}
	
	// 运行工具测试
	if config.RunToolTest {
		hlog.Infof("=== 开始工具测试 ===")
		runToolTest(ctx, registry)
	}
	
	// 运行集成测试
	if config.RunIntegrationTest {
		hlog.Infof("=== 开始集成测试 ===")
		runIntegrationTest(ctx, testModel, registry)
	}
	
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
		ModelName:  config.ModelName,
		BaseURL:    config.ModelURL,
		APIToken:   config.APIToken,
		Debug:      config.Debug,
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
			Text:   "示例文本",
			Number: 42,
			Flag:   true,
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
