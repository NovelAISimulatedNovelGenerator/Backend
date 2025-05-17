package background

import (
	"context"
	"encoding/json"
	"fmt"

	model "novelai/biz/model/background"
	"novelai/pkg/constants"
	"novelai/pkg/llm/deepseek"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// DeepSeekConfig DeepSeek API的配置参数
type DeepSeekConfig struct {
	Model       string  `json:"model"`       // DeepSeek模型名称
	Temperature float64 `json:"temperature"` // 生成温度，默认0.7
	MaxTokens   int     `json:"max_tokens"`  // 最大生成令牌数，默认2048
	APIKey      string  `json:"api_key"`     // DeepSeek API密钥
}

// NewDefaultDeepSeekConfig 创建默认DeepSeek配置
//
// 工作流程:
// 1. 创建DeepSeekConfig结构体实例
// 2. 设置预定义的默认值
//
// 返回值:
//   - *DeepSeekConfig: 使用默认值初始化的配置
//
// 注意事项:
//   - 默认使用deepseek-chat模型
//   - 默认温度为0.7
//   - 默认最大令牌数为2048
//   - 需要通过环境变量或者配置传入API密钥
func NewDefaultDeepSeekConfig() *DeepSeekConfig {
	return &DeepSeekConfig{
		Model:       constants.DeepSeekChat, // 默认使用deepseek-chat模型
		Temperature: 0.7,                    // 默认温度
		MaxTokens:   2048,                   // 默认最大令牌数
	}
}

// NewDeepSeekConfig 从JSON字符串创建DeepSeek配置
//
// 工作流程:
// 1. 创建一个默认配置
// 2. 如果提供了配置JSON字符串，则解析并覆盖默认值
//
// 参数:
//   - configJSON: 包含配置参数的JSON字符串，可以为空
//
// 返回值:
//   - *DeepSeekConfig: 配置对象
//   - error: 解析错误，如果JSON格式无效
//
// 注意事项:
//   - 如果configJSON为空，将返回默认配置
//   - 部分参数配置无效不会导致错误，只会使用默认值
//   - API密钥必须通过环境变量或配置提供
func NewDeepSeekConfig(configJSON string) (*DeepSeekConfig, error) {
	config := NewDefaultDeepSeekConfig()
	
	if configJSON == "" {
		return config, nil
	}
	
	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return nil, fmt.Errorf("解析DeepSeek配置失败: %v", err)
	}
	
	return config, nil
}

// GenerateWorldviewWithDeepSeek 使用DeepSeek API生成世界观
//
// 工作流程:
// 1. 创建DeepSeek客户端
// 2. 构建提示词，要求生成JSON格式的世界观信息
// 3. 调用DeepSeek API生成内容
// 4. 解析响应并创建世界观对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: DeepSeek API配置参数
//   - theme: 可选主题，指定世界观的主题，为空则随机生成
//
// 返回值:
//   - *model.Worldview: 生成的世界观对象
//   - error: 生成过程中的错误
//
// 注意事项:
//   - 返回的世界观对象不包含ID、创建时间等字段，需要后续保存到数据库
//   - 提示词指定了严格的JSON输出格式，以便解析
//   - 需要有效的API密钥才能成功调用
//   - 如果提供了主题，会生成与主题相关的世界观
func GenerateWorldviewWithDeepSeek(ctx context.Context, config *DeepSeekConfig, theme string) (*model.Worldview, error) {
	// 检查API密钥
	if config.APIKey == "" {
		hlog.CtxErrorf(ctx, "DeepSeek API密钥不能为空")
		return nil, fmt.Errorf("DeepSeek API密钥不能为空")
	}

	// 创建DeepSeek客户端
	client, err := deepseek.NewClient(config.APIKey)
	if err != nil {
		hlog.CtxErrorf(ctx, "创建DeepSeek客户端失败: %v", err)
		return nil, fmt.Errorf("创建DeepSeek客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始使用DeepSeek生成世界观，模型: %s", config.Model)
	
	// 构建消息
	msgBuilder := deepseek.NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个小说世界观生成助手，请生成一个有创意的世界观，包括名称、描述、标签。请严格按照JSON格式输出。")
	
	var prompt string
	if theme != "" {
		hlog.CtxInfof(ctx, "基于主题'%s'生成世界观", theme)
		prompt = fmt.Sprintf(
			"请生成一个与'%s'相关的小说世界观，包括名称、描述、标签。请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}",
			theme,
		)
	} else {
		prompt = "请生成一个有创意的小说世界观，包括名称、描述、标签。请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}"
	}
	
	msgBuilder.AddUserMessage(prompt)
	
	// 创建聊天请求
	req := &deepseek.ChatRequest{
		Model:       config.Model,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		ResponseFormat: deepseek.ResponseFormat{
			Type: "json_object",
		},
	}

	// 调用API
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		hlog.CtxErrorf(ctx, "调用DeepSeek API生成世界观失败: %v", err)
		return nil, err
	}

	// 提取回复内容
	var jsonResponse string
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					jsonResponse = content
				}
			}
		}
	}

	hlog.CtxInfof(ctx, "世界观生成原始响应: %s", jsonResponse)

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

