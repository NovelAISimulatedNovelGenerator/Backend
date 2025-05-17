package background

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	model "novelai/biz/model/background"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/ollama/ollama/api"
)

// OllamaConfig Ollama API的配置参数
type OllamaConfig struct {
	Model       string  `json:"model"`       // Ollama模型名称
	Temperature float32 `json:"temperature"` // 生成温度，默认0.7
	MaxTokens   int     `json:"max_tokens"`  // 最大生成令牌数，默认2048
}

// NewDefaultOllamaConfig 创建默认Ollama配置
//
// 工作流程:
// 1. 创建OllamaConfig结构体实例
// 2. 设置预定义的默认值
//
// 返回值:
//   - *OllamaConfig: 使用默认值初始化的配置
//
// 注意事项:
//   - 默认使用deepseek-r1:14b模型
//   - 默认温度为0.7
//   - 默认最大令牌数为2048
func NewDefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		Model:       "deepseek-r1:14b", // 默认使用deepseek-r1模型
		Temperature: 0.7,               // 默认温度
		MaxTokens:   2048,              // 默认最大令牌数
	}
}

// NewOllamaConfig 从JSON字符串创建Ollama配置
//
// 工作流程:
// 1. 创建一个默认配置
// 2. 如果提供了配置JSON字符串，则解析并覆盖默认值
//
// 参数:
//   - configJSON: 包含配置参数的JSON字符串，可以为空
//
// 返回值:
//   - *OllamaConfig: 配置对象
//   - error: 解析错误，如果JSON格式无效
//
// 注意事项:
//   - 如果configJSON为空，将返回默认配置
//   - 部分参数配置无效不会导致错误，只会使用默认值
func NewOllamaConfig(configJSON string) (*OllamaConfig, error) {
	config := NewDefaultOllamaConfig()
	
	if configJSON == "" {
		return config, nil
	}
	
	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return nil, fmt.Errorf("解析Ollama配置失败: %v", err)
	}
	
	return config, nil
}

// GenerateWorldviewWithOllama 使用Ollama API生成世界观
//
// 工作流程:
// 1. 创建Ollama客户端
// 2. 构建提示词，要求生成JSON格式的世界观信息
// 3. 调用Ollama API生成内容
// 4. 解析响应并创建世界观对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: Ollama API配置参数
//   - theme: 可选主题，指定世界观的主题，为空则随机生成
//
// 返回值:
//   - *model.Worldview: 生成的世界观对象
//   - error: 生成过程中的错误
//
// 注意事项:
//   - 返回的世界观对象不包含ID、创建时间等字段，需要后续保存到数据库
//   - 提示词指定了严格的JSON输出格式，以便解析
//   - 如果提供了主题，会生成与主题相关的世界观
func GenerateWorldviewWithOllama(ctx context.Context, config *OllamaConfig, theme string) (*model.Worldview, error) {
	// 创建Ollama客户端
	client, err := api.ClientFromEnvironment()
	if err != nil {
		hlog.CtxErrorf(ctx, "创建Ollama客户端失败: %v", err)
		return nil, fmt.Errorf("创建Ollama客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始使用Ollama生成世界观，模型: %s", config.Model)
	
	// 构建提示词
	var prompt string
	if theme != "" {
		hlog.CtxInfof(ctx, "基于主题'%s'生成世界观", theme)
		prompt = fmt.Sprintf(
			"你是一个小说世界观生成助手，请生成一个与'%s'相关的世界观，包括名称、描述、标签。"+
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。",
			theme,
		)
	} else {
		prompt = "你是一个小说世界观生成助手，请生成一个有创意的世界观，包括名称、描述、标签。" +
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。"
	}
	
	// 创建请求
	req := &api.GenerateRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: new(bool), // 非流式输出
		Format: json.RawMessage(`"json"`),
	}

	var jsonResponse string
	// 定义响应处理函数
	respFunc := func(resp api.GenerateResponse) error {
		jsonResponse = resp.Response
		hlog.CtxInfof(ctx, "世界观生成原始响应: %s", jsonResponse)
		return nil
	}

	// 调用API
	if err := client.Generate(ctx, req, respFunc); err != nil {
		hlog.CtxErrorf(ctx, "调用Ollama API生成世界观失败: %v", err)
		return nil, err
	}

	// 解析JSON响应
	var result struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Tag         string `json:"tag"`
	}

	// 清理可能的格式问题
	jsonResponse = cleanJsonResponse(jsonResponse)

	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		hlog.CtxErrorf(ctx, "解析世界观JSON响应失败: %v, 原始响应: %s", err, jsonResponse)
		return nil, fmt.Errorf("解析世界观JSON响应失败: %v", err)
	}

	// 创建世界观对象
	worldview := &model.Worldview{
		Name:        result.Name,
		Description: result.Description,
		Tag:         result.Tag,
	}

	hlog.CtxInfof(ctx, "成功生成世界观: %s", worldview.Name)
	return worldview, nil
}

