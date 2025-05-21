// Package model 提供多层代理系统的模型接口层实现
package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/llms"
)

// GetStructuredOutput 从模型响应中提取JSON结构化数据
// 支持所有实现Model接口的模型，如Ollama和DeepSeek
func GetStructuredOutput(ctx context.Context, model Model, prompt string, outputType interface{}) error {
	// 构建提示词，明确要求JSON输出
	formattedPrompt := fmt.Sprintf(`%s

请以有效的JSON格式回复，不要有任何额外文本。JSON结构应该是：
%s`, prompt, describeSampleStructure(outputType))

	// 调用模型并要求JSON输出
	response, err := model.Call(ctx, formattedPrompt, llms.WithJSONMode())
	if err != nil {
		return fmt.Errorf("调用模型获取结构化输出失败: %w", err)
	}

	// 清理响应中的非JSON内容
	cleanResponse := cleanJSONResponse(response)
	
	// 解析JSON响应到目标结构
	err = json.Unmarshal([]byte(cleanResponse), outputType)
	if err != nil {
		hlog.Errorf("解析JSON响应失败: %v, 原始响应: %s, 清理后: %s", err, response, cleanResponse)
		return fmt.Errorf("解析JSON响应失败: %w", err)
	}

	return nil
}

// GenerateWithTemplate 使用模板生成内容
// 模板中的{{.Param}}占位符将被替换为params中相应的值
func GenerateWithTemplate(ctx context.Context, model Model, template string, params map[string]string) (string, error) {
	// 替换模板中的参数
	prompt := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		prompt = strings.Replace(prompt, placeholder, value, -1)
	}

	// 调用模型
	response, err := model.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("使用模板生成内容失败: %w", err)
	}

	return response, nil
}

// StreamContent 从模型中流式获取内容
// 每当接收到新内容片段时调用chunkHandler函数
func StreamContent(ctx context.Context, model Model, prompt string, chunkHandler func(chunk string) error) error {
	if !model.SupportsStreaming() {
		return fmt.Errorf("模型 %s (%s) 不支持流式输出", model.ModelName(), model.ModelType())
	}

	// 定义用于接收流式内容的函数
	streamingFunc := func(ctx context.Context, chunk []byte) error {
		return chunkHandler(string(chunk))
	}

	// 调用模型
	_, err := model.Call(ctx, prompt, llms.WithStreamingFunc(streamingFunc))
	if err != nil {
		return fmt.Errorf("流式生成内容失败: %w", err)
	}

	return nil
}

// describeSampleStructure 返回一个输出类型的示例结构描述
func describeSampleStructure(outputType interface{}) string {
	// 使用反射或类型断言获取结构信息
	// 这里简化处理，只返回JSON格式的空结构
	emptyInstance, err := json.MarshalIndent(outputType, "", "  ")
	if err != nil {
		// 如果序列化失败，返回简单提示
		return "JSON对象"
	}
	return string(emptyInstance)
}

// cleanJSONResponse 清理模型响应中的非JSON内容
func cleanJSONResponse(response string) string {
	// 寻找第一个"{"或"["，表示JSON开始
	startPos := strings.IndexAny(response, "{[")
	if startPos == -1 {
		return response // 没有找到JSON开始标记
	}

	// 找到最后一个"}"或"]"，表示JSON结束
	endPos := strings.LastIndexAny(response, "}]")
	if endPos == -1 || endPos < startPos {
		return response // 没有找到匹配的JSON结束标记
	}

	// 提取JSON部分
	jsonPart := response[startPos : endPos+1]

	// 尝试验证JSON是否有效
	var dummy interface{}
	if json.Unmarshal([]byte(jsonPart), &dummy) == nil {
		return jsonPart
	}

	// 如果简单提取失败，尝试更复杂的提取方法
	// 例如处理嵌套的JSON或多余的转义字符
	// 这里只是一个简化示例
	return response
}
