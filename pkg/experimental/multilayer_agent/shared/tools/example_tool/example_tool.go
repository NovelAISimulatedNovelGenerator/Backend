// Package example_tool 提供了一个langchaingo Tool接口的完整示例实现
// 展示了如何创建自定义工具以及所有可用的参数和配置选项
package example_tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/tools"
)

// ExampleToolParams 定义示例工具可接受的所有参数
type ExampleToolParams struct {
	// Text 是一个基本的字符串参数
	Text string `json:"text"`
	
	// Number 是一个数值参数
	Number int `json:"number"`
	
	// Flag 是一个布尔标志参数
	Flag bool `json:"flag"`
	
	// Options 是一个字符串列表参数
	Options []string `json:"options"`
	
	// NestedData 是一个嵌套的复杂数据参数
	NestedData *NestedParams `json:"nested_data,omitempty"`
}

// NestedParams 展示如何使用嵌套结构参数
type NestedParams struct {
	// Key 是嵌套参数中的键
	Key string `json:"key"`
	
	// Value 是嵌套参数中的值
	Value interface{} `json:"value"`
}

// ExampleTool 实现 langchaingo/tools.Tool 接口
// 用于展示工具开发的完整流程和参数使用方式
type ExampleTool struct {
	// 工具名称（必须）
	name string
	
	// 工具描述（必须）
	description string
	
	// 自定义配置选项
	config *ExampleToolConfig
}

// ExampleToolConfig 定义了工具的配置选项
type ExampleToolConfig struct {
	// Verbose 控制是否启用详细日志
	Verbose bool
	
	// MaxRetries 定义最大重试次数
	MaxRetries int
	
	// DefaultParams 定义默认参数值
	DefaultParams ExampleToolParams
}

// NewExampleTool 创建一个新的示例工具实例
// 参数:
//   - config: 可选的配置参数，传入nil将使用默认配置
//
// 返回:
//   - tools.Tool: 实现了Tool接口的示例工具
func NewExampleTool(config *ExampleToolConfig) tools.Tool {
	// 如果未提供配置，使用默认配置
	if config == nil {
		config = &ExampleToolConfig{
			Verbose:    false,
			MaxRetries: 3,
			DefaultParams: ExampleToolParams{
				Text:   "默认文本",
				Number: 42,
				Flag:   false,
				Options: []string{"选项1", "选项2"},
			},
		}
	}
	
	return &ExampleTool{
		name:        "example_tool",
		description: "这是一个示例工具，展示了langchaingo工具接口的所有可用参数和用法。",
		config:      config,
	}
}

// Name 返回工具名称
// 实现 tools.Tool 接口必需的方法
func (t *ExampleTool) Name() string {
	return t.name
}

// Description 返回工具描述
// 实现 tools.Tool 接口必需的方法
// 此描述应当详细说明工具的功能、输入格式和预期输出
func (t *ExampleTool) Description() string {
	return fmt.Sprintf(`%s
这个工具接受以下JSON格式的输入参数:
{
  "text": "字符串参数",
  "number": 数值参数,
  "flag": 布尔参数,
  "options": ["选项1", "选项2", ...],
  "nested_data": {
    "key": "嵌套键",
    "value": 嵌套值
  }
}`, t.description)
}

// Call 执行工具功能
// 参数:
//   - ctx: 上下文，包含调用相关信息
//   - input: JSON格式的输入字符串
//
// 返回:
//   - string: 工具执行结果
//   - error: 执行过程中的错误，如果有
//
// 实现 tools.Tool 接口必需的方法
func (t *ExampleTool) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用开始
	if t.config.Verbose {
		hlog.CtxInfof(ctx, "示例工具调用开始，输入: %s", input)
	}
	
	// 解析输入参数
	params, err := t.parseInput(ctx, input)
	if err != nil {
		return "", fmt.Errorf("解析输入失败: %w", err)
	}
	
	// 执行工具逻辑
	result, err := t.execute(ctx, params)
	if err != nil {
		if t.config.Verbose {
			hlog.CtxErrorf(ctx, "示例工具执行失败: %v", err)
		}
		return "", err
	}
	
	// 记录工具调用结束
	if t.config.Verbose {
		hlog.CtxInfof(ctx, "示例工具执行成功")
	}
	
	return result, nil
}

// parseInput 解析输入字符串为参数结构
func (t *ExampleTool) parseInput(ctx context.Context, input string) (*ExampleToolParams, error) {
	// 移除输入中可能存在的多余空白
	input = strings.TrimSpace(input)
	
	// 如果输入为空，使用默认参数
	if input == "" {
		return &t.config.DefaultParams, nil
	}
	
	// 尝试解析JSON输入
	var params ExampleToolParams
	err := json.Unmarshal([]byte(input), &params)
	if err != nil {
		return nil, fmt.Errorf("JSON解析错误: %w", err)
	}
	
	// 应用默认值到未指定的字段
	t.applyDefaults(&params)
	
	return &params, nil
}

// applyDefaults 对未设置的参数应用默认值
func (t *ExampleTool) applyDefaults(params *ExampleToolParams) {
	defaults := t.config.DefaultParams
	
	// 如果未设置文本，使用默认值
	if params.Text == "" {
		params.Text = defaults.Text
	}
	
	// 如果未设置选项，使用默认选项
	if len(params.Options) == 0 {
		params.Options = defaults.Options
	}
}

// execute 执行工具主要逻辑
func (t *ExampleTool) execute(ctx context.Context, params *ExampleToolParams) (string, error) {
	// 构建结果
	var result strings.Builder
	
	// 添加参数处理结果
	result.WriteString(fmt.Sprintf("处理文本: %s\n", params.Text))
	result.WriteString(fmt.Sprintf("数值 x 2: %d\n", params.Number*2))
	
	if params.Flag {
		result.WriteString("标志已启用\n")
	} else {
		result.WriteString("标志未启用\n")
	}
	
	result.WriteString("选项列表:\n")
	for i, opt := range params.Options {
		result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, opt))
	}
	
	if params.NestedData != nil {
		result.WriteString(fmt.Sprintf("嵌套数据: 键=%s, 值=%v\n", 
			params.NestedData.Key, params.NestedData.Value))
	}
	
	return result.String(), nil
}

// WithCustomName 设置自定义工具名称
func WithCustomName(name string) func(*ExampleTool) {
	return func(t *ExampleTool) {
		t.name = name
	}
}

// WithCustomDescription 设置自定义工具描述
func WithCustomDescription(description string) func(*ExampleTool) {
	return func(t *ExampleTool) {
		t.description = description
	}
}

// NewExampleToolWithOptions 使用选项模式创建工具实例
// 参数:
//   - opts: 可变参数函数列表，用于自定义工具配置
//
// 返回:
//   - tools.Tool: 根据选项配置的工具实例
func NewExampleToolWithOptions(opts ...func(*ExampleTool)) tools.Tool {
	// 创建默认工具实例
	tool := &ExampleTool{
		name:        "example_tool",
		description: "这是一个示例工具，展示了langchaingo工具接口的所有可用参数和用法。",
		config: &ExampleToolConfig{
			Verbose:    false,
			MaxRetries: 3,
			DefaultParams: ExampleToolParams{
				Text:   "默认文本",
				Number: 42,
				Flag:   false,
				Options: []string{"选项1", "选项2"},
			},
		},
	}
	
	// 应用所有选项
	for _, opt := range opts {
		opt(tool)
	}
	
	return tool
}
