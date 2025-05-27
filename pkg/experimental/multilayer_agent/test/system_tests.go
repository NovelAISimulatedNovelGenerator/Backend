// package test 实现系统级别测试用例
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

// NovelGenerationTest 测试完整小说生成过程
type NovelGenerationTest struct {
	model          model.Model
	registry       *agenttools.ToolRegistry
	orchestrator   *core.Orchestrator
	agents         map[string]core.Agent
	generatedNovel map[string]string
}

// NewNovelGenerationTest 创建小说生成测试用例
func NewNovelGenerationTest() *NovelGenerationTest {
	return &NovelGenerationTest{
		agents:         make(map[string]core.Agent),
		generatedNovel: make(map[string]string),
	}
}

// Name 返回测试名称
func (t *NovelGenerationTest) Name() string {
	return "小说生成系统测试"
}

// Description 返回测试描述
func (t *NovelGenerationTest) Description() string {
	return "测试多层智能体系统生成完整小说的能力，包括世界观、角色、情节和格式化输出"
}

// Setup 设置测试环境
func (t *NovelGenerationTest) Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error {
	t.model = model
	t.registry = registry

	// 创建编排器
	config := core.DefaultOrchestratorConfig()
	// 增加超时时间，小说生成可能需要较长时间
	config.ProcessTimeout = 120 * time.Second
	t.orchestrator = core.NewOrchestrator(config)

	// 启动编排器
	err := t.orchestrator.Start()
	if err != nil {
		return fmt.Errorf("启动编排器失败: %v", err)
	}

	// 创建所有必要的智能体

	// 1. 策略层智能体
	strategyAgent := core.NewGenericAdvancedAgent(
		"strategy_agent",
		core.AgentTypeStrategy,
		"你是策略智能体，负责小说创作的整体规划和协调。你需要指导其他智能体完成各自的任务，并整合他们的成果。",
	)
	strategyAgent.SetModel(model)

	// 2. 评估智能体
	evaluatorAgent := core.NewGenericAdvancedAgent(
		"evaluator_agent",
		core.AgentTypeEvaluator,
		"你是评估智能体，负责评估其他智能体生成的内容质量，提供改进建议。",
	)
	evaluatorAgent.SetModel(model)

	// 3. 世界观智能体
	worldviewAgent := core.NewGenericAdvancedAgent(
		"worldview_agent",
		core.AgentTypeWorldview,
		"你是世界观智能体，负责创建小说的世界设定，包括环境、历史、社会结构和规则。",
	)
	worldviewAgent.SetModel(model)

	// 4. 角色智能体
	characterAgent := core.NewGenericAdvancedAgent(
		"character_agent",
		core.AgentTypeCharacter,
		"你是角色智能体，负责创建和发展小说中的人物，包括他们的特点、动机、关系和成长。",
	)
	characterAgent.SetModel(model)

	// 5. 剧情智能体
	plotAgent := core.NewGenericAdvancedAgent(
		"plot_agent",
		core.AgentTypePlot,
		"你是剧情智能体，负责设计故事的情节结构，包括冲突、转折和高潮。",
	)
	plotAgent.SetModel(model)

	// 6. 对话智能体
	dialogueAgent := core.NewGenericAdvancedAgent(
		"dialogue_agent",
		core.AgentTypeDialogue,
		"你是对话智能体，负责创建角色之间的对话，确保对话自然且符合角色特性。",
	)
	dialogueAgent.SetModel(model)

	// 7. 背景信息智能体
	backgroundAgent := core.NewGenericAdvancedAgent(
		"background_agent",
		core.AgentTypeBackground,
		"你是背景信息智能体，负责提供场景描述、氛围营造和环境细节。",
	)
	backgroundAgent.SetModel(model)

	// 8. 格式化智能体
	formatterAgent := core.NewGenericAdvancedAgent(
		"formatter_agent",
		core.AgentTypeFormatter,
		`你是一个小说格式化智能体，专门负责将各种素材整合成完整、连贯的小说作品。

特别注意：
1. 你必须生成纯文本形式的小说，结构包含标题和段落
2. 严格禁止生成JSON格式、XML格式或其他结构化数据格式
3. 不要使用花括号、中括号或大括号包裹整个输出
4. 生成的小说字数必须在1000-3000字之间
5. 小说应包含丰富的剧情和人物描写

请直接开始写作，不需要解释你的思考过程或其他元信息。`,
	)
	formatterAgent.SetModel(model)

	// 存储所有智能体
	agents := map[string]core.Agent{
		"strategy":   strategyAgent,
		"evaluator":  evaluatorAgent,
		"worldview":  worldviewAgent,
		"character":  characterAgent,
		"plot":       plotAgent,
		"dialogue":   dialogueAgent,
		"background": backgroundAgent,
		"formatter":  formatterAgent,
	}

	// 注册所有智能体到编排器
	for id, agent := range agents {
		hlog.Infof("注册智能体: %s (%s)", id, agent.GetType())
		err := t.orchestrator.RegisterAgent(agent)
		if err != nil {
			return fmt.Errorf("注册智能体 %s 失败: %v", id, err)
		}

		// 初始化智能体
		err = agent.Initialize(ctx)
		if err != nil {
			return fmt.Errorf("初始化智能体 %s 失败: %v", id, err)
		}

		t.agents[id] = agent
	}

	return nil
}

