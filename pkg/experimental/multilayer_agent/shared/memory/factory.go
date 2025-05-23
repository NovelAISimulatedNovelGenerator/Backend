package memory

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// MemoryType 内存存储类型
type MemoryType string

const (
	// MemoryTypeSimple 简单内存存储
	MemoryTypeSimple MemoryType = "simple"

	// 未来可以扩展其他类型的内存存储
	// MemoryTypePersistent 持久化内存存储
	// MemoryTypeDistributed 分布式内存存储
)

// NewMemoryManager 创建新的内存管理器
// 根据指定类型创建对应的内存管理器实现
// memType: 内存管理器类型
// 返回: 内存管理器实例
func NewMemoryManager(memType MemoryType) Manager {
	hlog.Infof("创建内存管理器: 类型=%s", memType)

	switch memType {
	case MemoryTypeSimple:
		return NewSimpleMemoryStore()
	default:
		hlog.Warnf("未知的内存管理器类型: %s，使用默认的简单内存存储", memType)
		return NewSimpleMemoryStore()
	}
}

// CreateTaggedKey 创建带标签的键
// 用于组织和分类内存中的数据
// agentID: 智能体ID
// category: 分类
// key: 原始键名
// 返回: 格式化的键名 "agentID:category:key"
func CreateTaggedKey(agentID, category, key string) string {
	return agentID + ":" + category + ":" + key
}

// ExtractKeyParts 从带标签的键中提取各部分
// taggedKey: 带标签的键 "agentID:category:key"
// 返回: agentID, category, key
func ExtractKeyParts(taggedKey string) (string, string, string) {
	parts := make([]string, 3)

	// 分割键字符串
	rawParts := strings.SplitN(taggedKey, ":", 3)

	// 复制可用部分
	copy(parts, rawParts)

	return parts[0], parts[1], parts[2]
}
