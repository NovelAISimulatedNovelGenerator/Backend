// Package tools 实现多层代理系统的工具调用功能
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// ToolRequest 表示工具调用请求
// 由大语言模型生成，用于指定要调用的工具和输入参数
type ToolRequest struct {
	// ToolName 要调用的工具名称
	ToolName string          `json:"tool_name"`
	// Input 工具输入，可以是简单字符串或复杂JSON
	Input    json.RawMessage `json:"input"`
}

// ToolResponse 表示工具调用响应
// 返回给大语言模型，包含调用结果或错误信息
type ToolResponse struct {
	// ToolName 调用的工具名称
	ToolName string `json:"tool_name"`
	// Result 工具返回的结果
	Result   string `json:"result"`
	// Error 如果调用失败，这里包含错误信息
	Error    string `json:"error,omitempty"`
	// Success 调用是否成功
	Success  bool   `json:"success"`
}

// ToolCaller 处理工具调用请求
// 作为智能体和工具之间的中间层，负责路由和执行工具调用
type ToolCaller struct {
	// registry 工具注册表，包含所有可用工具
	registry *ToolRegistry
}

// NewToolCaller 创建新的工具调用处理器
// 参数:
//   - registry: 工具注册表，包含所有可用工具
// 返回:
//   - *ToolCaller: 新创建的工具调用处理器
func NewToolCaller(registry *ToolRegistry) *ToolCaller {
	return &ToolCaller{
		registry: registry,
	}
}

// CallTool 执行工具调用
// 参数:
//   - ctx: 上下文，包含调用相关信息
//   - req: 工具调用请求
// 返回:
//   - *ToolResponse: 工具调用响应
//   - error: 处理过程错误，如果有
func (c *ToolCaller) CallTool(ctx context.Context, req ToolRequest) (*ToolResponse, error) {
	// 记录接收到的工具调用请求
	hlog.CtxInfof(ctx, "接收到工具调用请求: %s", req.ToolName)
	
	// 从注册表获取工具
	tool, err := c.registry.GetTool(req.ToolName)
	if err != nil {
		// 工具不存在，返回错误响应
		return &ToolResponse{
			ToolName: req.ToolName,
			Error:    fmt.Sprintf("工具不存在: %v", err),
			Success:  false,
		}, nil
	}
	
	// 处理输入参数
	var input string
	if req.Input == nil {
		// 如果输入为空，使用空字符串
		input = ""
	} else {
		// 尝试将JSON解析为字符串
		err = json.Unmarshal(req.Input, &input)
		if err != nil {
			// 如果解析失败，直接使用原始JSON字符串
			input = string(req.Input)
		}
	}
	
	// 创建适配器并调用工具
	adapter := NewLangChainAdapter(tool)
	result, err := adapter.Call(ctx, input)
	if err != nil {
		// 工具调用失败，返回错误响应
		return &ToolResponse{
			ToolName: req.ToolName,
			Error:    fmt.Sprintf("工具调用失败: %v", err),
			Success:  false,
		}, nil
	}
	
	// 返回成功响应
	return &ToolResponse{
		ToolName: req.ToolName,
		Result:   result,
		Success:  true,
	}, nil
}

// CallToolFromJSON 从JSON字符串执行工具调用
// 参数:
//   - ctx: 上下文，包含调用相关信息
//   - jsonStr: 包含工具调用请求的JSON字符串
// 返回:
//   - *ToolResponse: 工具调用响应
//   - error: 处理过程错误，如果有
func (c *ToolCaller) CallToolFromJSON(ctx context.Context, jsonStr string) (*ToolResponse, error) {
	// 解析JSON请求
	var req ToolRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return nil, fmt.Errorf("无效的工具调用JSON: %v", err)
	}
	
	// 调用工具
	return c.CallTool(ctx, req)
}
