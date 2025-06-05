//go:build unit
// +build unit

package tests

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    rag "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0"
    "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0/embedder"
    "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0/store"
)

func TestProvider_IndexAndRetrieve(t *testing.T) {
    p, err := rag.NewProvider(store.NewMemoryStore(), embedder.MockEmbedder{})
    assert.NoError(t, err)

    ctx := context.Background()
    docs := []string{"hello world", "foo bar"}
    metas := []map[string]string{{"source": "unit"}}
    _ = p.IndexDocuments(ctx, docs, metas)

    recs, err := p.Retrieve(ctx, "hello", 2, 0)
    assert.NoError(t, err)
    assert.NotEmpty(t, recs)
}
