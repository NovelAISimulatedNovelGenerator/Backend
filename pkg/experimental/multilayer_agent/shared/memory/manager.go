package memory

import (
	"context"
)

// Manager 定义内存管理器接口
// 所有内存管理功能都通过此接口实现
type Manager interface {
	// Save 保存键值对到内存
	// ctx: 上下文，用于控制超时和取消
	// key: 内存键
	// value: 内存值(任意类型)
	// 返回: 错误信息
	Save(ctx context.Context, key string, value interface{}) error

	// Load 从内存加载值
	// ctx: 上下文，用于控制超时和取消
	// key: 内存键
	// 返回: 内存值和错误信息
	Load(ctx context.Context, key string) (interface{}, error)

	// Delete 从内存删除键值对
	// ctx: 上下文，用于控制超时和取消
	// key: 内存键
	// 返回: 错误信息
	Delete(ctx context.Context, key string) error

	// List 列出所有键
	// ctx: 上下文，用于控制超时和取消
	// prefix: 键前缀(可选)
	// 返回: 键列表和错误信息
	List(ctx context.Context, prefix string) ([]string, error)

	// Clear 清空内存
	// ctx: 上下文，用于控制超时和取消
	// 返回: 错误信息
	Clear(ctx context.Context) error
}
