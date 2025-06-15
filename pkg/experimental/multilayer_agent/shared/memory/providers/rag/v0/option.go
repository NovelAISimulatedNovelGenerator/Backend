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

// Option 用于在创建 Provider 时自定义配置或依赖
// 典型调用：rag.NewProvider(store, embedder,
//     rag.WithChunkSize(1024),
//     rag.WithLLM(myLLM),
//     rag.WithSplitter(customSplitter),
// )
// Option 回调拿到 *Provider，因此既可修改 p.cfg，也可直接注入依赖
type Option func(*Provider)

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

// Below are concrete Option helpers ------------------------------------------------

func WithChunkSize(size int) Option {
    return func(p *Provider) {
        p.cfg.ChunkSize = size
        // 更新 splitter 以匹配新的 chunk 参数
        p.splitter = defaultSplitter(p.cfg.ChunkSize, p.cfg.ChunkOverlap)
    }
}

func WithChunkOverlap(overlap int) Option {
    return func(p *Provider) {
        p.cfg.ChunkOverlap = overlap
        p.splitter = defaultSplitter(p.cfg.ChunkSize, p.cfg.ChunkOverlap)
    }
}

func WithDefaultTopK(topK int) Option { return func(p *Provider) { p.cfg.DefaultTopK = topK } }

func WithMinScore(score float32) Option { return func(p *Provider) { p.cfg.MinScore = score } }

func WithEmbeddingTimeout(d time.Duration) Option { return func(p *Provider) { p.cfg.EmbeddingTimeout = d } }

// WithLLM 注入 LLM 依赖
func WithLLM(llm LLM) Option { return func(p *Provider) { p.llm = llm } }

// WithSplitter 注入自定义文本切分器
func WithSplitter(s Splitter) Option { return func(p *Provider) { p.splitter = s } }
