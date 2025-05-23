package memory

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var (
	// ErrKeyNotFound 表示请求的键不存在
	ErrKeyNotFound = errors.New("内存键不存在")
)

// SimpleMemoryStore 简单内存存储实现
// 使用内存映射存储键值对
// 这是一个用于测试的简单实现
type SimpleMemoryStore struct {
	data  map[string]interface{} // 内存数据存储
	mutex sync.RWMutex           // 读写锁，保证并发安全
}

// NewSimpleMemoryStore 创建新的简单内存存储
func NewSimpleMemoryStore() *SimpleMemoryStore {
	return &SimpleMemoryStore{
		data: make(map[string]interface{}),
	}
}

// Save 实现Manager接口的Save方法
func (s *SimpleMemoryStore) Save(ctx context.Context, key string, value interface{}) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
	return nil
}

// Load 实现Manager接口的Load方法
func (s *SimpleMemoryStore) Load(ctx context.Context, key string) (interface{}, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	return value, nil
}

// Delete 实现Manager接口的Delete方法
func (s *SimpleMemoryStore) Delete(ctx context.Context, key string) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.data[key]; !exists {
		return ErrKeyNotFound
	}

	delete(s.data, key)
	return nil
}

// List 实现Manager接口的List方法
func (s *SimpleMemoryStore) List(ctx context.Context, prefix string) ([]string, error) {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var keys []string
	for key := range s.data {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Clear 实现Manager接口的Clear方法
func (s *SimpleMemoryStore) Clear(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data = make(map[string]interface{})
	return nil
}
