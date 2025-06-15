package embedder

// OpenAIEmbedder 使用 OpenAI Embeddings API 实现 rag.Embedder 接口
// 依赖环境变量 OPENAI_API_KEY 或传入 apiKey 参数
//
// 参考: https://platform.openai.com/docs/api-reference/embeddings

import (
    "context"
    "errors"
    "os"

    openai "github.com/sashabaranov/go-openai"
)

// 默认模型: text-embedding-3-small (兼顾价格/质量)
const defaultEmbeddingModel = "text-embedding-3-small"

type OpenAIEmbedder struct {
    client *openai.Client
    model  string
}

// NewOpenAIEmbedder 使用给定 APIKey (若为空读 OPENAI_API_KEY) 与模型名创建 Embedder
func NewOpenAIEmbedder(apiKey, model string) (*OpenAIEmbedder, error) {
    if apiKey == "" {
        apiKey = os.Getenv("OPENAI_API_KEY")
    }
    if apiKey == "" {
        return nil, errors.New("OpenAI API key not provided")
    }
    if model == "" {
        model = defaultEmbeddingModel
    }
    client := openai.NewClient(apiKey)
    return &OpenAIEmbedder{client: client, model: model}, nil
}

func (o *OpenAIEmbedder) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
    if len(texts) == 0 {
        return [][]float32{}, nil
    }

    req := openai.EmbeddingRequest{
        Model: o.model,
        Input: texts,
    }
    resp, err := o.client.CreateEmbeddings(ctx, req)
    if err != nil {
        return nil, err
    }
    if len(resp.Data) != len(texts) {
        return nil, errors.New("openai embeddings response size mismatch")
    }
    vectors := make([][]float32, len(resp.Data))
    for i, d := range resp.Data {
        v := make([]float32, len(d.Embedding))
        for j, f := range d.Embedding {
            v[j] = float32(f)
        }
        vectors[i] = v
    }
    return vectors, nil
}