// GenerateRuleWithOllama 使用Ollama API生成规则
//
// 工作流程:
// 1. 验证世界观参数
// 2. 创建Ollama客户端
// 3. 构建基于世界观的提示词
// 4. 调用Ollama API生成内容
// 5. 解析响应并创建规则对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: Ollama API配置参数
//   - worldview: 作为规则基础的世界观对象
//   - ruleType: 可选规则类型，指定规则的特定类型或方向，为空则基于世界观自动生成
//
// 返回值:
//   - *model.Rule: 生成的规则对象
//   - error: 生成过程中的错误
//
// 注意事项:
//   - 返回的规则对象已自动关联到传入的世界观ID
//   - 规则生成基于世界观的名称和描述，以保持一致性
//   - 如果提供了规则类型，会生成符合该类型的规则
func GenerateRuleWithOllama(ctx context.Context, config *OllamaConfig, worldview *model.Worldview, ruleType string) (*model.Rule, error) {
	if worldview == nil {
		return nil, fmt.Errorf("世界观不能为空")
	}
	
	// 创建Ollama客户端
	client, err := api.ClientFromEnvironment()
	if err != nil {
		hlog.CtxErrorf(ctx, "创建Ollama客户端失败: %v", err)
		return nil, fmt.Errorf("创建Ollama客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始为世界观 '%s' 生成规则...", worldview.Name)
	
	// 构建提示词
	var prompt string
	if ruleType != "" {
		hlog.CtxInfof(ctx, "基于规则类型'%s'为世界观'%s'生成规则", ruleType, worldview.Name)
		prompt = fmt.Sprintf(
			"你是一个小说规则生成助手，请为以下世界观生成一个'%s'类型的规则，包括名称、描述、标签。\n\n" +
			"世界观: %s\n" +
			"世界观描述: %s\n\n" +
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。",
			ruleType, worldview.Name, worldview.Description,
		)
	} else {
		prompt = fmt.Sprintf(
			"你是一个小说规则生成助手，请为以下世界观生成一个规则，包括名称、描述、标签。\n\n" +
			"世界观: %s\n" +
			"世界观描述: %s\n\n" +
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。",
			worldview.Name, worldview.Description,
		)
	}
	
	// 创建请求
	req := &api.GenerateRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: new(bool), // 非流式输出
		Format: json.RawMessage(`"json"`),
	}

	var jsonResponse string
	// 定义响应处理函数
	respFunc := func(resp api.GenerateResponse) error {
		jsonResponse = resp.Response
		hlog.CtxInfof(ctx, "规则生成原始响应: %s", jsonResponse)
		return nil
	}

	// 调用API
	if err := client.Generate(ctx, req, respFunc); err != nil {
		hlog.CtxErrorf(ctx, "调用Ollama API生成规则失败: %v", err)
		return nil, err
	}

	// 解析JSON响应
	var result struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Tag         string `json:"tag"`
	}

	// 清理可能的格式问题
	jsonResponse = cleanJsonResponse(jsonResponse)

	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		hlog.CtxErrorf(ctx, "解析规则JSON响应失败: %v, 原始响应: %s", err, jsonResponse)
		return nil, fmt.Errorf("解析规则JSON响应失败: %v", err)
	}

	// 创建规则对象
	rule := &model.Rule{
		WorldviewId: worldview.Id,
		Name:        result.Name,
		Description: result.Description,
		Tag:         result.Tag,
	}

	hlog.CtxInfof(ctx, "成功生成规则: %s", rule.Name)
	return rule, nil
}

