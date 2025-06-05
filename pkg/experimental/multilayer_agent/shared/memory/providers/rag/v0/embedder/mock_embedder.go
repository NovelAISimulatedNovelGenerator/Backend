package embedder

import "context"

// MockEmbedder 用于单元测试，返回长度为 1 的伪向量（文本长度）
//go:generate go vet .
type MockEmbedder struct{}

func (m MockEmbedder) EmbedTexts(_ context.Context, texts []string) ([][]float32, error) {
    out := make([][]float32, len(texts))
    for i, t := range texts {
        out[i] = []float32{float32(len(t))}
    }
    return out, nil
}
