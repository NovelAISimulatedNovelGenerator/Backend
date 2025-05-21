// Package tools 实现多层代理系统的工具调用功能
package tools

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/tmc/langchaingo/tools"
)

// LangChainAdapter 将LangChain工具适配到我们的系统
// 这是一个适配器模式的实现，用于兼容第三方工具库
type LangChainAdapter struct {
	// originalTool 保存原始的LangChain工具实例
	originalTool tools.Tool
}

// NewLangChainAdapter 创建新的LangChain工具适配器
// 参数:
//   - tool: 要适配的LangChain工具
//
// 返回:
//   - *LangChainAdapter: 新创建的适配器实例
func NewLangChainAdapter(tool tools.Tool) *LangChainAdapter {
	return &LangChainAdapter{
		originalTool: tool,
	}
}

// Name 返回工具名称
// 返回:
//   - string: 工具的名称
func (a *LangChainAdapter) Name() string {
	return a.originalTool.Name()
}

// Description 返回工具描述
// 返回:
//   - string: 工具的描述信息
func (a *LangChainAdapter) Description() string {
	return a.originalTool.Description()
}

// Call 使用我们系统的上下文调用原始工具
// 参数:
//   - ctx: 上下文，包含调用相关信息
//   - input: 工具输入字符串
//
// 返回:
//   - string: 工具调用结果
//   - error: 工具调用错误，如果有
func (a *LangChainAdapter) Call(ctx context.Context, input string) (string, error) {
	// 记录工具调用开始
	hlog.CtxInfof(ctx, "调用工具 %s，输入: %s", a.Name(), input)

	// 调用原始工具
	result, err := a.originalTool.Call(ctx, input)
	if err != nil {
		// 记录工具调用失败
		hlog.CtxErrorf(ctx, "工具调用失败 %s: %v", a.Name(), err)
		return "", err
	}

	// 记录工具调用成功
	hlog.CtxInfof(ctx, "工具调用成功 %s，结果长度: %d字节", a.Name(), len(result))
	return result, nil
}