// GenerateBackgroundInfoWithOllama 使用Ollama API生成背景信息
//
// 工作流程:
// 1. 验证世界观和规则参数
// 2. 创建Ollama客户端
// 3. 构建基于世界观和规则的提示词
// 4. 调用Ollama API生成内容
// 5. 解析响应并创建背景信息对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: Ollama API配置参数
//   - worldview: 作为背景信息基础的世界观对象
//   - rule: 作为背景信息基础的规则对象
//   - character: 可选角色描述，为背景信息提供特定角色方向，为空则随机生成
//
// 返回值:
//   - *model.BackgroundInfo: 生成的背景信息对象
//   - error: 生成过程中的错误
//
// 注意事项:
//   - 返回的背景信息对象已自动关联到传入的世界观ID
//   - 背景信息生成基于世界观和规则的内容，以保持一致性
//   - 如果提供了角色描述，生成的背景信息将围绕该角色展开
func GenerateBackgroundInfoWithOllama(ctx context.Context, config *OllamaConfig, worldview *model.Worldview, rule *model.Rule, character string) (*model.BackgroundInfo, error) {
	if worldview == nil {
		return nil, fmt.Errorf("世界观不能为空")
	}
	
	if rule == nil {
		return nil, fmt.Errorf("规则不能为空")
	}
	
	// 创建Ollama客户端
	client, err := api.ClientFromEnvironment()
	if err != nil {
		hlog.CtxErrorf(ctx, "创建Ollama客户端失败: %v", err)
		return nil, fmt.Errorf("创建Ollama客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始为世界观 '%s' 和规则 '%s' 生成背景信息...", worldview.Name, rule.Name)
	
	// 构建提示词
	var prompt string
	if character != "" {
		hlog.CtxInfof(ctx, "基于角色'%s'为世界观'%s'和规则'%s'生成背景信息", character, worldview.Name, rule.Name)
		prompt = fmt.Sprintf(
			"你是一个小说背景信息生成助手，请根据以下世界观和规则生成一个与'%s'相关的背景信息，包括名称、描述、标签。\n\n" +
			"世界观: %s\n" +
			"世界观描述: %s\n" +
			"规则: %s\n" +
			"规则描述: %s\n\n" +
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。",
			character, worldview.Name, worldview.Description, rule.Name, rule.Description,
		)
	} else {
		prompt = fmt.Sprintf(
			"你是一个小说背景信息生成助手，请根据以下世界观和规则生成一个背景信息，包括名称、描述、标签。\n\n" +
			"世界观: %s\n" +
			"世界观描述: %s\n" +
			"规则: %s\n" +
			"规则描述: %s\n\n" +
			"请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}" +
			"不要输出除JSON以外的内容。",
			worldview.Name, worldview.Description, rule.Name, rule.Description,
		)
	}
	
	// 创建请求
	req := &api.GenerateRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: new(bool), // 非流式输出
		Format: json.RawMessage(`"json"`),
	}

	var jsonResponse string
	// 定义响应处理函数
	respFunc := func(resp api.GenerateResponse) error {
		jsonResponse = resp.Response
		hlog.CtxInfof(ctx, "背景信息生成原始响应: %s", jsonResponse)
		return nil
	}

	// 调用API
	if err := client.Generate(ctx, req, respFunc); err != nil {
		hlog.CtxErrorf(ctx, "调用Ollama API生成背景信息失败: %v", err)
		return nil, err
	}

	// 解析JSON响应
	var result struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Tag         string `json:"tag"`
	}

	// 清理可能的格式问题
	jsonResponse = cleanJsonResponse(jsonResponse)

	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		hlog.CtxErrorf(ctx, "解析背景信息JSON响应失败: %v, 原始响应: %s", err, jsonResponse)
		return nil, fmt.Errorf("解析背景信息JSON响应失败: %v", err)
	}

	// 创建背景信息对象
	backgroundInfo := &model.BackgroundInfo{
		WorldviewId: worldview.Id,
		Name:        result.Name,
		Description: result.Description,
		Tag:         result.Tag,
	}

	hlog.CtxInfof(ctx, "成功生成背景信息: %s", backgroundInfo.Name)
	return backgroundInfo, nil
}

