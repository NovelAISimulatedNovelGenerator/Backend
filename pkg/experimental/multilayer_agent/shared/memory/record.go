package memory

// Record 表示存储在向量数据库中的一条文本与向量记录
// ID 可选，由具体 VectorStore 决定是否生成
// Score 在检索结果中使用，表示与查询向量的相似度分值
// Embedding 长度与具体嵌入模型一致，通常为 1536/768 等
// Metadata 允许存储额外标签，便于后续过滤或排序
// NOTE: 该结构体仅用于 RAG 相关功能，其他内存键值对不受影响
//go:generate go run golang.org/x/tools/cmd/stringer -type Record
// (stringer 指令仅做示例，无实际生成需求，可按需删除)
type Record struct {
    ID        string             `json:"id"`
    Text      string             `json:"text"`
    Embedding []float32          `json:"embedding"`
    Metadata  map[string]string  `json:"metadata,omitempty"`
    Score     float32            `json:"score,omitempty"`
}
