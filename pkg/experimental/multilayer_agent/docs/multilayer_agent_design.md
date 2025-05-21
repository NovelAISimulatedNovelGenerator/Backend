# NovelAI 分层智能体系统设计

## 1. 整体架构

NovelAI 分层智能体系统采用"决策-执行"双层架构，通过清晰的职责分离实现小说内容的高质量生成。

```
┌───────────────────────────────────────────────────────────────┐
│                     决策层 (Decision Layer)                    │
│                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌────────────┐ │
│  │  策略Agent      │    │  规划Agent      │    │ 评估Agent   │ │
│  │ (Strategy)      │◄───►│ (Planner)      │◄───►│(Evaluator) │ │
│  └───────┬─────────┘    └────────┬────────┘    └─────┬──────┘ │
└──────────┼──────────────────────┼────────────────────┼────────┘
           │                      │                    │
           ▼                      ▼                    ▼
┌──────────────────────────────────────────────────────────────┐
│                    执行层 (Execution Layer)                   │
│                                                              │
│  ┌────────────────┐   ┌────────────────┐   ┌───────────────┐ │
│  │  世界观Agent   │   │   角色Agent     │   │  剧情Agent    │ │
│  │  (Worldview)   │   │  (Character)   │   │   (Plot)      │ │
│  └────────────────┘   └────────────────┘   └───────────────┘ │
│                                                              │
│  ┌────────────────┐   ┌────────────────┐   ┌───────────────┐ │
│  │   对话Agent    │   │  背景Agent     │   │ JSON格式化Agent│ │
│  │  (Dialogue)    │   │ (Background)   │   │  (Formatter)  │ │
│  └────────────────┘   └────────────────┘   └───────────────┘ │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────┐
│                     共享资源 (Shared Resources)               │
│                                                              │
│   ┌─────────────────┐    ┌─────────────────┐   ┌──────────┐  │
│   │   记忆管理器     │    │    模型接口      │   │ 工具调用器│  │
│   │ (Memory Manager)│    │(Model Interface)│   │(Tool Reg)│  │
│   └─────────────────┘    └─────────────────┘   └──────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## 2. 角色与职责

### 2.1 决策层

决策层负责高层次决策和规划，主要包括以下智能体：

#### 策略智能体 (Strategy Agent)
- **职责**：制定创作策略，确定小说主题、风格和整体方向
- **输入**：用户需求、创作目标
- **输出**：创作策略、高层次指令

#### 规划智能体 (Planner Agent)
- **职责**：将策略转化为具体计划，设计小说结构
- **输入**：创作策略
- **输出**：详细的创作计划，包括世界观要素、角色设计指南、情节架构等

#### 评估智能体 (Evaluator Agent)
- **职责**：评估生成内容的质量、一致性和逻辑性
- **输入**：执行层生成的内容
- **输出**：评估报告、修改建议

### 2.2 执行层

执行层负责具体内容生成，按照决策层的规划执行细节工作：

#### 世界观智能体 (Worldview Agent)
- **职责**：生成小说世界观细节
- **输入**：规划智能体提供的世界观框架
- **输出**：详细的世界设定，包括历史、地理、文化等

#### 角色智能体 (Character Agent)
- **职责**：设计小说角色
- **输入**：规划智能体提供的角色框架、世界观
- **输出**：角色详情，包括背景、性格、动机、关系网络等

#### 剧情智能体 (Plot Agent)
- **职责**：生成小说情节
- **输入**：规划智能体提供的情节框架、世界观、角色
- **输出**：故事情节，包括冲突、发展、高潮和结局

#### 对话智能体 (Dialogue Agent)
- **职责**：生成角色对话
- **输入**：角色信息、情节上下文
- **输出**：自然、符合角色特性的对话内容

#### 背景智能体 (Background Agent)
- **职责**：生成场景描述和背景细节
- **输入**：世界观信息、情节需求
- **输出**：环境描述、氛围营造等内容

#### JSON格式化智能体 (Formatter Agent)
- **职责**：将生成内容转换为结构化JSON格式
- **输入**：各智能体生成的内容
- **输出**：符合特定模式的JSON数据

### 2.3 共享资源

#### 记忆管理器 (Memory Manager)
- **职责**：存储和管理智能体间的共享信息
- **功能**：记忆保存、检索、更新
- **特点**：支持多维度查询和关联分析

#### 模型接口 (Model Interface)
- **职责**：统一管理与底层LLM模型的交互
- **功能**：提示词构建、模型调用、响应处理
- **特点**：基于LangChain Go库，支持多种模型后端（Ollama、DeepSeek-api等）

#### 工具调用系统 (Tool Calling System)
- **职责**：提供智能体调用外部功能的能力
- **功能**：工具注册、调用处理、结果返回
- **特点**：基于LangChain Go库，支持自定义工具和第三方工具

## 3. 工作流程

### 3.1 基本工作流

1. **需求收集**：系统接收用户需求（例如小说主题、风格等）
2. **策略制定**：策略智能体基于需求制定高层次创作策略
3. **计划设计**：规划智能体将策略转化为详细计划
4. **内容生成**：执行层智能体根据计划生成具体内容
   - 世界观智能体生成世界设定
   - 角色智能体生成角色细节
   - 剧情智能体生成故事情节
   - 对话智能体生成角色对话
   - 背景智能体生成环境描述
5. **内容格式化**：格式化智能体将内容转换为结构化JSON
6. **质量评估**：评估智能体评价生成内容的质量
7. **迭代优化**：根据评估结果进行必要的调整和优化
8. **最终输出**：生成最终的小说内容

### 3.2 智能体间通信

智能体间通信采用统一的消息接口：

- **消息结构**：类型、内容、数据、元数据
- **消息传递**：通过编排器协调，或直接点对点传递
- **状态共享**：通过记忆管理器实现持久状态共享

## 4. 技术实现计划

### 4.1 核心接口设计

- **Agent接口**：定义所有智能体必须实现的基本方法
- **消息结构**：设计统一的消息格式和处理流程
- **记忆接口**：设计高效的记忆存储和检索机制

### 4.2 实现阶段

1. **基础架构阶段**
   - 实现基础接口和组件
   - 构建记忆管理系统
   - 开发模型接口层

2. **智能体开发阶段**
   - 实现决策层智能体
   - 实现执行层智能体
   - 开发通信机制

3. **集成测试阶段**
   - 系统集成测试
   - 生成质量评估
   - 性能优化

4. **生产部署阶段**
   - 与现有系统集成
   - 监控和日志系统
   - 用户反馈机制

### 4.3 与现有系统集成

- 集成现有的世界观生成能力
- 复用背景信息管理系统
- 利用现有的规则生成功能

## 5. 工具调用系统设计

### 5.1 系统概述

工具调用系统为NovelAI多层代理架构提供了访问外部功能的能力，使智能体能够查询信息、执行操作并与用户或其他系统交互。该系统基于 LangChain Go 库（langchaingo）实现，充分利用其成熟的工具生态系统和扩展能力。

### 5.2 核心组件

#### 5.2.1 工具适配器 (Tool Adapter)

工具适配器将 langchaingo 的工具接口与我们的代理系统无缝集成：

```go
// shared/tools/adapter.go

