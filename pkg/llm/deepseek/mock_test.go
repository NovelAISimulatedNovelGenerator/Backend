// Package deepseek 提供了与constants.DeepSeekChat API交互的功能，基于OpenAI官方SDK
package deepseek


import (
	"bytes"
	"io"
	"testing"
)

// 测试用的模拟流读取器
type mockReadCloser struct {
	*bytes.Reader
	closeFunc func() error
}

// Close 实现io.ReadCloser接口
func (m *mockReadCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// 创建模拟读取器
func newMockReadCloser(data string, closeFunc func() error) io.ReadCloser {
	return &mockReadCloser{
		Reader:    bytes.NewReader([]byte(data)),
		closeFunc: closeFunc,
	}
}

// TestStreamReader_Recv 测试流读取器的接收功能
func TestStreamReader_Recv(t *testing.T) {
	// 创建模拟数据
	mockSSE := `
data: {"id":"cmpl-123","choices":[{"text":"这是"}]}

data: {"id":"cmpl-123","choices":[{"text":"一个"}]}

data: {"id":"cmpl-123","choices":[{"text":"测试"}]}

data: [DONE]
`
	
	// 创建模拟读取器
	mockBody := newMockReadCloser(mockSSE, nil)
	
	// 创建流读取器
	streamReader := NewStreamReader(mockBody)
	
	// 读取数据并验证
	expectedTexts := []string{
		"这是",
		"一个",
		"测试",
	}
	
	for i, expectedText := range expectedTexts {
		resp, err := streamReader.Recv()
		if err != nil {
			t.Fatalf("第%d次读取失败: %v", i+1, err)
		}
		
		// 验证响应中包含choices数组
		choices, ok := resp["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			t.Fatalf("第%d次读取响应中没有choices字段或为空", i+1)
		}
		
		// 获取第一个选择项
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			t.Fatalf("第%d次读取响应中的choices[0]不是一个有效的对象", i+1)
		}
		
		// 获取文本内容
		text, ok := choice["text"].(string)
		if !ok {
			t.Fatalf("第%d次读取响应中的choices[0].text不是一个有效的字符串", i+1)
		}
		
		// 验证文本内容
		if text != expectedText {
			t.Errorf("第%d次读取响应不匹配，期望'%s'，实际为'%s'", i+1, expectedText, text)
		}
	}
	
	// 验证读取完毕后是否返回EOF
	_, err := streamReader.Recv()
	if err != io.EOF {
		t.Errorf("期望EOF错误，实际为%v", err)
	}
}

// TestStreamReader_Close 测试流读取器的关闭功能
func TestStreamReader_Close(t *testing.T) {
	// 创建一个标志变量，用于检查Close是否被调用
	closeCalled := false
	
	// 创建模拟读取器，带有自定义Close函数
	mockBody := newMockReadCloser("", func() error {
		closeCalled = true
		return nil
	})
	
	// 创建流读取器
	streamReader := NewStreamReader(mockBody)
	
	// 调用Close方法
	err := streamReader.Close()
	if err != nil {
		t.Errorf("关闭流读取器失败: %v", err)
	}
	
	// 验证Close是否被调用
	if !closeCalled {
		t.Errorf("Close方法没有调用底层读取器的Close方法")
	}
	
	// 验证isFinished标志是否被设置
	if !streamReader.isFinished {
		t.Errorf("Close方法没有将isFinished设置为true")
	}
}

// MockAdapter 是一个模拟的适配器，用于测试
type MockAdapter struct {
	client *Client
}

// 模拟的readCompletionStream方法
func (a *MockAdapter) readCompletionStream(reader *StreamReader) (string, error) {
	var result string
	
	for {
		resp, err := reader.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		
		choices, ok := resp["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}
		
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}
		
		text, ok := choice["text"].(string)
		if !ok {
			continue
		}
		
		result += text
	}
	
	return result, nil
}

// TestReadCompletionStream 测试读取完成流的功能
func TestReadCompletionStream(t *testing.T) {
	// 创建一个模拟适配器实例
	adapter := &MockAdapter{
		client: &Client{
			config: DefaultConfig("test-api-key"),
		},
	}
	
	// 创建模拟数据
	mockSSE := `
data: {"id":"cmpl-123","choices":[{"text":"这是"}]}

data: {"id":"cmpl-123","choices":[{"text":"一个"}]}

data: {"id":"cmpl-123","choices":[{"text":"测试"}]}

data: [DONE]
`
	
	// 创建模拟读取器
	mockBody := newMockReadCloser(mockSSE, nil)
	streamReader := NewStreamReader(mockBody)
	
	// 调用读取流方法
	result, err := adapter.readCompletionStream(streamReader)
	if err != nil {
		t.Fatalf("读取完成流失败: %v", err)
	}
	
	// 验证结果
	expectedText := "这是一个测试"
	if result != expectedText {
		t.Errorf("期望结果为'%s'，实际为'%s'", expectedText, result)
	}
}

