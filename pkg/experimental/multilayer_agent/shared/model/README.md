# 多层代理模型接口设计

## 概述

本模块实现了多层代理系统中的模型接口层，基于LangChain Go库，为上层代理提供统一、灵活的大语言模型访问能力。模型接口层负责处理与底层LLM的通信，包括请求构建、响应解析和错误处理，使代理系统能够专注于业务逻辑而不必关心模型交互细节。

## 核心组件

### 1. 模型接口 (`interface.go`)

定义了统一的模型接口，支持两种主要的交互模式：

- **标准文本补全**：输入提示词，输出文本回应
- **内容生成**：支持多模态输入（文本、图像等），返回结构化内容

接口特点：
- 基于LangChain Go的`llms.Model`接口扩展
- 支持同步和流式交互方式
- 提供统一的参数设置API

### 2. Ollama模型实现 (`ollama.go`)

基于Ollama API的模型实现，适用于本地部署的开源模型：

- 支持多种开源模型，如Llama、Mistral等
- 支持高级功能，如JSON输出格式化
- 优化本地调用性能

### 3. DeepSeek API模型实现 (`deepseek-api.go`)

基于DeepSeek API的模型实现，提供云端高性能模型服务：

- 支持DeepSeek系列模型
- 提供企业级稳定性和扩展性
- 针对复杂推理任务优化

## 使用方法

### 基本调用流程

```go
// 1. 初始化模型
model, err := ollama.New(ollama.WithModel("llama2"))
if err != nil {
    // 处理错误
}

// 2. 调用模型（简单文本补全）
response, err := model.Call(ctx, "讲一个关于AI的故事")
if err != nil {
    // 处理错误
}

// 3. 使用多模态内容生成
messages := []llms.MessageContent{
    llms.TextParts(llms.ChatMessageTypeHuman, "描述这张图片"),
    // 可以添加图像、系统提示等
}
contentResponse, err := model.GenerateContent(ctx, messages)
if err != nil {
    // 处理错误
}
```

### 高级参数配置

```go
// 使用CallOption设置生成参数
response, err := model.Call(ctx, prompt,
    llms.WithTemperature(0.7),
    llms.WithMaxTokens(2048),
    llms.WithStopWords([]string{"END"}),
)

// JSON格式化输出
response, err := model.Call(ctx, prompt, 
    llms.WithJSONMode(),
)
```

### 流式输出处理

```go
// 定义流式处理函数
streamFunc := func(ctx context.Context, chunk []byte) error {
    fmt.Print(string(chunk))
    return nil
}

// 使用流式输出
response, err := model.Call(ctx, prompt,
    llms.WithStreamingFunc(streamFunc),
)
```

## 自定义模型扩展

如需扩展支持其他模型，只需实现`Model`接口：

```go
// 1. 定义新的模型结构
type CustomModel struct {
    // 模型配置和状态
}

// 2. 实现Call方法
func (m *CustomModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
    // 实现模型调用逻辑
}

// 3. 实现GenerateContent方法
func (m *CustomModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
    // 实现内容生成逻辑
}
```

## 系统集成

模型接口在多层代理系统中的位置和交互：

- **模型接口作为共享资源**：所有代理都可通过统一接口访问不同底层模型
- **模型隔离**：不同代理可以使用不同的模型实例，避免交叉影响
- **灵活配置**：支持运行时切换和配置模型参数
