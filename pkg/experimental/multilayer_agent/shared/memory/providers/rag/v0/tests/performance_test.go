//go:build perf
// +build perf

package tests

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    rag "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0"
    "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0/embedder"
    "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0/store"
)

// TestProvider_HighConcurrency 进行高并发压力测试并输出性能指标
// 通过 build tag "perf" 启用: go test -tags=perf -run TestProvider_HighConcurrency
func TestProvider_HighConcurrency(t *testing.T) {
    if testing.Short() {
        t.Skip("skip perf test in short mode")
    }

    p, err := rag.NewProvider(store.NewMemoryStore(), embedder.MockEmbedder{})
    assert.NoError(t, err)

    // 预索引 1k 文档，保证 Retrieve 有数据
    baseDocs := make([]string, 1000)
    for i := range baseDocs {
        baseDocs[i] = fmt.Sprintf("document %d", i)
    }
    assert.NoError(t, p.IndexDocuments(context.Background(), baseDocs, nil))

    const workers = 100      // 并发 goroutine 数
    const opsPerWorker = 500 // 每个 goroutine 操作次数

    var totalOps uint64
    wg := sync.WaitGroup{}
    wg.Add(workers)

    start := time.Now()

    for w := 0; w < workers; w++ {
        go func(id int) {
            defer wg.Done()
            ctx := context.Background()
            for i := 0; i < opsPerWorker; i++ {
                if i%2 == 0 {
                    // 50% 进行索引
                    _ = p.IndexDocuments(ctx, []string{fmt.Sprintf("new doc %d-%d", id, i)}, nil)
                } else {
                    // 50% 进行检索
                    _, _ = p.Retrieve(ctx, "document", 5, 0)
                }
                atomic.AddUint64(&totalOps, 1)
            }
        }(w)
    }

    wg.Wait()
    elapsed := time.Since(start)
    t.Logf("High-concurrency test finished: goroutines=%d totalOps=%d elapsed=%s ops/sec=%.2f", workers, totalOps, elapsed, float64(totalOps)/elapsed.Seconds())
}
