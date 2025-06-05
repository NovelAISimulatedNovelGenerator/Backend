package rag

import "errors"

var (
    // ErrInvalidDependency 必要依赖未注入
    ErrInvalidDependency = errors.New("invalid dependency: vector store and embedder are required")
    // ErrLLMNotProvided 调用生成时未注入 LLM
    ErrLLMNotProvided  = errors.New("llm not provided for generate operation")
)
