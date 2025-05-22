// Package example_tool 提供了一个langchaingo Tool接口的完整示例实现
package example_tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/tools"
)

// TestExampleTool 测试示例工具的基本功能
func TestExampleTool(t *testing.T) {
	// 创建上下文
	ctx := context.Background()

	// 创建工具实例
	tool := NewExampleTool(nil)

	// 测试工具名称和描述
	if tool.Name() != "example_tool" {
		t.Errorf("期望工具名称为 'example_tool'，得到 '%s'", tool.Name())
	}

	if len(tool.Description()) == 0 {
		t.Error("工具描述不应为空")
	}

	// 测试工具调用 - 默认参数
	result, err := tool.Call(ctx, "")
	if err != nil {
		t.Errorf("工具调用失败: %v", err)
	}

	// 确保结果包含默认参数的处理信息
	if result == "" {
		t.Error("工具返回结果不应为空")
	}
	hlog.Infof("默认参数调用结果: %s", result)

	// 测试工具调用 - 自定义参数
	params := ExampleToolParams{
		Text:    "测试文本",
		Number:  100,
		Flag:    true,
		Options: []string{"选项A", "选项B", "选项C"},
		NestedData: &NestedParams{
			Key:   "测试键",
			Value: "测试值",
		},
	}

	// 序列化参数为JSON
	inputJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("参数序列化失败: %v", err)
	}

	// 使用自定义参数调用工具
	result, err = tool.Call(ctx, string(inputJSON))
	if err != nil {
		t.Errorf("自定义参数工具调用失败: %v", err)
	}

	hlog.Infof("自定义参数调用结果: %s", result)
}

// TestExampleToolWithOptions 测试使用选项模式创建工具
func TestExampleToolWithOptions(t *testing.T) {
	// 创建上下文
	ctx := context.Background()

	// 使用选项模式创建工具
	customTool := NewExampleToolWithOptions(
		WithCustomName("custom_example_tool"),
		WithCustomDescription("这是一个自定义名称和描述的示例工具"),
	)

	// 测试自定义工具名称和描述
	if customTool.Name() != "custom_example_tool" {
		t.Errorf("期望工具名称为 'custom_example_tool'，得到 '%s'", customTool.Name())
	}

	if customTool.Description() != "这是一个自定义名称和描述的示例工具\n这个工具接受以下JSON格式的输入参数:\n{\n  \"text\": \"字符串参数\",\n  \"number\": 数值参数,\n  \"flag\": 布尔参数,\n  \"options\": [\"选项1\", \"选项2\", ...],\n  \"nested_data\": {\n    \"key\": \"嵌套键\",\n    \"value\": 嵌套值\n  }\n}" {
		t.Errorf("自定义描述不匹配")
	}

	// 测试工具调用
	result, err := customTool.Call(ctx, "")
	if err != nil {
		t.Errorf("自定义工具调用失败: %v", err)
	}

	hlog.Infof("自定义工具调用结果: %s", result)
}

// ExampleNewExampleTool 展示如何创建和使用示例工具
func ExampleNewExampleTool() {
	// 创建工具实例
	tool := NewExampleTool(nil)

	// 获取工具名称和描述
	name := tool.Name()
	description := tool.Description()

	// 输出工具信息
	hlog.Infof("工具名称: %s", name)
	hlog.Infof("工具描述: %s", description)

	// 准备输入参数
	input := `{
		"text": "示例文本",
		"number": 123,
		"flag": true,
		"options": ["A", "B", "C"],
		"nested_data": {
			"key": "test",
			"value": 42
		}
	}`

	// 调用工具
	result, err := tool.Call(context.Background(), input)
	if err != nil {
		hlog.Errorf("工具调用失败: %v", err)
		return
	}

	// 输出结果
	hlog.Infof("工具调用结果: %s", result)
}

// Example 展示如何将示例工具注册到工具注册表
func Example() {
	// 假设我们有一个工具注册表实例 registry
	// registry := tools.NewToolRegistry()

	// 创建自定义工具配置
	config := &ExampleToolConfig{
		Verbose:    true,
		MaxRetries: 5,
		DefaultParams: ExampleToolParams{
			Text:   "自定义默认文本",
			Number: 100,
			Flag:   true,
			Options: []string{"自定义选项1", "自定义选项2"},
		},
	}

	// 创建工具实例
	exampleTool := NewExampleTool(config)

	// 注册工具到注册表
	// err := registry.RegisterTool(exampleTool)
	// if err != nil {
	//     hlog.Errorf("工具注册失败: %v", err)
	//     return
	// }

	// 获取工具列表
	// toolList := registry.ListTools()
	// for _, t := range toolList {
	//     hlog.Infof("已注册工具: %s", t.Name())
	// }

	// 根据名称获取工具
	// tool, err := registry.GetTool("example_tool")
	// if err != nil {
	//     hlog.Errorf("获取工具失败: %v", err)
	//     return
	// }

	// 使用获取的工具
	// result, err := tool.Call(context.Background(), "{}")
	// if err != nil {
	//     hlog.Errorf("工具调用失败: %v", err)
	//     return
	// }
	// hlog.Infof("工具调用结果: %s", result)

	// 为了让这个示例编译通过，我们至少使用一次 exampleTool
	_ = exampleTool
}

// Example_customConfig 展示如何创建一个仅修改特定配置项的工具
func Example_customConfig() {
	// 创建一个自定义配置
	config := &ExampleToolConfig{
		// 只修改我们关心的配置项
		Verbose: true,
		// 其他配置保持默认值
	}

	// 创建工具实例
	tool := NewExampleTool(config)

	// 使用工具
	result, _ := tool.Call(context.Background(), "")
	hlog.Infof("结果: %s", result)
}

// 示例：如何实现一个新的工具类型
type CustomTool struct {
	// 实现 tools.Tool 接口所需的字段和方法
}

func (t *CustomTool) Name() string {
	return "custom_tool"
}

func (t *CustomTool) Description() string {
	return "这是一个自定义工具"
}

func (t *CustomTool) Call(ctx context.Context, input string) (string, error) {
	// 自定义工具逻辑
	return "自定义工具结果", nil
}

// 确保 CustomTool 实现了 tools.Tool 接口
var _ tools.Tool = (*CustomTool)(nil)
