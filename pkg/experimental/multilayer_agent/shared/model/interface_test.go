package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// 模拟llms.Model的实现，用于测试ModelWrapper
type mockLLMModel struct {
	callResponse           string
	callError              error
	generateContentResp    *llms.ContentResponse
	generateContentError   error
	callCount              int
	generateContentCount   int
}

func (m *mockLLMModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	m.callCount++
	return m.callResponse, m.callError
}

func (m *mockLLMModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	m.generateContentCount++
	return m.generateContentResp, m.generateContentError
}

// TestModelWrapper 测试ModelWrapper的基本功能
func TestModelWrapper(t *testing.T) {
	// 创建一个模拟的LLM模型
	mockModel := &mockLLMModel{
		callResponse: "测试响应",
		generateContentResp: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: "测试内容",
				},
			},
		},
	}

	// 创建ModelWrapper实例
	wrapper := &ModelWrapper{
		BaseModel:        mockModel,
		Type:             ModelTypeOllama,
		Name:             "test-model",
		TokenLimit:       4096,
		JSONSupport:      true,
		StreamingSupport: true,
		VisionSupport:    false,
	}

	// 测试ModelType方法
	t.Run("ModelType方法应返回正确的模型类型", func(t *testing.T) {
		assert.Equal(t, ModelTypeOllama, wrapper.ModelType())
	})

	// 测试ModelName方法
	t.Run("ModelName方法应返回正确的模型名称", func(t *testing.T) {
		assert.Equal(t, "test-model", wrapper.ModelName())
	})

	// 测试GetTokenLimit方法
	t.Run("GetTokenLimit方法应返回正确的Token限制", func(t *testing.T) {
		assert.Equal(t, 4096, wrapper.GetTokenLimit())
	})

	// 测试EstimateTokens方法
	t.Run("EstimateTokens方法应能粗略估算文本的token数量", func(t *testing.T) {
		tokens, err := wrapper.EstimateTokens("这是一个测试文本，用于测试EstimateTokens方法")
		assert.NoError(t, err)
		assert.Greater(t, tokens, 0)
	})

	// 测试SupportsJSON方法
	t.Run("SupportsJSON方法应返回正确的JSON支持状态", func(t *testing.T) {
		assert.True(t, wrapper.SupportsJSON())
	})

	// 测试SupportsStreaming方法
	t.Run("SupportsStreaming方法应返回正确的流式输出支持状态", func(t *testing.T) {
		assert.True(t, wrapper.SupportsStreaming())
	})

	// 测试SupportsVision方法
	t.Run("SupportsVision方法应返回正确的视觉输入支持状态", func(t *testing.T) {
		assert.False(t, wrapper.SupportsVision())
	})

	// 测试Call方法
	t.Run("Call方法应正确代理到基础模型", func(t *testing.T) {
		ctx := context.Background()
		response, err := wrapper.Call(ctx, "测试提示词")
		
		assert.NoError(t, err)
		assert.Equal(t, "测试响应", response)
		assert.Equal(t, 1, mockModel.callCount)
	})

	// 测试GenerateContent方法
	t.Run("GenerateContent方法应正确代理到基础模型", func(t *testing.T) {
		ctx := context.Background()
		messages := []llms.MessageContent{
			{
				Role:    "user",
				Parts:   []llms.ContentPart{llms.TextPart("测试消息")},
			},
		}
		
		response, err := wrapper.GenerateContent(ctx, messages)
		
		assert.NoError(t, err)
		assert.Equal(t, "测试内容", response.Choices[0].Content)
		assert.Equal(t, 1, mockModel.generateContentCount)
	})
}

// TestDefaultModelFactory 测试DefaultModelFactory的功能
func TestDefaultModelFactory(t *testing.T) {
	// 创建工厂实例
	factory := NewModelFactory()

	// 测试创建未知类型的模型
	t.Run("创建未知类型的模型应返回错误", func(t *testing.T) {
		_, err := factory.CreateModel("unknown", ModelOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "未知的模型类型")
	})

	// 测试创建未实现的模型类型
	t.Run("创建未实现的模型类型应返回错误", func(t *testing.T) {
		_, err := factory.CreateModel(ModelTypeOpenAI, ModelOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "尚未实现")
	})

	// 注意：由于Ollama和DeepSeek模型创建依赖外部服务，
	// 这里不进行实际创建测试，应该在集成测试中进行
}