// 模拟的readChatCompletionStream方法
func (a *MockAdapter) readChatCompletionStream(reader *StreamReader) (string, error) {
	var result string
	
	for {
		resp, err := reader.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		
		choices, ok := resp["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}
		
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}
		
		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}
		
		content, ok := delta["content"].(string)
		if !ok {
			continue
		}
		
		result += content
	}
	
	return result, nil
}

// TestReadChatCompletionStream 测试读取聊天完成流的功能
func TestReadChatCompletionStream(t *testing.T) {
	// 创建一个模拟适配器实例
	adapter := &MockAdapter{
		client: &Client{
			config: DefaultConfig("test-api-key"),
		},
	}
	
	// 创建模拟数据
	mockSSE := `
data: {"id":"cmpl-123","choices":[{"delta":{"content":"这是"}}]}

data: {"id":"cmpl-123","choices":[{"delta":{"content":"一个"}}]}

data: {"id":"cmpl-123","choices":[{"delta":{"content":"聊天"}}]}

data: {"id":"cmpl-123","choices":[{"delta":{"content":"测试"}}]}

data: [DONE]
`
	
	// 创建模拟读取器
	mockBody := newMockReadCloser(mockSSE, nil)
	streamReader := NewStreamReader(mockBody)
	
	// 调用读取流方法
	result, err := adapter.readChatCompletionStream(streamReader)
	if err != nil {
		t.Fatalf("读取聊天完成流失败: %v", err)
	}
	
	// 验证结果
	expectedText := "这是一个聊天测试"
	if result != expectedText {
		t.Errorf("期望结果为'%s'，实际为'%s'", expectedText, result)
	}
}

// TestStreamReader_EmptyLine 测试流读取器对空行的处理
func TestStreamReader_EmptyLine(t *testing.T) {
	// 创建包含空行的模拟数据
	mockSSE := `

data: {"id":"cmpl-123","choices":[{"text":"测试"}]}


data: [DONE]
`
	
	// 创建模拟读取器
	mockBody := newMockReadCloser(mockSSE, nil)
	
	// 创建流读取器
	streamReader := NewStreamReader(mockBody)
	
	// 读取数据并验证
	resp, err := streamReader.Recv()
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	
	// 验证响应是否包含预期数据
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("响应中没有choices字段或为空")
	}
	
	// 获取第一个选择项
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0]不是一个有效的对象")
	}
	
	// 获取文本内容
	text, ok := choice["text"].(string)
	if !ok {
		t.Fatalf("choices[0].text不是一个有效的字符串")
	}
	
	// 验证文本内容
	if text != "测试" {
		t.Errorf("响应不匹配，期望'测试'，实际为'%s'", text)
	}
	
	// 验证读取完毕后是否返回EOF
	_, err = streamReader.Recv()
	if err != io.EOF {
		t.Errorf("期望EOF错误，实际为%v", err)
	}
}

// TestStreamReader_InvalidJSON 测试流读取器对无效JSON的处理
func TestStreamReader_InvalidJSON(t *testing.T) {
	// 创建包含无效JSON的模拟数据
	mockSSE := `
data: {"id":"cmpl-123",invalid json

data: {"id":"cmpl-123","choices":[{"text":"有效JSON"}]}

data: [DONE]
`
	
	// 创建模拟读取器
	mockBody := newMockReadCloser(mockSSE, nil)
	
	// 创建流读取器
	streamReader := NewStreamReader(mockBody)
	
	// 跳过无效的JSON行
	// 注意：在实际实现中，StreamReader应该能够跳过无效JSON
	
	// 读取有效的JSON数据
	resp, err := streamReader.Recv()
	if err != nil {
		// 如果在无效JSON时返回错误，这也是可接受的
		t.Logf("读取无效JSON时返回错误: %v", err)
		
		// 尝试读取下一个有效JSON
		resp, err = streamReader.Recv()
		if err != nil {
			t.Fatalf("读取有效JSON失败: %v", err)
		}
	}
	
	// 验证响应是否包含预期数据
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatalf("响应中没有choices字段或为空")
	}
	
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0]不是一个有效的JSON对象")
	}
	
	text, ok := choice["text"].(string)
	if !ok {
		t.Fatalf("choices[0].text不是一个有效的字符串")
	}
	
	if text != "有效JSON" {
		t.Errorf("期望文本为'有效JSON'，实际为'%s'", text)
	}
}
