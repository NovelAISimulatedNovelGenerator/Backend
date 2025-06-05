package store

// 内存向量存储（线程安全，仅供测试或小规模场景）
// 写操作 Append，不做去重；ID 由外部控制。
// 查询实现最简余弦相似度，适配演示 TopK & MinScore 功能。

import (
    "context"
    "math"
    "sort"
    "sync"

    "novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// MemoryStore 线程安全的向量存储
//go:generate go vet .
type MemoryStore struct {
    mu      sync.RWMutex
    records []*memory.Record
}

// NewMemoryStore 创建实例
func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

// Upsert 直接追加记录（演示用）
func (m *MemoryStore) Upsert(_ context.Context, recs []*memory.Record) error {
    if len(recs) == 0 {
        return nil
    }
    m.mu.Lock()
    m.records = append(m.records, recs...)
    m.mu.Unlock()
    return nil
}

// Query 按余弦相似度排序返回 topK
func (m *MemoryStore) Query(_ context.Context, vec []float32, topK int, minScore float32) ([]memory.Record, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    type scored struct {
        rec   memory.Record
        score float32
    }
    list := make([]scored, 0, len(m.records))
    for _, r := range m.records {
        if len(r.Embedding) == 0 {
            continue
        }
        s := cosine(vec, r.Embedding)
        if s < minScore {
            continue
        }
        recCopy := *r
        recCopy.Score = s
        list = append(list, scored{recCopy, s})
    }

    sort.Slice(list, func(i, j int) bool { return list[i].score > list[j].score })
    if len(list) > topK {
        list = list[:topK]
    }
    results := make([]memory.Record, len(list))
    for i, s := range list {
        results[i] = s.rec
    }
    return results, nil
}

// cosine 计算向量余弦相似度（假设两个向量长度一致）
func cosine(a, b []float32) float32 {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    n := len(a)
    if len(b) < n {
        n = len(b)
    }
    var dot, normA, normB float32
    for i := 0; i < n; i++ {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / float32(math.Sqrt(float64(normA*normB)))
}
