package rag

import (
    "context"
    "fmt"
    "strings"
    "sync"

    "github.com/cloudwego/hertz/pkg/common/hlog"
    "novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// Provider 实现核心 RAG 能力
// 并发安全：通过 mu 保护 splitter 可变场景；store/embedder 自身需保证并发安全
//
//go:generate go vet .
type Provider struct {
    cfg      *Config
    store    VectorStore
    embedder Embedder
    llm      LLM
    splitter Splitter
    mu       sync.RWMutex
}

// NewProvider 初始化 Provider
// 必须传入 VectorStore 与 Embedder；其余使用默认
// opts 可修改 Config 或注入自定义 Splitter/LLM
func NewProvider(store VectorStore, embedder Embedder, opts ...Option) (*Provider, error) {
    if store == nil || embedder == nil {
        return nil, ErrInvalidDependency
    }

    cfg := defaultConfig()

    p := &Provider{
        cfg:      cfg,
        store:    store,
        embedder: embedder,
        splitter: defaultSplitter(cfg.ChunkSize, cfg.ChunkOverlap),
    }

    for _, o := range opts {
        o(p)
    }

    return p, nil
}

// IndexDocuments 将文档切块后写入向量数据库
func (p *Provider) IndexDocuments(ctx context.Context, docs []string, metas []map[string]string) error {
    if len(docs) == 0 {
        return nil
    }

    p.mu.RLock()
    splitter := p.splitter
    p.mu.RUnlock()

    var chunks []string
    for _, d := range docs {
        cs, err := splitter.SplitText(d)
        if err != nil {
            hlog.CtxWarnf(ctx, "split text error: %v", err)
            continue
        }
        chunks = append(chunks, cs...)
    }

    vecs, err := p.embedder.EmbedTexts(ctx, chunks)
    if err != nil {
        hlog.CtxErrorf(ctx, "embed texts error: %v", err)
        return err
    }
    if len(vecs) != len(chunks) {
        hlog.CtxErrorf(ctx, "embedder returned mismatched size: want %d got %d", len(chunks), len(vecs))
        return ErrInvalidDependency
    }

    records := make([]*memory.Record, len(chunks))
    for i, c := range chunks {
        var meta map[string]string
        if len(metas) > 0 {
            meta = metas[i%len(metas)]
        }
        records[i] = &memory.Record{
            Text:      c,
            Embedding: vecs[i],
            Metadata:  meta,
        }
    }

    return p.store.Upsert(ctx, records)
}

// Retrieve 检索相似段落
func (p *Provider) Retrieve(ctx context.Context, query string, topK int, minScore float32) ([]memory.Record, error) {
    if topK <= 0 {
        topK = p.cfg.DefaultTopK
    }
    if minScore == 0 {
        minScore = p.cfg.MinScore
    }

    vecs, err := p.embedder.EmbedTexts(ctx, []string{query})
    if err != nil {
        hlog.CtxErrorf(ctx, "embed query error: %v", err)
        return nil, err
    }
    return p.store.Query(ctx, vecs[0], topK, minScore)
}

// Generate 检索 + LLM 生成
func (p *Provider) Generate(ctx context.Context, query string, topK int) (string, error) {
    p.mu.RLock()
    llm := p.llm
    p.mu.RUnlock()

    if llm == nil {
        return "", ErrLLMNotProvided
    }
    recs, err := p.Retrieve(ctx, query, topK, p.cfg.MinScore)
    if err != nil {
        return "", err
    }

    prompt := BuildPrompt(query, recs)
    return llm.Generate(ctx, prompt)
}

// BuildPrompt 简单将检索结果串联后拼接查询
func BuildPrompt(query string, recs []memory.Record) string {
    var sb strings.Builder
    sb.WriteString("请根据以下知识回答问题:\n")
    for i, r := range recs {
        sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Text))
    }
    sb.WriteString("\n问题: ")
    sb.WriteString(query)
    sb.WriteString("\n答案: ")
    return sb.String()
}

// SetLLM 设置 LLM 依赖；可在运行时注入
func (p *Provider) SetLLM(llm LLM) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.llm = llm
}
