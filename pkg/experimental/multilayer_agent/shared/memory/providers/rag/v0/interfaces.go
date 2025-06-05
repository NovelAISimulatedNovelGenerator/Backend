package rag

import (
    "context"

    "github.com/tmc/langchaingo/textsplitter"
    "novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// VectorStore 向量存储接口
// Upsert 应保证同 ID 覆盖写，或由实现自行生成 ID
// Query 返回的 Record 中 Score 字段需填充相似度，便于上层排序
// topK 与 minScore 由调用者传入，VectorStore 可再次校验取值范围
// 所有实现必须并发安全
// 返回切片长度不得大于 topK
// 若暂无结果，可返回空切片而不是 nil 以避免调用处判空
//
//go:generate go vet .
type VectorStore interface {
    Upsert(ctx context.Context, records []*memory.Record) error
    Query(ctx context.Context, vector []float32, topK int, minScore float32) ([]memory.Record, error)
}

// Embedder 文本向量接口
// EmbedTexts 返回与输入文本长度一致的向量切片
// 若底层模型报错，应返回 error 并保持 nil 或空向量切片
// 向量长度必须一致
//
//go:generate go vet .
type Embedder interface {
    EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
}

// LLM 生成接口（可选依赖）
// Generate 输入 prompt，输出完整 LLM 响应
// 错误时返回空字符串+error
//
//go:generate go vet .
type LLM interface {
    Generate(ctx context.Context, prompt string) (string, error)
}

// Splitter 文本切块接口，默认使用 langchaingo.textsplitter
// 允许注入自定义实现（例如 MarkdownSplitter）
//
//go:generate go vet .
type Splitter interface {
    SplitText(text string) ([]string, error)
}

// defaultSplitter 返回基于 TokenSplitter 的默认实现
func defaultSplitter(chunkSize, overlap int) Splitter {
    return textsplitter.NewTokenSplitter(chunkSize, overlap)
}
