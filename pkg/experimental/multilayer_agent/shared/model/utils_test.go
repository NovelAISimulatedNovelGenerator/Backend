package model

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// 用于测试的简单结构体
type testStruct struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	IsValid bool   `json:"is_valid"`
}

// 扩展mockLLMModel以支持utils测试需求
type utilsTestModel struct {
	*ModelWrapper
	mockResponse     string
	mockError        error
	supportsJSON     bool
	supportsStreaming bool
}

func (m *utilsTestModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return m.mockResponse, m.mockError
}

func (m *utilsTestModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return nil, errors.New("未实现")
}

func (m *utilsTestModel) SupportsJSON() bool {
	return m.supportsJSON
}

func (m *utilsTestModel) SupportsStreaming() bool {
	return m.supportsStreaming
}

// TestGetStructuredOutput 测试从模型响应中提取JSON结构化数据
func TestGetStructuredOutput(t *testing.T) {
	// 测试有效的JSON响应
	t.Run("有效的JSON响应应被正确解析", func(t *testing.T) {
		validJSON := `{"name": "测试用户", "age": 30, "is_valid": true}`
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockResponse: validJSON,
			supportsJSON: true,
		}
		
		var result testStruct
		err := GetStructuredOutput(context.Background(), model, "测试提示词", &result)
		
		assert.NoError(t, err)
		assert.Equal(t, "测试用户", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.IsValid)
	})
	
	// 测试带有额外文本的JSON响应
	t.Run("带有额外文本的JSON响应应被清理并正确解析", func(t *testing.T) {
		mixedResponse := `以下是您请求的JSON数据：
		
		{"name": "测试用户", "age": 30, "is_valid": true}
		
		希望这对您有所帮助！`
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockResponse: mixedResponse,
			supportsJSON: true,
		}
		
		var result testStruct
		err := GetStructuredOutput(context.Background(), model, "测试提示词", &result)
		
		assert.NoError(t, err)
		assert.Equal(t, "测试用户", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.True(t, result.IsValid)
	})
	
	// 测试模型调用错误
	t.Run("模型调用错误应被正确传播", func(t *testing.T) {
		expectedErr := errors.New("模型调用失败")
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockError: expectedErr,
			supportsJSON: true,
		}
		
		var result testStruct
		err := GetStructuredOutput(context.Background(), model, "测试提示词", &result)
		
		assert.Error(t, err)
		assert.ErrorContains(t, err, "调用模型获取结构化输出失败")
	})
	
	// 测试无效的JSON响应
	t.Run("无效的JSON响应应返回解析错误", func(t *testing.T) {
		invalidJSON := `{"name": "测试用户", "age": "三十", "is_valid": true}`
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockResponse: invalidJSON,
			supportsJSON: true,
		}
		
		var result testStruct
		err := GetStructuredOutput(context.Background(), model, "测试提示词", &result)
		
		assert.Error(t, err)
		assert.ErrorContains(t, err, "解析JSON响应失败")
	})
}

// TestGenerateWithTemplate 测试使用模板生成内容
func TestGenerateWithTemplate(t *testing.T) {
	// 测试成功的模板替换
	t.Run("模板参数应被正确替换", func(t *testing.T) {
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockResponse: "这是一个关于测试主题的内容，针对测试用户群体。",
		}
		
		template := "请生成一个关于{{.Topic}}的内容，针对{{.Audience}}。"
		params := map[string]string{
			"Topic": "测试主题",
			"Audience": "测试用户群体",
		}
		
		result, err := GenerateWithTemplate(context.Background(), model, template, params)
		
		assert.NoError(t, err)
		assert.Equal(t, "这是一个关于测试主题的内容，针对测试用户群体。", result)
	})
	
	// 测试模型调用错误
	t.Run("模型调用错误应被正确传播", func(t *testing.T) {
		expectedErr := errors.New("模型调用失败")
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockError: expectedErr,
		}
		
		template := "测试模板{{.Param}}"
		params := map[string]string{"Param": "值"}
		
		_, err := GenerateWithTemplate(context.Background(), model, template, params)
		
		assert.Error(t, err)
		assert.ErrorContains(t, err, "使用模板生成内容失败")
	})
}

