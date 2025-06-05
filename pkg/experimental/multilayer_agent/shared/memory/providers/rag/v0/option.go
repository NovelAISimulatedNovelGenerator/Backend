package rag

import "time"

// Config 保存可调参数
// ChunkSize/ChunkOverlap 影响文本切块
// DefaultTopK/MinScore 用于检索阶段的默认行为
// EmbeddingTimeout 控制 Embedder 的调用超时
// 通过 Option 模式进行灵活注入
// 所有字段均为值类型，Provider 初始化后视为只读
// 若需运行时修改，可在 Provider 上暴露对应 Setter
// 避免读取时竞态，本实现使用不可变配置
//
//go:generate go vet .
type Config struct {
    ChunkSize        int
    ChunkOverlap     int
    DefaultTopK      int
    MinScore         float32
    EmbeddingTimeout time.Duration
}

// Option 功能选项
// 典型调用：rag.NewProvider(store, embedder, rag.WithChunkSize(1024))
type Option func(*Config)

// defaultConfig 返回默认配置
func defaultConfig() *Config {
    return &Config{
        ChunkSize:        512,
        ChunkOverlap:     64,
        DefaultTopK:      4,
        MinScore:         0.0,
        EmbeddingTimeout: 30 * time.Second,
    }
}

// WithChunkSize 设置 ChunkSize
func WithChunkSize(size int) Option { return func(c *Config) { c.ChunkSize = size } }

// WithChunkOverlap 设置 ChunkOverlap
func WithChunkOverlap(overlap int) Option { return func(c *Config) { c.ChunkOverlap = overlap } }

// WithDefaultTopK 设置 DefaultTopK
func WithDefaultTopK(topK int) Option { return func(c *Config) { c.DefaultTopK = topK } }

// WithMinScore 设置 MinScore
func WithMinScore(score float32) Option { return func(c *Config) { c.MinScore = score } }

// WithEmbeddingTimeout 设置嵌入请求超时
func WithEmbeddingTimeout(d time.Duration) Option { return func(c *Config) { c.EmbeddingTimeout = d } }
