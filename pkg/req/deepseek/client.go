// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	
	"github.com/openai/openai-go"
)

// Client 是DeepSeek API的客户端
type Client struct {
	// config 是客户端配置
	config *Config
	
	// openaiClient 是OpenAI官方SDK的客户端实例
	openaiClient *openai.Client
}

// NewClient 创建一个新的DeepSeek客户端
func NewClient(apiKey string) (*Client, error) {
	config := DefaultConfig(apiKey)
	openaiClient, err := config.CreateClient()
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %w", err)
	}
	
	return &Client{
		config:       config,
		openaiClient: openaiClient,
	}, nil
}

// NewClientWithConfig 使用指定配置创建客户端
func NewClientWithConfig(config *Config) (*Client, error) {
	openaiClient, err := config.CreateClient()
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %w", err)
	}
	
	return &Client{
		config:       config,
		openaiClient: openaiClient,
	}, nil
}

// Models 获取可用模型列表
func (c *Client) Models(ctx context.Context) ([]string, error) {
	// 使用直接的 API 调用获取模型列表
	url := fmt.Sprintf("%s/models", c.config.BaseURL)
	response, err := c.sendJSONRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %w", err)
	}
	
	// 处理响应
	var models []string
	if data, ok := response["data"].([]interface{}); ok {
		for _, item := range data {
			if model, ok := item.(map[string]interface{}); ok {
				if id, ok := model["id"].(string); ok {
					models = append(models, id)
				}
			}
		}
	}
	
	return models, nil
}

// Completion 发送非流式文本生成请求
func (c *Client) Completion(ctx context.Context, request *CompletionRequest) (map[string]interface{}, error) {
	// 确保不是流式请求
	request.Stream = false
	
	// 拼接 beta 路径，保证 completions 只用 beta
	url := fmt.Sprintf("%s/beta/completions", strings.TrimRight(c.config.BaseURL, "/"))
	response, err := c.sendJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		return nil, fmt.Errorf("文本生成请求失败: %w", err)
	}
	
	return response, nil
}

// ChatCompletion 发送非流式聊天完成请求
func (c *Client) ChatCompletion(ctx context.Context, request *ChatRequest) (map[string]interface{}, error) {
	// 确保不是流式请求
	request.Stream = false
	
	// 拼接 v1 路径，chat 只用 v1
	url := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(c.config.BaseURL, "/"))
	response, err := c.sendJSONRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		return nil, fmt.Errorf("聊天请求失败: %w", err)
	}
	
	return response, nil
}

// CompletionStream 发送流式文本生成请求
func (c *Client) CompletionStream(ctx context.Context, request *CompletionRequest) (*StreamReader, error) {
	// 确保是流式请求
	request.Stream = true
	
	// 拼接 beta 路径，保证 completions stream 只用 beta
	url := fmt.Sprintf("%s/beta/completions", strings.TrimRight(c.config.BaseURL, "/"))
	resp, err := c.sendStreamRequest(ctx, url, request)
	if err != nil {
		return nil, fmt.Errorf("流式文本生成请求失败: %w", err)
	}
	
	return NewStreamReader(resp.Body), nil
}

// ChatCompletionStream 发送流式聊天完成请求
func (c *Client) ChatCompletionStream(ctx context.Context, request *ChatRequest) (*StreamReader, error) {
	// 确保是流式请求
	request.Stream = true
	
	// 拼接 v1 路径，chat stream 只用 v1
	url := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(c.config.BaseURL, "/"))
	resp, err := c.sendStreamRequest(ctx, url, request)
	if err != nil {
		return nil, fmt.Errorf("流式聊天请求失败: %w", err)
	}
	
	return NewStreamReader(resp.Body), nil
}

// sendJSONRequest 发送JSON请求并解析响应
func (c *Client) sendJSONRequest(ctx context.Context, method, url string, body interface{}) (map[string]interface{}, error) {
	// 将请求体编码为JSON
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("User-Agent", c.config.UserAgent)
	
	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
	
	// 发送请求
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	// 检查响应状态码
	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("API错误: %v (状态码: %d)", errResp, resp.StatusCode)
		}
		return nil, fmt.Errorf("API错误 (状态码: %d): %s", resp.StatusCode, string(respBody))
	}
	
	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return result, nil
}

// sendStreamRequest 发送流式请求
func (c *Client) sendStreamRequest(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	// 将请求体编码为JSON
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "text/event-stream")
	
	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
	
	// 发送请求
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	
	// 检查响应状态码
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		
		var errResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("API错误: %v (状态码: %d)", errResp, resp.StatusCode)
		}
		return nil, fmt.Errorf("API错误 (状态码: %d): %s", resp.StatusCode, string(respBody))
	}
	
	return resp, nil
}

// StreamReader 是流式响应的读取器
type StreamReader struct {
	// reader 是用于读取流的缓冲读取器
	reader *bufio.Reader
	
	// isFinished 标记流是否已结束
	isFinished bool
	
	// body 是HTTP响应体
	body io.ReadCloser
}

// bufio包已在导入中声明

// NewStreamReader 创建新的流读取器
func NewStreamReader(body io.ReadCloser) *StreamReader {
	return &StreamReader{
		reader:     bufio.NewReader(body),
		isFinished: false,
		body:      body,
	}
}

// Close 关闭流读取器
func (s *StreamReader) Close() error {
	s.isFinished = true
	return s.body.Close()
}

// Recv 从流中接收下一个事件
func (s *StreamReader) Recv() (map[string]interface{}, error) {
	if s.isFinished {
		return nil, io.EOF
	}
	
	for {
		// 读取一行
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			s.isFinished = true
			return nil, err
		}
		
		// 删除前后空白
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		
		// 检查是否是数据前缀
		const prefix = "data: "
		if !bytes.HasPrefix(line, []byte(prefix)) {
			// 检查是否是结束标记
			if bytes.Equal(line, []byte("[DONE]")) || strings.HasPrefix(string(line), "data: [DONE]") {
				s.isFinished = true
				return nil, io.EOF
			}
			continue
		}
		
		// 删除数据前缀
		data := bytes.TrimPrefix(line, []byte(prefix))
		
		// 解析JSON
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			continue
		}
		
		return response, nil
	}
}
