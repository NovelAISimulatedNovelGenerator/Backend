# LangChainGo 工具示例实现

本文档详细介绍了基于 `github.com/tmc/langchaingo/tools` 包的示例工具实现，展示了创建自定义工具时所有可用的参数和配置选项。

## 1. 工具接口概述

LangChainGo 的 Tool 接口定义如下：

```go
type Tool interface {
    Name() string
    Description() string
    Call(ctx context.Context, input string) (string, error)
}
```

每个工具必须实现这三个方法：
- `Name()` - 返回工具的唯一名称
- `Description()` - 返回工具的详细描述，应包含使用说明
- `Call()` - 执行工具功能，接受上下文和输入字符串，返回结果字符串或错误

## 2. 示例工具结构

示例工具实现了以下核心组件：

### 2.1 参数结构

```go
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
```

### 2.2 工具配置

```go
// ExampleToolConfig 定义了工具的配置选项
type ExampleToolConfig struct {
    // Verbose 控制是否启用详细日志
    Verbose bool
    
    // MaxRetries 定义最大重试次数
    MaxRetries int
    
    // DefaultParams 定义默认参数值
    DefaultParams ExampleToolParams
}
```

### 2.3 工具结构

```go
// ExampleTool 实现 langchaingo/tools.Tool 接口
type ExampleTool struct {
    // 工具名称（必须）
    name string
    
    // 工具描述（必须）
    description string
    
    // 自定义配置选项
    config *ExampleToolConfig
}
```

## 3. 创建工具实例

有两种创建工具实例的方式：

### 3.1 使用配置对象

```go
// 创建默认配置的工具
tool := NewExampleTool(nil)

// 创建自定义配置的工具
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
customTool := NewExampleTool(config)
```

### 3.2 使用选项模式

```go
// 使用选项模式创建工具
optionsTool := NewExampleToolWithOptions(
    WithCustomName("custom_example_tool"),
    WithCustomDescription("这是一个自定义名称和描述的示例工具"),
)
```

## 4. 使用工具

### 4.1 基本调用

```go
// 创建上下文
ctx := context.Background()

// 创建工具实例
tool := NewExampleTool(nil)

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
result, err := tool.Call(ctx, input)
if err != nil {
    // 处理错误
    return
}

// 使用结果
fmt.Println(result)
```

### 4.2 在工具注册表中使用

```go
// 创建工具注册表
registry := tools.NewToolRegistry()

// 创建工具实例
exampleTool := NewExampleTool(nil)

// 注册工具到注册表
err := registry.RegisterTool(exampleTool)
if err != nil {
    // 处理错误
    return
}

// 获取工具
tool, err := registry.GetTool("example_tool")
if err != nil {
    // 处理错误
    return
}

// 使用工具
result, err := tool.Call(context.Background(), "{}")
```

## 5. 参数说明

示例工具接受以下JSON格式的输入参数：

```json
{
  "text": "字符串参数",
  "number": 数值参数,
  "flag": 布尔参数,
  "options": ["选项1", "选项2", ...],
  "nested_data": {
    "key": "嵌套键",
    "value": 嵌套值
  }
}
```

- **text**: 字符串参数，用于处理文本信息
- **number**: 整数参数，用于数值计算
- **flag**: 布尔参数，控制工具行为
- **options**: 字符串数组，提供多个选项
- **nested_data**: 嵌套对象，展示复杂参数处理
  - **key**: 嵌套对象的键
  - **value**: 嵌套对象的值，可以是任意类型

## 6. 自定义工具开发指南

### 6.1 实现 Tool 接口

创建自定义工具时，必须实现 `tools.Tool` 接口的三个方法：

```go
// CustomTool 示例
type CustomTool struct {
    // 自定义字段...
}

func (t *CustomTool) Name() string {
    return "custom_tool_name"
}

func (t *CustomTool) Description() string {
    return "详细描述工具的功能和使用方法"
}

func (t *CustomTool) Call(ctx context.Context, input string) (string, error) {
    // 1. 解析输入
    // 2. 执行核心逻辑
    // 3. 返回结果
    return "处理结果", nil
}
```

### 6.2 参数处理最佳实践

1. **始终验证输入**：检查必要参数是否存在和有效
2. **提供默认值**：为可选参数提供合理的默认值
3. **详细错误信息**：返回具体的错误原因，帮助调试
4. **使用JSON结构**：标准化输入和输出格式
5. **记录关键日志**：记录工具调用的重要信息

## 7. 工具集成指南

### 7.1 在多层代理系统中注册

```go
// 创建工具注册表
registry := tools.NewToolRegistry()

// 创建并注册工具
exampleTool := NewExampleTool(nil)
err := registry.RegisterTool(exampleTool)

// 创建适配器（如果需要）
adapter := tools.NewLangChainAdapter(exampleTool)
```

### 7.2 错误处理

工具应当处理以下类型的错误：

1. 输入解析错误
2. 参数验证错误
3. 执行过程错误
4. 资源访问错误

每种错误应当有明确的错误信息，便于调试和处理。

## 8. 测试指南

示例工具包含完整的测试案例，展示了如何测试工具的各项功能：

```go
// 测试工具基本功能
func TestExampleTool(t *testing.T) {...}

// 测试使用选项模式创建工具
func TestExampleToolWithOptions(t *testing.T) {...}

// 示例：创建和使用工具
func ExampleNewExampleTool() {...}

// 示例：注册工具到注册表
func ExampleRegisterExampleTool() {...}
```

测试应覆盖以下方面：
- 工具名称和描述
- 默认参数处理
- 自定义参数处理
- 错误情况处理
- 边界情况测试

## 9. 参考资料

- [LangChainGo 官方文档](https://pkg.go.dev/github.com/tmc/langchaingo)
- [Go 标准库 - context](https://golang.org/pkg/context/)
- [Go 标准库 - json](https://golang.org/pkg/encoding/json/)
