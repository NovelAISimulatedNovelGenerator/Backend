//go:build integration
// +build integration

package store

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// TestPostgresStore_UpsertQuery 在真实 PostgreSQL + pgvector 上验证基本 CRUD
// 运行前请设置环境变量 PG_DSN，例如：
// PG_DSN="host=localhost port=5432 user=postgres password=secret dbname=rag_test sslmode=disable"
// 并确保数据库已安装 pgvector 扩展。
func TestPostgresStore_UpsertQuery(t *testing.T) {
    dsn := os.Getenv("PG_DSN")
    if dsn == "" {
        t.Skip("PG_DSN not set; skipping Postgres integration test")
    }
    store, err := NewPostgresStore(dsn)
    if err != nil {
        t.Fatalf("create store: %v", err)
    }

    ctx := context.Background()
    rec := &memory.Record{
        Text:      "hello world",
        Embedding: []float32{0.1, 0.2, 0.3},
        Metadata:  map[string]string{"source": "test"},
    }
    assert.NoError(t, store.Upsert(ctx, []*memory.Record{rec}))

    res, err := store.Query(ctx, []float32{0.1, 0.2, 0.3}, 1, 0)
    assert.NoError(t, err)
    assert.NotEmpty(t, res)
    if len(res) > 0 {
        assert.Equal(t, "hello world", res[0].Text)
    }
}
