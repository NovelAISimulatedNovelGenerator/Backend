package llm

// OpenAILLM 使用 OpenAI ChatCompletion 实现 rag.LLM 接口
// 依赖环境变量 OPENAI_API_KEY

import (
    "context"
    "errors"
    "os"

    openai "github.com/sashabaranov/go-openai"
    rag "novelai/pkg/experimental/multilayer_agent/shared/memory/providers/rag/v0"
)

const defaultChatModel = "gpt-3.5-turbo-0125"

type OpenAILLM struct {
    client *openai.Client
    model  string
}

func NewOpenAILLM(apiKey, model string) (*OpenAILLM, error) {
    if apiKey == "" {
        apiKey = os.Getenv("OPENAI_API_KEY")
    }
    if apiKey == "" {
        return nil, errors.New("OpenAI API key not provided")
    }
    if model == "" {
        model = defaultChatModel
    }
    return &OpenAILLM{client: openai.NewClient(apiKey), model: model}, nil
}

func (o *OpenAILLM) Generate(ctx context.Context, prompt string) (string, error) {
    req := openai.ChatCompletionRequest{
        Model: o.model,
        Messages: []openai.ChatCompletionMessage{
            {Role: "user", Content: prompt},
        },
    }
    resp, err := o.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", err
    }
    if len(resp.Choices) == 0 {
        return "", errors.New("no choices returned")
    }
    return resp.Choices[0].Message.Content, nil
}

// ensure interface compliance
var _ rag.LLM = (*OpenAILLM)(nil)
