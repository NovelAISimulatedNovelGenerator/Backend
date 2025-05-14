# `RuleService` - 背景规则服务层文档

## 1. 目标 (Goal)

`RuleService` (`biz/service/background/rule_service.go`) 是背景模块中负责处理"规则 (Rule)" 实体相关业务逻辑的服务层。它作为 Handler (API 层) 和 DAL (数据访问层) 之间的桥梁，封装了对规则数据进行增、删、改、查、列 (CRUDL) 操作的核心逻辑。

## 2. 服务结构 (Structure)

### 2.1. 结构体定义

```go
// RuleService 用于管理规则相关的业务逻辑
type RuleService struct {
    ctx context.Context     // 当前上下文
    app *app.RequestContext // Hertz 的请求上下文
}
```

- `ctx`: 标准库的 `context.Context`，用于传递请求范围内的截止日期、取消信号以及其他值。
- `app`: Hertz 框架的 `app.RequestContext`，提供了访问请求和响应的便捷方法。

### 2.2. 构造函数

```go
// NewRuleService 创建 RuleService 实例
func NewRuleService(ctx context.Context, appCtx *app.RequestContext) *RuleService
```

通过此函数创建 `RuleService` 的新实例，传入必要的上下文信息。

## 3. 核心功能 (Core Functions)

RuleService 提供对 Rule 实体的完整 CRUD 操作：

### 3.1. CreateRule - 创建新规则

```go
// CreateRule 创建新的规则
func (s *RuleService) CreateRule(req *background.CreateRuleRequest) (*background.Rule, error)
```

- **功能**: 创建一个新的规则记录。
- **参数**: `req` 包含创建新规则所需的全部信息，如 `WorldviewId`、`Name`、`Description`、`Tag` 和 `ParentId`。
- **返回**: 成功返回创建的规则对象；失败返回错误。
- **错误处理**:
  - 如果请求为 nil 或参数无效 (如 `WorldviewId` <= 0)，返回 `errno.InvalidParameterError`。
  - 如果 `Name` 为空，返回 `errno.InvalidParameterError`。
  - 创建过程中出现数据库错误，返回 `errno.DatabaseError`。

### 3.2. GetRuleByID - 获取规则信息

```go
// GetRuleByID 根据 ID 获取规则信息
func (s *RuleService) GetRuleByID(req *background.GetRuleRequest) (*background.Rule, error)
```

- **功能**: 根据 ID 查询单个规则。
- **参数**: `req` 包含要查询的规则 ID (`RuleId`)。
- **返回**: 如果找到，返回规则对象；如果未找到或出错，返回错误。
- **错误处理**:
  - 如果 ID 无效 (<=0)，返回 `errno.InvalidParameterError`。
  - 如果规则不存在，返回 `errno.NotFoundError("Rule")`。
  - 其他数据库错误，返回 `errno.DatabaseError`。

### 3.3. UpdateRule - 更新规则

```go
// UpdateRule 更新规则信息
func (s *RuleService) UpdateRule(req *background.UpdateRuleRequest) (*background.Rule, error)
```

- **功能**: 根据请求更新现有规则记录。
- **参数**: `req` 包含规则 ID 及需要更新的字段。
- **返回**: 成功返回更新后的规则对象；失败返回错误。
- **错误处理**:
  - 如果 ID 无效，返回 `errno.InvalidParameterError`。
  - 如果规则不存在，返回 `errno.NotFoundError("Rule")`。
  - 如果更新过程中出现数据库错误，返回 `errno.DatabaseError`。
- **注意**:
  - 只有非空字段会被更新
  - `ParentID` 是特例，总是被更新，包括为 0 的情况 (表示顶级规则)
  - 如果没有字段需要更新，会重新获取并返回当前规则

### 3.4. DeleteRule - 删除规则

```go
// DeleteRule 删除规则
func (s *RuleService) DeleteRule(req *background.DeleteRuleRequest) error
```

- **功能**: 根据 ID 删除规则。
- **参数**: `req` 包含要删除的规则 ID (`RuleId`)。
- **返回**: 成功返回 nil；失败返回错误。
- **错误处理**:
  - 如果 ID 无效，返回 `errno.InvalidParameterError`。
  - 如果规则不存在，返回 `errno.NotFoundError("Rule")`。
  - 如果删除过程中出现数据库错误，返回 `errno.DatabaseError`。