// Execute 执行测试
func (t *NovelGenerationTest) Execute() error {
	hlog.Infof("开始执行小说生成系统测试...")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// 1. 发送小说创作任务给策略智能体
	initialRequest := core.NewMessage(core.MessageTypeRequest, "user", "strategy_agent")
	initialRequest.Subject = "小说创作任务"
	initialRequest.Content = "请创作一个篇幅约2000字的短篇科幻小说，主题是'人工智能与人类共存'。"

	hlog.Infof("发送小说创作任务给策略智能体")
	strategyResp, err := t.orchestrator.SendMessage(ctx, initialRequest)
	if err != nil {
		return fmt.Errorf("发送策略消息失败: %v", err)
	}

	t.generatedNovel["strategy_plan"] = strategyResp.Content
	hlog.Infof("策略智能体制定了创作计划")

	// 2. 创建世界观
	worldviewRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "worldview_agent")
	worldviewRequest.Subject = "创建科幻世界观"
	worldviewRequest.Content = "请创建一个关于'人工智能与人类共存'的科幻世界观，描述这个世界的技术水平、社会结构和AI与人类关系。"

	hlog.Infof("发送世界观创建任务")
	worldviewResp, err := t.orchestrator.SendMessage(ctx, worldviewRequest)
	if err != nil {
		return fmt.Errorf("发送世界观消息失败: %v", err)
	}

	t.generatedNovel["worldview"] = worldviewResp.Content
	hlog.Infof("世界观智能体创建了世界设定")

	// 3. 评估世界观
	evalWorldviewRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "evaluator_agent")
	evalWorldviewRequest.Subject = "评估世界观"
	evalWorldviewRequest.Content = fmt.Sprintf("请评估以下世界观设定的质量和一致性:\n\n%s", worldviewResp.Content)

	hlog.Infof("发送世界观评估任务")
	evalWorldviewResp, err := t.orchestrator.SendMessage(ctx, evalWorldviewRequest)
	if err != nil {
		return fmt.Errorf("发送评估消息失败: %v", err)
	}

	t.generatedNovel["worldview_evaluation"] = evalWorldviewResp.Content
	hlog.Infof("评估智能体完成了世界观评估")

	// 4. 创建角色
	characterRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "character_agent")
	characterRequest.Subject = "创建角色"
	characterRequest.Content = fmt.Sprintf("基于以下世界观，创建2-3个主要角色，包括人类和AI:\n\n%s", worldviewResp.Content)

	hlog.Infof("发送角色创建任务")
	characterResp, err := t.orchestrator.SendMessage(ctx, characterRequest)
	if err != nil {
		return fmt.Errorf("发送角色消息失败: %v", err)
	}

	t.generatedNovel["characters"] = characterResp.Content
	hlog.Infof("角色智能体创建了故事角色")

	// 5. 创建剧情
	plotRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "plot_agent")
	plotRequest.Subject = "创建剧情"
	plotRequest.Content = fmt.Sprintf("基于以下世界观和角色，创建故事剧情框架:\n\n世界观:\n%s\n\n角色:\n%s",
		worldviewResp.Content, characterResp.Content)

	hlog.Infof("发送剧情创建任务")
	plotResp, err := t.orchestrator.SendMessage(ctx, plotRequest)
	if err != nil {
		return fmt.Errorf("发送剧情消息失败: %v", err)
	}

	t.generatedNovel["plot"] = plotResp.Content
	hlog.Infof("剧情智能体创建了故事情节")

	// 6. 创建背景描述
	backgroundRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "background_agent")
	backgroundRequest.Subject = "创建场景描述"
	backgroundRequest.Content = fmt.Sprintf("基于以下世界观和剧情，创建关键场景的背景描述:\n\n世界观:\n%s\n\n剧情:\n%s",
		worldviewResp.Content, plotResp.Content)

	hlog.Infof("发送背景描述创建任务")
	backgroundResp, err := t.orchestrator.SendMessage(ctx, backgroundRequest)
	if err != nil {
		return fmt.Errorf("发送背景消息失败: %v", err)
	}

	t.generatedNovel["background"] = backgroundResp.Content
	hlog.Infof("背景信息智能体创建了场景描述")

	// 7. 创建对话
	dialogueRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "dialogue_agent")
	dialogueRequest.Subject = "创建对话"
	dialogueRequest.Content = fmt.Sprintf("基于以下角色和剧情，创建关键场景的对话:\n\n角色:\n%s\n\n剧情:\n%s",
		characterResp.Content, plotResp.Content)

	hlog.Infof("发送对话创建任务")
	dialogueResp, err := t.orchestrator.SendMessage(ctx, dialogueRequest)
	if err != nil {
		return fmt.Errorf("发送对话消息失败: %v", err)
	}

	t.generatedNovel["dialogue"] = dialogueResp.Content
	hlog.Infof("对话智能体创建了角色对话")

	// 8. 最终整合和格式化
	formatterRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "formatter_agent")
	formatterRequest.Subject = "整合和格式化小说"
	formatterRequest.Content = fmt.Sprintf(`请将以下内容整合成一篇完整、连贯的科幻短篇小说。

非常重要：
1. 输出必须是纯文本形式的小说，结构包含标题和段落，不要生成JSON或其他结构化数据格式
2. 最终生成的小说字数必须在 1000-3000 字之间
3. 确保内容丰富、详尽，包含充分的剧情发展、人物对话和环境描写
4. 注意小说的连贯性和可读性，确保情节自然过渡

请根据以下素材创作：

世界观:
%s

角色:
%s

剧情:
%s

背景描述:
%s

对话:
%s`,
		worldviewResp.Content, characterResp.Content, plotResp.Content, backgroundResp.Content, dialogueResp.Content)

	hlog.Infof("发送最终整合任务")
	formatterResp, err := t.orchestrator.SendMessage(ctx, formatterRequest)
	if err != nil {
		return fmt.Errorf("发送格式化消息失败: %v", err)
	}

	t.generatedNovel["final_novel"] = formatterResp.Content
	hlog.Infof("格式化智能体完成了最终小说整合")
	
	// 记录最终生成的小说内容
	hlog.Infof("最终生成小说内容：\n%s", formatterResp.Content)
	
	// 记录字数统计
	wordCount := len([]rune(formatterResp.Content))
	hlog.Infof("小说字数统计：%d 字符", wordCount)

	// 9. 最终评估
	finalEvalRequest := core.NewMessage(core.MessageTypeRequest, "strategy_agent", "evaluator_agent")
	finalEvalRequest.Subject = "评估最终小说"
	finalEvalRequest.Content = fmt.Sprintf("请评估以下科幻短篇小说的质量:\n\n%s", formatterResp.Content)

	hlog.Infof("发送最终评估任务")
	finalEvalResp, err := t.orchestrator.SendMessage(ctx, finalEvalRequest)
	if err != nil {
		return fmt.Errorf("发送最终评估消息失败: %v", err)
	}

	t.generatedNovel["final_evaluation"] = finalEvalResp.Content
	hlog.Infof("评估智能体完成了最终小说评估")

	return nil
}

// Verify 验证测试结果
func (t *NovelGenerationTest) Verify() (bool, string) {
	// 检查是否生成了所有必要的内容
	requiredComponents := []string{
		"worldview", "characters", "plot", "background", "dialogue", "final_novel",
	}

	for _, component := range requiredComponents {
		if content, exists := t.generatedNovel[component]; !exists || content == "" {
			return false, fmt.Sprintf("缺少 %s 组件", component)
		}
	}

	// 验证最终小说是否符合要求
	finalNovel := t.generatedNovel["final_novel"]

	// 检查字数是否在合理范围内（允许一定误差）
	wordCount := len([]rune(finalNovel)) // 使用rune计算字符数，适合中文
	if wordCount < 1000 || wordCount > 3000 {
		return false, fmt.Sprintf("最终小说字数不符合要求: %d 字符 (应为1000-3000字符)", wordCount)
	}

	// 检查是否包含所有必要元素
	// 这里只是简单检查小说是否提到了各个组件中的一些关键信息
	// 实际测试中可能需要更复杂的语义分析

	hlog.Infof("生成的小说字数: %d", wordCount)

	return true, ""
}

// Cleanup 清理测试资源
func (t *NovelGenerationTest) Cleanup() {
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