// GenerateRuleWithDeepSeek 使用DeepSeek API生成规则
//
// 工作流程:
// 1. 验证世界观参数
// 2. 创建DeepSeek客户端
// 3. 构建基于世界观的提示词
// 4. 调用DeepSeek API生成内容
// 5. 解析响应并创建规则对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: DeepSeek API配置参数
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
//   - 需要有效的API密钥才能成功调用
//   - 如果提供了规则类型，会生成符合该类型的规则
func GenerateRuleWithDeepSeek(ctx context.Context, config *DeepSeekConfig, worldview *model.Worldview, ruleType string) (*model.Rule, error) {
	if worldview == nil {
		return nil, fmt.Errorf("世界观不能为空")
	}
	
	// 检查API密钥
	if config.APIKey == "" {
		hlog.CtxErrorf(ctx, "DeepSeek API密钥不能为空")
		return nil, fmt.Errorf("DeepSeek API密钥不能为空")
	}

	// 创建DeepSeek客户端
	client, err := deepseek.NewClient(config.APIKey)
	if err != nil {
		hlog.CtxErrorf(ctx, "创建DeepSeek客户端失败: %v", err)
		return nil, fmt.Errorf("创建DeepSeek客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始为世界观 '%s' 生成规则...", worldview.Name)
	
	// 构建消息
	msgBuilder := deepseek.NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个小说规则生成助手，需要为特定世界观生成合适的规则。")
	
	var prompt string
	if ruleType != "" {
		hlog.CtxInfof(ctx, "基于规则类型'%s'为世界观'%s'生成规则", ruleType, worldview.Name)
		prompt = fmt.Sprintf(
			"请为以下世界观生成一个'%s'类型的规则，包括名称、描述、标签。\n\n"+
			"世界观: %s\n"+
			"世界观描述: %s\n\n"+
			"请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}",
			ruleType, worldview.Name, worldview.Description,
		)
	} else {
		prompt = fmt.Sprintf(
			"请为以下世界观生成一个规则，包括名称、描述、标签。\n\n"+
			"世界观: %s\n"+
			"世界观描述: %s\n\n"+
			"请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}",
			worldview.Name, worldview.Description,
		)
	}
	msgBuilder.AddUserMessage(prompt)
	
	// 创建聊天请求
	req := &deepseek.ChatRequest{
		Model:       config.Model,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		ResponseFormat: deepseek.ResponseFormat{
			Type: "json_object",
		},
	}

	// 调用API
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		hlog.CtxErrorf(ctx, "调用DeepSeek API生成规则失败: %v", err)
		return nil, err
	}

	// 提取回复内容
	var jsonResponse string
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					jsonResponse = content
				}
			}
		}
	}

	hlog.CtxInfof(ctx, "规则生成原始响应: %s", jsonResponse)

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
		Name:        result.Name,
		Description: result.Description,
		Tag:         result.Tag,
		WorldviewId: worldview.Id,  // 关联到世界观
	}

	hlog.CtxInfof(ctx, "成功生成规则: %s", rule.Name)
	return rule, nil
}

