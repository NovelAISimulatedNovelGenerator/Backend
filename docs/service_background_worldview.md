# `WorldviewService` - 背景世界观服务层文档

## 1. 目标 (Goal)

`WorldviewService` (`biz/service/background/worldview_service.go`) 是背景模块中负责处理“世界观 (Worldview)” 实体相关业务逻辑的服务层。它作为 Handler (API 层) 和 DAL (数据访问层) 之间的桥梁，封装了对世界观数据进行增、删、改、查、列 (CRUDL) 操作的核心逻辑。

## 2. 服务结构 (Structure)

### 2.1. 结构体定义

```go
// WorldviewService 负责处理世界观相关的业务逻辑
type WorldviewService struct {
	ctx context.Context        // 请求级别的上下文
	c   *app.RequestContext  // Hertz 请求上下文
}
```

- `ctx`: 标准库的 `context.Context`，用于传递请求范围内的截止日期、取消信号以及其他值。
- `c`: Hertz 框架的 `app.RequestContext`，提供了访问请求和响应的便捷方法。

### 2.2. 构造函数

```go
// NewWorldviewService 创建 WorldviewService 实例
func NewWorldviewService(ctx context.Context, c *app.RequestContext) *WorldviewService
```

通过此函数创建 `WorldviewService` 的新实例，传入必要的上下文信息。

## 3. 核心功能 (Core Functions)

服务层暴露了以下核心方法来操作世界观数据：

### 3.1. `CreateWorldview`

- **目的**: 创建一个新的世界观记录。
- **参数**: `req *background.CreateWorldviewRequest` - 包含新世界观的名称、描述、标签和父 ID。
- **返回**: `*background.Worldview`, `error` - 返回创建成功后的世界观信息 (API 模型) 和潜在的错误。
- **逻辑**: 
    - 验证请求参数 (非空检查)。
    - 构建 `db.Worldview` 数据库模型。
    - 调用 `db.CreateWorldview` 将数据存入数据库。
    - 处理并记录 DAL 返回的错误。
    - 将创建成功的数据库模型转换为 API 模型返回。

### 3.2. `GetWorldviewByID`

- **目的**: 根据提供的 ID 获取单个世界观的详细信息。
- **参数**: `req *background.GetWorldviewRequest` - 包含要查询的世界观 ID。
- **返回**: `*background.Worldview`, `error` - 返回查询到的世界观信息 (API 模型) 和潜在的错误。
- **逻辑**: 
    - 验证请求参数 (ID 是否有效)。
    - 调用 `db.GetWorldviewByID` 从数据库查询数据。
    - 处理特定的 `db.ErrWorldviewNotFound` 错误。
    - 记录并处理其他 DAL 错误。
    - 调用 `convertDBWorldviewToModel` 将数据库模型转换为 API 模型返回。

### 3.3. `UpdateWorldview`

- **目的**: 更新一个已存在的世界观记录。
- **参数**: `req *background.UpdateWorldviewRequest` - 包含要更新的世界观 ID 以及需要更新的字段 (名称、描述、标签、父 ID)。
- **返回**: `error` - 返回操作中可能发生的错误。
- **逻辑**: 
    - 验证请求参数 (ID 是否有效)。
    - 检查是否有任何字段需要更新，如果请求中所有可选字段都为空/零值，则直接返回。
    - 构建一个 `map[string]interface{}`，只包含请求中明确提供的非零值/非空字段，用于部分更新。
    - 调用 `db.UpdateWorldview` 执行数据库更新。
    - 记录并处理 DAL 返回的错误。

### 3.4. `DeleteWorldview`

- **目的**: 根据 ID 删除一个世界观记录。
- **参数**: `req *background.DeleteWorldviewRequest` - 包含要删除的世界观 ID (`WorldviewId`)。
- **返回**: `error` - 返回操作中可能发生的错误。
- **逻辑**: 
    - 验证请求参数 (ID 是否有效)。
    - 调用 `db.DeleteWorldview` 执行数据库删除。
    - 处理特定的 `db.ErrWorldviewNotFound` 错误。
    - 记录并处理其他 DAL 错误。

### 3.5. `ListWorldviews`

- **目的**: 获取世界观列表，支持分页和过滤。
- **参数**: `req *background.ListWorldviewsRequest` - 包含分页参数 (页码 `page`，每页大小 `page_size`) 和过滤参数 (标签 `tag_filter`，父 ID `parent_id_filter`)。
- **返回**: `[]*background.Worldview`, `int64`, `error` - 返回查询到的世界观列表 (API 模型)、满足条件的总记录数以及潜在的错误。
- **逻辑**: 
    - 验证请求参数 (非空检查)。
    - 设置默认分页参数 (page=1, pageSize=10)。
    - 解析过滤参数:
        - `parentID`: 根据 IDL 定义，使用 `-1` 传递给 DAL 表示不按父 ID 过滤 (需要 DAL 实现支持)。
        - `tagFilter`: 从 `req.GetTagFilter()` 获取。
    - 调用 `db.ListWorldviews` 从数据库获取列表和总数。
    - 记录并处理 DAL 返回的错误。
    - 遍历查询结果，使用 `convertDBWorldviewToModel` 将每个数据库模型转换为 API 模型。
    - 返回转换后的列表、总数和 nil 错误。

## 4. 辅助函数 (Helper Functions)

### 4.1. `convertDBWorldviewToModel`

- **目的**: 将 DAL 层返回的 `*db.Worldview` 结构 (数据库模型) 转换为 API 层使用的 `*background.Worldview` 结构 (Protobuf 模型)。
- **注意**: 由于最新的 IDL 更改，`CreatedAt` 和 `UpdatedAt` 字段在数据库模型和 Protobuf 模型中都已是 `int64` (Unix 时间戳)，因此该函数现在直接进行赋值，不再需要 `timestamppb` 的转换。

## 5. 依赖与交互 (Dependencies & Interactions)

- **数据访问层 (DAL)**: 通过 `novelai/biz/dal/db` 包与数据库交互，调用如 `db.CreateWorldview`, `db.GetWorldviewByID` 等函数。
- **API 数据模型**: 使用 `novelai/biz/model/background` 包中由 Protobuf 定义生成的 Go 结构体 (如 `background.Worldview`, `background.CreateWorldviewRequest` 等) 作为函数的参数和返回值类型。
- **日志库**: 使用 Hertz 框架提供的 `hlog` 进行日志记录，特别是带上下文的 `hlog.Ctx*` 函数。

## 6. 错误处理 (Error Handling)

- 服务层会捕获并记录来自 DAL 层的错误。
- 对于特定的错误 (如 `db.ErrWorldviewNotFound`)，会进行特殊处理或记录警告日志。
- 通常会将原始错误或包装后的错误向上传递给 Handler 层。
- 输入验证错误通常直接返回，并记录错误日志。

## 7. 日志记录 (Logging)

- 使用 `hlog.CtxInfof`, `hlog.CtxWarnf`, `hlog.CtxErrorf` 进行日志记录。
- 关键操作的开始、成功和失败都会有相应的日志输出。
- 日志中会包含关键信息，如操作的 ID、错误详情等，以便于追踪和调试。