// SaveGeneratedContent 保存生成的内容到数据库
//
// 工作流程:
// 1. 创建服务实例
// 2. 保存世界观，获取新的ID
// 3. 保存规则，更新世界观ID并获取新的ID
// 4. 保存背景信息，更新世界观ID
//
// 参数:
//   - ctx: 上下文，用于服务实例创建和日志记录
//   - c: 请求上下文，用于服务实例创建
//   - worldview: 要保存的世界观对象
//   - rule: 要保存的规则对象
//   - backgroundInfo: 要保存的背景信息对象
//
// 返回值:
//   - error: 保存过程中的错误
//
// 注意事项:
//   - 函数会自动更新世界观ID的关联关系
//   - 传入的对象会被更新为保存后的状态，包含数据库分配的ID
func SaveGeneratedContent(ctx context.Context, c *app.RequestContext, worldview *model.Worldview, rule *model.Rule, backgroundInfo *model.BackgroundInfo) error {
	// 创建服务实例
	worldviewService := NewWorldviewService(ctx, c)
	ruleService := NewRuleService(ctx, c)
	backgroundInfoService := NewBackgroundInfoService(ctx, c)
	
	hlog.CtxInfof(ctx, "开始保存生成的内容...")
	
	// 保存世界观
	createdWorldview, err := worldviewService.CreateWorldview(&model.CreateWorldviewRequest{
		Name:        worldview.Name,
		Description: worldview.Description,
		Tag:         worldview.Tag,
		ParentId:    worldview.ParentId,
	})
	
	if err != nil {
		hlog.CtxErrorf(ctx, "保存世界观到数据库失败: %v", err)
		return fmt.Errorf("保存世界观到数据库失败: %v", err)
	}
	
	// 更新世界观ID
	worldview.Id = createdWorldview.Id
	hlog.CtxInfof(ctx, "成功保存世界观 [ID: %d] '%s'", worldview.Id, worldview.Name)
	
	// 保存规则
	rule.WorldviewId = worldview.Id // 确保规则关联到正确的世界观ID
	
	createdRule, err := ruleService.CreateRule(&model.CreateRuleRequest{
		WorldviewId: rule.WorldviewId,
		Name:        rule.Name,
		Description: rule.Description,
		Tag:         rule.Tag,
		ParentId:    rule.ParentId,
	})
	
	if err != nil {
		hlog.CtxErrorf(ctx, "保存规则到数据库失败: %v", err)
		return fmt.Errorf("保存规则到数据库失败: %v", err)
	}
	
	// 更新规则ID
	rule.Id = createdRule.Id
	hlog.CtxInfof(ctx, "成功保存规则 [ID: %d] '%s'", rule.Id, rule.Name)
	
	// 保存背景信息
	backgroundInfo.WorldviewId = worldview.Id // 确保背景信息关联到正确的世界观ID
	
	createdBgInfo, err := backgroundInfoService.CreateBackgroundInfo(&model.CreateBackgroundInfoRequest{
		WorldviewId: backgroundInfo.WorldviewId,
		Name:        backgroundInfo.Name,
		Description: backgroundInfo.Description,
		Tag:         backgroundInfo.Tag,
		ParentId:    backgroundInfo.ParentId,
	})
	
	if err != nil {
		hlog.CtxErrorf(ctx, "保存背景信息到数据库失败: %v", err)
		return fmt.Errorf("保存背景信息到数据库失败: %v", err)
	}
	
	// 更新背景信息ID
	backgroundInfo.Id = createdBgInfo.Id
	hlog.CtxInfof(ctx, "成功保存背景信息 [ID: %d] '%s'", backgroundInfo.Id, backgroundInfo.Name)
	
	hlog.CtxInfof(ctx, "所有内容已成功保存到数据库")
	return nil
}