// GenerateBackgroundInfoWithDeepSeek 使用DeepSeek API生成背景信息
//
// 工作流程:
// 1. 验证世界观和规则参数
// 2. 创建DeepSeek客户端
// 3. 构建基于世界观和规则的提示词
// 4. 调用DeepSeek API生成内容
// 5. 解析响应并创建背景信息对象
//
// 参数:
//   - ctx: 上下文，用于控制API调用和日志记录
//   - config: DeepSeek API配置参数
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
//   - 需要有效的API密钥才能成功调用
//   - 如果提供了角色描述，生成的背景信息将围绕该角色展开
func GenerateBackgroundInfoWithDeepSeek(ctx context.Context, config *DeepSeekConfig, worldview *model.Worldview, rule *model.Rule, character string) (*model.BackgroundInfo, error) {
	if worldview == nil {
		return nil, fmt.Errorf("世界观不能为空")
	}
	
	if rule == nil {
		return nil, fmt.Errorf("规则不能为空")
	}
	
	// 检查API密钥
	if config.APIKey == "" {
		hlog.CtxErrorf(ctx, "DeepSeek API密钥不能为空")
		return nil, fmt.Errorf("DeepSeek API密钥不能为空")
	}

	// 创建DeepSeek客户端
	client, err := deepseek.NewClient(config.APIKey)
	if err != nil {
		hlog.CtxErrorf(ctx, "创建DeepSeek客户端失败: %v", err)
		return nil, fmt.Errorf("创建DeepSeek客户端失败: %v", err)
	}

	hlog.CtxInfof(ctx, "开始为世界观 '%s' 和规则 '%s' 生成背景信息...", worldview.Name, rule.Name)
	
	// 构建消息
	msgBuilder := deepseek.NewMessageBuilder()
	msgBuilder.AddSystemMessage("你是一个小说背景信息生成助手，需要为特定世界观和规则生成合适的背景信息。")
	
	var prompt string
	if character != "" {
		hlog.CtxInfof(ctx, "基于角色'%s'为世界观'%s'和规则'%s'生成背景信息", character, worldview.Name, rule.Name)
		prompt = fmt.Sprintf(
			"请根据以下世界观和规则生成一个与'%s'相关的背景信息，包括名称、描述、标签。\n\n"+
			"世界观: %s\n"+
			"世界观描述: %s\n"+
			"规则: %s\n"+
			"规则描述: %s\n\n"+
			"请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}",
			character, worldview.Name, worldview.Description, rule.Name, rule.Description,
		)
	} else {
		prompt = fmt.Sprintf(
			"请根据以下世界观和规则生成一个背景信息，包括名称、描述、标签。\n\n"+
			"世界观: %s\n"+
			"世界观描述: %s\n"+
			"规则: %s\n"+
			"规则描述: %s\n\n"+
			"请严格按照如下JSON格式输出，不要输出除JSON以外的内容：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}",
			worldview.Name, worldview.Description, rule.Name, rule.Description,
		)
	}
	msgBuilder.AddUserMessage(prompt)
	
	// 创建聊天请求
	req := &deepseek.ChatRequest{
		Model:       config.Model,
		Messages:    msgBuilder.Messages(),
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		ResponseFormat: deepseek.ResponseFormat{
			Type: "json_object",
		},
	}

	// 调用API
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		hlog.CtxErrorf(ctx, "调用DeepSeek API生成背景信息失败: %v", err)
		return nil, err
	}

	// 提取回复内容
	var jsonResponse string
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					jsonResponse = content
				}
			}
		}
	}

	hlog.CtxInfof(ctx, "背景信息生成原始响应: %s", jsonResponse)

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
		Name:        result.Name,
		Description: result.Description,
		Tag:         result.Tag,
		WorldviewId: worldview.Id,  // 关联到世界观
	}

	hlog.CtxInfof(ctx, "成功生成背景信息: %s", backgroundInfo.Name)
	return backgroundInfo, nil
}

// GenerateAndSaveWithDeepSeek 一站式生成并保存小说背景内容
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
//   - configJSON: DeepSeek配置的JSON字符串
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
//   - 需要在配置中或环境变量中提供有效的API密钥
//   - 可以通过传入参数控制生成内容的方向和类型
func GenerateAndSaveWithDeepSeek(ctx context.Context, c *app.RequestContext, configJSON string, theme string, ruleType string, character string) (*model.Worldview, *model.Rule, *model.BackgroundInfo, error) {
	// 解析配置
	config, err := NewDeepSeekConfig(configJSON)
	if err != nil {
		hlog.CtxErrorf(ctx, "解析DeepSeek配置失败: %v", err)
		return nil, nil, nil, err
	}
	
	// 生成世界观
	hlog.CtxInfof(ctx, "开始生成世界观...")
	worldview, err := GenerateWorldviewWithDeepSeek(ctx, config, theme)
	if err != nil {
		hlog.CtxErrorf(ctx, "生成世界观失败: %v", err)
		return nil, nil, nil, err
	}
	
	// 生成规则
	hlog.CtxInfof(ctx, "开始生成规则...")
	rule, err := GenerateRuleWithDeepSeek(ctx, config, worldview, ruleType)
	if err != nil {
		hlog.CtxErrorf(ctx, "生成规则失败: %v", err)
		return worldview, nil, nil, err
	}
	
	// 生成背景信息
	hlog.CtxInfof(ctx, "开始生成背景信息...")
	backgroundInfo, err := GenerateBackgroundInfoWithDeepSeek(ctx, config, worldview, rule, character)
	if err != nil {
		hlog.CtxErrorf(ctx, "生成背景信息失败: %v", err)
		return worldview, rule, nil, err
	}
	
	// 保存所有内容到数据库
	hlog.CtxInfof(ctx, "开始保存生成的内容到数据库...")
	err = SaveGeneratedContent(ctx, c, worldview, rule, backgroundInfo)
	if err != nil {
		hlog.CtxErrorf(ctx, "保存生成内容失败: %v", err)
		return worldview, rule, backgroundInfo, err
	}
	
	hlog.CtxInfof(ctx, "成功生成并保存内容，世界观ID: %d, 规则ID: %d, 背景信息ID: %d", 
		worldview.Id, rule.Id, backgroundInfo.Id)
	return worldview, rule, backgroundInfo, nil
}

// 注意：使用共享的cleanJsonResponse函数，该函数定义在ollama_generate.go中
