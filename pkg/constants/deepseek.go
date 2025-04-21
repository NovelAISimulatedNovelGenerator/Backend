// Package constants 统一管理项目常量
// DeepSeek 相关常量
package constants

// DeepSeek 模型名称常量
const (
	DeepSeekChat    = "deepseek-chat"           // DeepSeek 通用聊天模型
	DeepSeekCoder   = "deepseek-coder"          // DeepSeek 代码生成模型
	DeepSeekMax     = "deepseek-llm-67b-max"    // DeepSeek 大参数模型
	DeepSeek7B      = "deepseek-llm-7b-base"    // DeepSeek 7B 参数模型
)

// DeepSeek 角色常量，与 OpenAI 兼容
const (
	RoleSystem    = "system"    // 系统角色
	RoleUser      = "user"      // 用户角色
	RoleAssistant = "assistant" // 助手角色
)

// DeepSeek 默认配置常量
const (
	DefaultDeepSeekBaseURL = "https://api.deepseek.com/v1" // 默认 API 基础 URL
)
