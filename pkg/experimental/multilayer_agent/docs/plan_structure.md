pkg/experimental/multilayer_agent/
├── core/
│   ├── agent.go        // 基础智能体接口定义
│   ├── message.go      // 消息结构定义
│   └── orchestrator.go // 智能体编排器
├── decision/
│   ├── strategy.go     // 策略智能体
│   ├── planner.go      // 规划智能体
│   └── evaluator.go    // 评估智能体
├── execution/
│   ├── worldview.go    // 世界观智能体
│   ├── character.go    // 角色智能体
│   ├── plot.go         // 剧情智能体
│   ├── dialogue.go     // 对话智能体
│   ├── background.go   // 背景智能体
│   └── formatter.go    // JSON格式化智能体
├── shared/
│   ├── memory/         // 记忆管理系统
│   │   ├── manager.go
│   │   └── store.go
│   ├── model/          // 模型接口
│   │   ├── interface.go
│   │   ├── ollama.go
│   │   └── deepseek-api.go
│   └── tools/          // 工具调用系统
│       ├── registry.go // 工具注册表
│       ├── adapter.go  // langchaingo适配器
│       ├── caller.go   // 工具调用处理器
│       └── worldtools/ // 特定领域工具
│           ├── worldview.go
│           ├── character.go
│           └── plot.go
└── workflow/
    └── engine.go       // 工作流引擎