// TestStreamContent 测试流式获取内容
func TestStreamContent(t *testing.T) {
	// 测试不支持流式输出的模型
	t.Run("不支持流式输出的模型应返回错误", func(t *testing.T) {
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			supportsStreaming: false,
		}
		
		err := StreamContent(context.Background(), model, "测试提示词", func(chunk string) error {
			return nil
		})
		
		assert.Error(t, err)
		assert.ErrorContains(t, err, "不支持流式输出")
	})
	
	// 测试支持流式输出但调用失败的模型
	t.Run("模型调用错误应被正确传播", func(t *testing.T) {
		expectedErr := errors.New("模型调用失败")
		
		model := &utilsTestModel{
			ModelWrapper: &ModelWrapper{
				Type: ModelTypeOllama,
				Name: "test-model",
			},
			mockError: expectedErr,
			supportsStreaming: true,
		}
		
		err := StreamContent(context.Background(), model, "测试提示词", func(chunk string) error {
			return nil
		})
		
		assert.Error(t, err)
		assert.ErrorContains(t, err, "流式生成内容失败")
	})
}

// TestCleanJSONResponse 测试清理模型响应中的非JSON内容
func TestCleanJSONResponse(t *testing.T) {
	// 测试完全有效的JSON
	t.Run("完全有效的JSON应保持不变", func(t *testing.T) {
		validJSON := `{"name": "测试", "value": 123}`
		result := cleanJSONResponse(validJSON)
		assert.Equal(t, validJSON, result)
		
		// 验证结果是有效的JSON
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(result), &obj)
		assert.NoError(t, err)
	})
	
	// 测试带有前缀文本的JSON
	t.Run("带有前缀文本的JSON应被正确提取", func(t *testing.T) {
		mixedJSON := `以下是您请求的JSON:
		{"name": "测试", "value": 123}`
		result := cleanJSONResponse(mixedJSON)
		
		// 验证结果是有效的JSON
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(result), &obj)
		assert.NoError(t, err)
		assert.Equal(t, "测试", obj["name"])
		assert.Equal(t, float64(123), obj["value"])
	})
	
	// 测试带有后缀文本的JSON
	t.Run("带有后缀文本的JSON应被正确提取", func(t *testing.T) {
		mixedJSON := `{"name": "测试", "value": 123}
		
		希望这对您有所帮助！`
		result := cleanJSONResponse(mixedJSON)
		
		// 验证结果是有效的JSON
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(result), &obj)
		assert.NoError(t, err)
		assert.Equal(t, "测试", obj["name"])
		assert.Equal(t, float64(123), obj["value"])
	})
	
	// 测试带有前后缀文本的JSON
	t.Run("带有前后缀文本的JSON应被正确提取", func(t *testing.T) {
		mixedJSON := `以下是您请求的JSON:
		
		{"name": "测试", "value": 123}
		
		希望这对您有所帮助！`
		result := cleanJSONResponse(mixedJSON)
		
		// 验证结果是有效的JSON
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(result), &obj)
		assert.NoError(t, err)
		assert.Equal(t, "测试", obj["name"])
		assert.Equal(t, float64(123), obj["value"])
	})
	
	// 测试无JSON内容的情况
	t.Run("不包含JSON的内容应原样返回", func(t *testing.T) {
		noJSON := "这是一段不包含任何JSON的文本内容。"
		result := cleanJSONResponse(noJSON)
		assert.Equal(t, noJSON, result)
	})
}

// TestDescribeSampleStructure 测试生成输出类型的示例结构描述
func TestDescribeSampleStructure(t *testing.T) {
	// 测试结构体类型
	t.Run("结构体类型应生成有效的JSON示例", func(t *testing.T) {
		sample := testStruct{}
		result := describeSampleStructure(&sample)
		
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "age")
		assert.Contains(t, result, "is_valid")
		
		// 验证结果是有效的JSON
		var obj testStruct
		err := json.Unmarshal([]byte(result), &obj)
		assert.NoError(t, err)
	})
	
	// 测试无法序列化的类型
	t.Run("无法序列化的类型应返回默认提示", func(t *testing.T) {
		// 创建一个循环引用，导致无法正确序列化
		type circular struct {
			Self *circular
		}
		c := &circular{}
		c.Self = c
		
		result := describeSampleStructure(c)
		assert.Contains(t, result, "JSON")
	})
}
