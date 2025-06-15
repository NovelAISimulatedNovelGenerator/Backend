package store

// PostgresStore 基于 PostgreSQL + pgvector 的向量存储实现
// 依赖数据库已安装 pgvector 扩展 (CREATE EXTENSION IF NOT EXISTS vector)
// 表模式: rag_embeddings(id SERIAL PRIMARY KEY, text TEXT, embedding VECTOR(???), metadata JSONB)
// 其中 embedding 维度由首条写入时决定 (pgvector 会校验长度一致)
//
//go:generate go vet .

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "sync"

    "github.com/cloudwego/hertz/pkg/common/hlog"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "novelai/pkg/experimental/multilayer_agent/shared/memory"
)

// EmbeddingDBModel GORM 映射
// 表名通过 TableName() 指定，默认 "rag_embeddings"
type EmbeddingDBModel struct {
    ID        uint            `gorm:"primaryKey"`
    Text      string          `gorm:"type:text"`
    Embedding string          `gorm:"type:vector"` // 存储为 "[1,2,3]" 字符串
    Metadata  json.RawMessage `gorm:"type:jsonb"`
}

func (EmbeddingDBModel) TableName() string { return "rag_embeddings" }

// PostgresStore 实现 VectorStore 接口
// 内部使用 RWMutex 保护 *gorm.DB 的 Init (gorm 本身连接池线程安全)

type PostgresStore struct {
    db   *gorm.DB
    mu   sync.RWMutex
    dims int // 向量维度 (首条写入决定)
}

// dsnTemplate 用于快速拼接 Postgres DSN，可选使用
const dsnTemplate = "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"

// NewPostgresStore 创建连接并初始化表/扩展

func NewPostgresStore(dsn string) (*PostgresStore, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    // init extension & table
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
        return nil, err
    }
    if err := db.AutoMigrate(&EmbeddingDBModel{}); err != nil {
        return nil, err
    }
    return &PostgresStore{db: db}, nil
}

// helper 转换向量 slice -> '[1,2,3]' 字符串
func vecToPg(v []float32) string {
    parts := make([]string, len(v))
    for i, f := range v {
        parts[i] = fmt.Sprintf("%f", f)
    }
    return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

func (p *PostgresStore) Upsert(ctx context.Context, recs []*memory.Record) error {
    for _, r := range recs {
        metaJSON, _ := json.Marshal(r.Metadata)
        model := EmbeddingDBModel{Text: r.Text, Embedding: vecToPg(r.Embedding), Metadata: metaJSON}
        // INSERT ... ON CONFLICT(id) DO UPDATE SET ...  因简化: 直接插入匿名主键
        if err := p.db.WithContext(ctx).Create(&model).Error; err != nil {
            hlog.CtxErrorf(ctx, "pg Upsert error: %v", err)
            return err
        }
    }
    return nil
}

func (p *PostgresStore) Query(ctx context.Context, vec []float32, topK int, minScore float32) ([]memory.Record, error) {
    queryVec := vecToPg(vec)
    sql := `SELECT id, text, embedding, metadata, 1 - (embedding <=> $1)::float AS score
            FROM rag_embeddings
            ORDER BY embedding <=> $1
            LIMIT $2`
    type row struct {
        ID        uint
        Text      string
        Embedding string
        Metadata  json.RawMessage
        Score     float32
    }
    var rows []row
    if err := p.db.WithContext(ctx).Raw(sql, queryVec, topK).Scan(&rows).Error; err != nil {
        return nil, err
    }
    results := make([]memory.Record, 0, len(rows))
    for _, r := range rows {
        if r.Score < minScore {
            continue
        }
        var meta map[string]string
        _ = json.Unmarshal(r.Metadata, &meta)
        results = append(results, memory.Record{ID: fmt.Sprintf("%d", r.ID), Text: r.Text, Metadata: meta, Score: r.Score})
    }
    return results, nil
}