// GenerateAndSaveWithOllama 一站式生成并保存小说背景内容
//
// 工作流程:
// 1. 解析配置
// 2. 生成世界观
// 3. 生成规则
// 4. 生成背景信息
// 5. 保存所有内容到数据库
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - c: 请求上下文，用于服务实例创建
//   - configJSON: Ollama配置的JSON字符串
//   - theme: 可选主题，指定世界观的主题
//   - ruleType: 可选规则类型，指定规则的特定类型
//   - character: 可选角色描述，为背景信息提供特定角色方向
//
// 返回值:
//   - *model.Worldview: 生成并保存的世界观
//   - *model.Rule: 生成并保存的规则
//   - *model.BackgroundInfo: 生成并保存的背景信息
//   - error: 生成或保存过程中的错误
//
// 注意事项:
//   - 返回的对象包含数据库分配的ID和关联关系
//   - 如果任何一步失败，将返回错误
//   - 可以通过传入参数控制生成内容的方向和类型
func GenerateAndSaveWithOllama(ctx context.Context, c *app.RequestContext, configJSON string, theme string, ruleType string, character string) (*model.Worldview, *model.Rule, *model.BackgroundInfo, error) {
	// 解析配置
	config, err := NewOllamaConfig(configJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	
	hlog.CtxInfof(ctx, "开始使用Ollama生成小说背景内容，模型: %s", config.Model)
	
	// 生成世界观
	hlog.CtxInfof(ctx, "开始生成世界观...")
	worldview, err := GenerateWorldviewWithOllama(ctx, config, theme)
	if err != nil {
		return nil, nil, nil, err
	}
	
	// 生成规则
	hlog.CtxInfof(ctx, "开始生成规则...")
	rule, err := GenerateRuleWithOllama(ctx, config, worldview, ruleType)
	if err != nil {
		return worldview, nil, nil, err
	}
	
	// 生成背景信息
	hlog.CtxInfof(ctx, "开始生成背景信息...")
	backgroundInfo, err := GenerateBackgroundInfoWithOllama(ctx, config, worldview, rule, character)
	if err != nil {
		return worldview, rule, nil, err
	}
	
	// 保存所有内容
	err = SaveGeneratedContent(ctx, c, worldview, rule, backgroundInfo)
	if err != nil {
		return worldview, rule, backgroundInfo, err
	}
	
	hlog.CtxInfof(ctx, "成功生成并保存小说背景内容")
	return worldview, rule, backgroundInfo, nil
}

// 内部助手函数

// cleanJsonResponse 清理JSON响应字符串，修复常见的格式问题
func cleanJsonResponse(response string) string {
	response = strings.TrimSpace(response)
	
	// 处理可能的Markdown代码块
	if strings.HasPrefix(response, "```") {
		// 寻找第一个和最后一个 ``` 标记
		firstMark := strings.Index(response, "```")
		lastMark := strings.LastIndex(response, "```")
		
		if firstMark != lastMark {
			// 提取代码块内容
			content := response[firstMark+3:lastMark]
			
			// 如果第一行是语言标识(如json)，则去掉这一行
			content = strings.TrimSpace(content)
			if strings.HasPrefix(content, "json") {
				content = strings.TrimPrefix(content, "json")
			}
			
			response = strings.TrimSpace(content)
		}
	}
	
	return response
}
