// Package deepseek 提供了与DeepSeek API交互的功能，基于OpenAI官方SDK
package deepseek

import (
	"context"
	"net/http"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	// DefaultDeepSeekBaseURL 是DeepSeek API的默认基础URL
	DefaultDeepSeekBaseURL = "https://api.deepseek.com/v1"

	// DefaultTimeout 是HTTP请求的默认超时时间
	DefaultTimeout = 30 * time.Second
)

// Config 存储DeepSeek API客户端配置
type Config struct {
	// BaseURL 是DeepSeek API的基础URL
	BaseURL string

	// APIKey 是DeepSeek API的认证密钥
	APIKey string

	// OrgID 是组织ID（可选）
	OrgID string

	// Timeout 是请求超时时间
	Timeout time.Duration

	// HTTPClient 是用于发送HTTP请求的客户端
	HTTPClient *http.Client

	// UserAgent 是请求的User-Agent头
	UserAgent string
}

// DefaultConfig 返回一个默认的配置
func DefaultConfig(apiKey string) *Config {
	return &Config{
		BaseURL:    "https://api.deepseek.com/v1",
		APIKey:     apiKey,
		Timeout:    30 * time.Second,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		UserAgent:  "deepseek-go/1.0.0",
	}
}

// WithBaseURL 设置基础URL
func (c *Config) WithBaseURL(baseURL string) *Config {
	c.BaseURL = baseURL
	return c
}

// WithOrgID 设置组织ID
func (c *Config) WithOrgID(orgID string) *Config {
	c.OrgID = orgID
	return c
}

// WithTimeout 设置超时时间
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	c.HTTPClient.Timeout = timeout
	return c
}

// WithHTTPClient 设置自定义HTTP客户端
func (c *Config) WithHTTPClient(client *http.Client) *Config {
	c.HTTPClient = client
	return c
}

// WithUserAgent 设置用户代理字符串
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// CreateClient 创建一个OpenAI SDK客户端
func (c *Config) CreateClient() (*openai.Client, error) {
	// 准备选项
	opts := []option.RequestOption{}
	
	// 添加基础URL
	if c.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(c.BaseURL))
	}
	
	// 添加自定义HTTP客户端
	if c.HTTPClient != nil {
		opts = append(opts, option.WithHTTPClient(c.HTTPClient))
	}
	
	// 添加组织ID
	if c.OrgID != "" {
		opts = append(opts, option.WithOrganization(c.OrgID))
	}
	
	// 设置自定义头部信息（包括User-Agent）
	if c.UserAgent != "" {
		opts = append(opts, option.WithHeader("User-Agent", c.UserAgent))
	}
	
	// 添加API密钥
	opts = append(opts, option.WithAPIKey(c.APIKey))
	
	// 创建客户端
	client := openai.NewClient(opts...)
	
	return &client, nil
}

// TestConnection 测试连接到DeepSeek API
func (c *Config) TestConnection() error {
	// 创建客户端
	client, err := c.CreateClient()
	if err != nil {
		return err
	}

	// 测试连接
	_, err = client.Models.List(context.Background(), nil)
	return err
}