package tools

import (
    "context"
    
    "github.com/tmc/langchaingo/tools"
)

// LangChainAdapter 将 langchaingo 工具适配到我们的系统
// 实现了无缝集成第三方工具
// 同时保持我们的系统设计不受影响
// 这是适配器模式的典型应用
```

#### 5.2.2 工具注册表 (Tool Registry)

工具注册表管理系统中所有可用的工具，实现集中式管理：

```go
// shared/tools/registry.go

package tools

import (
    "github.com/tmc/langchaingo/tools"
)

// ToolRegistry 管理系统中所有可用的工具
// 提供注册、获取和列举功能
```

#### 5.2.3 工具调用器 (Tool Caller)

工具调用器负责执行工具调用请求：

```go
// shared/tools/caller.go

package tools

import (
    "context"
    
    "github.com/tmc/langchaingo/tools"
)

// ToolCaller 处理工具调用请求
// 负责解析请求、调用相应工具并返回结果
```

### 5.3 预定义工具

系统提供以下预定义工具，用于小说生成场景：

#### 5.3.1 世界观工具 (Worldview Tools)

```go
// shared/tools/worldtools/worldview.go

package worldtools

import (
    "context"
    "encoding/json"
    
    "github.com/tmc/langchaingo/tools"
)

// WorldviewSearchTool 实现世界观搜索功能
// 允许智能体查询现有世界观