### 3.5. ListRules - 列出规则

```go
// ListRules 列出规则，支持分页和过滤
func (s *RuleService) ListRules(req *background.ListRulesRequest) ([]*background.Rule, int64, error)
```

- **功能**: 根据筛选条件获取规则列表，支持分页。
- **参数**: `req` 包含筛选条件和分页参数。
  - `WorldviewIdFilter`: 按世界观 ID 筛选
  - `ParentIdFilter`: 按父规则 ID 筛选 (-1 表示不筛选，0 表示筛选顶级规则)
  - `TagFilter`: 按标签筛选
  - `Page` 和 `PageSize`: 分页参数
- **返回**: 规则列表、总记录数，出错时返回错误。
- **默认值**:
  - 如果 `Page` <= 0，默认为 1
  - 如果 `PageSize` <= 0，默认为 10

## 4. 辅助功能 (Helper Functions)

### 4.1. convertDBRuleToModel - 数据模型转换

```go
// convertDBRuleToModel 将 DAL 层 Rule 结构转换为 API 模型结构
func convertDBRuleToModel(dbRule *db.Rule) *background.Rule
```

- **功能**: 将 DAL 层返回的 `*db.Rule` 结构 (数据库模型) 转换为 API 层使用的 `*background.Rule` 结构 (Protobuf 模型)。
- **参数**: `dbRule` 数据库规则对象
- **返回**: API 层规则对象

## 5. 依赖与交互 (Dependencies & Interactions)

### 5.1. 内部依赖

- **biz/dal/db**: 提供数据访问层的 Rule 相关操作 (`CreateRule`, `GetRuleByID`, `UpdateRule`, `DeleteRule`, `ListRules`)。
- **biz/model/background**: 包含 Rule 相关的 Protobuf 生成结构体和请求/响应定义。
- **pkg/errno**: 提供统一的错误处理机制。

### 5.2. 错误处理

RuleService 对 DAL 层返回的错误进行统一转换:
- 参数验证失败: 返回 `errno.InvalidParameterError`
- 规则不存在: 返回 `errno.NotFoundError("Rule")`
- 数据库错误: 返回 `errno.DatabaseError`

### 5.3. 日志记录

使用 Hertz 框架提供的 `hlog` 包记录关键操作和错误:
- 请求开始: 使用 `hlog.CtxInfof` 记录操作开始
- 警告: 使用 `hlog.CtxWarnf` 记录非致命问题
- 错误: 使用 `hlog.CtxErrorf` 记录操作失败

## 6. 使用示例 (Usage Examples)

### 6.1. 创建规则

```go
// 创建规则服务实例
ruleService := background.NewRuleService(ctx, c)

// 准备创建规则请求
req := &background.CreateRuleRequest{
    WorldviewId: 1,
    Name: "魔法规则",
    Description: "描述魔法系统的规则",
    Tag: "magic,system",
    ParentId: 0, // 作为顶级规则
}

// 调用创建方法
rule, err := ruleService.CreateRule(req)
if err != nil {
    // 错误处理
    return err
}

// 使用创建的规则
fmt.Printf("Created rule with ID: %d\n", rule.Id)
```

### 6.2. 列出某世界观下的规则

```go
// 创建规则服务实例
ruleService := background.NewRuleService(ctx, c)

// 准备列出规则请求
req := &background.ListRulesRequest{
    WorldviewIdFilter: 1,  // 筛选世界观 ID 为 1 的规则
    ParentIdFilter: 0,     // 只列出顶级规则
    Page: 1,
    PageSize: 10,
}

// 调用列出方法
rules, total, err := ruleService.ListRules(req)
if err != nil {
    // 错误处理
    return err
}

// 处理结果
fmt.Printf("Found %d rules, total %d\n", len(rules), total)
for _, rule := range rules {
    fmt.Printf("Rule ID: %d, Name: %s\n", rule.Id, rule.Name)
}
```

## 7. 未来改进计划 (Future Improvements)

1. 增加批量操作支持，例如批量创建规则、批量删除等。
2. 改进搜索功能，支持文本内容的全文检索。
3. 实现规则层次结构的完整递归查询，如获取包含所有子规则的树状结构。
4. 添加规则版本控制，支持规则的历史版本管理。
