// Package rag 提供基于 langchaingo 的 RAG（Retrieval-Augmented Generation）Provider 实现。
//
// 设计目标：
// 1. 抽象向量存储（VectorStore）、嵌入模型（Embedder）、大语言模型（LLM）接口，支持多实现热插拔
// 2. 采用 Option 方式注入可调配置，例如 ChunkSize、ChunkOverlap 等
// 3. 默认使用 langchaingo 的 textsplitter 进行文本切块，ChunkSize/Overlap 可调
// 4. 暴露 IndexDocuments / Retrieve / Generate 等核心方法，已实现 Top-K+Score 检索，预留拓展搜索接口
// 5. 内置线程安全 MemoryStore 供单元测试与轻量场景，后续可接入 pgvector、Pinecone、Milvus 等
// 6. 全量日志输出统一使用 Hertz 的 hlog 包，方便服务端接入
// 7. 对外仅暴露稳定的 v0 版本包路径，未来升级时通过 v1、v2 目录并行演进
package rag