// WorldviewCreateTool 实现世界观创建功能
// 允许智能体创建新的世界观设定
```

#### 5.3.2 角色工具 (Character Tools)

```go
// shared/tools/worldtools/character.go

package worldtools

import (
    "context"
    "encoding/json"
    
    "github.com/tmc/langchaingo/tools"
)

// CharacterSearchTool 实现角色搜索功能
// 允许智能体查询现有角色

// CharacterCreateTool 实现角色创建功能
// 允许智能体创建新的角色
```

#### 5.3.3 剧情工具 (Plot Tools)

```go
// shared/tools/worldtools/plot.go

package worldtools

import (
    "context"
    "encoding/json"
    
    "github.com/tmc/langchaingo/tools"
)

// PlotGeneratorTool 实现剧情生成功能
// 基于世界观和角色生成剧情框架

// PlotAnalysisTool 实现剧情分析功能
// 分析现有剧情的结构和元素
```

### 5.4 与Agent集成

通过扩展 Agent 接口，使智能体能够使用工具调用功能：

```go
// core/agent.go

package core

import (
    "context"
    
    "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/shared/tools"
)

// Agent 定义所有智能体必须实现的基本接口
// 增加了工具调用相关方法
```

### 5.5 工具调用流程

1. 智能体生成包含工具调用请求的消息
2. 编排器识别工具调用消息并路由到工具调用器
3. 工具调用器执行工具调用并返回结果
4. 编排器将结果路由回原始智能体
5. 智能体处理工具调用结果并生成响应

### 5.6 集成示例

下面是一个集成示例，展示如何在规划智能体中使用工具调用功能：

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/tools"
    
    "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/core"
    "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/decision"
    "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/shared/model"
    customtools "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/shared/tools"
    "github.com/yourusername/novelai/pkg/experimental/multilayer_agent/shared/tools/worldtools"
)

func main() {
    // 注册工具
    registry := customtools.NewToolRegistry()
    worldviewSearchTool := worldtools.NewWorldviewSearchTool()
    registry.RegisterTool(worldviewSearchTool)
    
    // 创建工具调用器
    toolCaller := customtools.NewToolCaller(registry)
    
    // 创建规划智能体
    plannerAgent := decision.NewPlannerAgent(model.NewOllamaModel("http://localhost:11434"))
    plannerAgent.SetToolCaller(toolCaller)
    
    // 使用智能体
    // ...
}
```

## 6. 扩展性考虑

### 6.1 水平扩展

- **新智能体类型**：可通过实现Agent接口添加新的专业智能体
- **调整层次结构**：可根据需要调整决策层和执行层的职责划分

### 6.2 垂直扩展

- **模型替换**：支持更换底层生成模型
- **算法优化**：可优化各智能体内部算法

## 7. 预期优势

1. **专业分工**：各智能体专注于特定任务，提高生成质量
2. **决策与执行分离**：高层决策与具体执行解耦，便于控制和调整
3. **一致性保证**：通过记忆管理和评估机制确保内容一致性
4. **可扩展性**：模块化设计便于功能扩展和系统演进
5. **可维护性**：清晰的责任边界便于调试和改进

## 8. 后续工作

1. **智能体能力优化**：提升各专业智能体的生成能力
2. **记忆管理高级功能**：实现更复杂的记忆关联和检索
3. **评估机制完善**：开发更全面的质量评估标准
4. **用户交互优化**：提供更灵活的用户干预机制
5. **工具调用能力扩展**：增强工具生态系统，支持更多带有智能体业务逻辑的复杂工具

## 9. 结语

本设计采用分层架构将小说生成过程分解为决策和执行两个主要层次，通过专业智能体协同工作提高生成内容的质量和一致性。系统具有良好的可扩展性和可维护性，能够有效集成到现有的NovelAI项目中。

通过集成 LangChain Go 库，系统获得了强大的工具调用能力，使智能体能够与外部系统交互，获取信息并执行操作。这种集成方式不仅保持了系统的灵活性和可扩展性，还能使用行业标准工具库，减少开发成本并提高代码质量。
