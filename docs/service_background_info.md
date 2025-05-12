# Service Layer: BackgroundInfoService

路径: `biz/service/background/background_info_service.go`

## 概述

`BackgroundInfoService` 负责处理与背景信息 (`BackgroundInfo`) 相关的业务逻辑。它作为 API Handler 层和 Data Access Layer (DAL) 之间的桥梁，封装了创建、读取、更新和删除 (CRUD) 背景信息的操作，以及列出背景信息的功能。

该服务旨在确保业务规则得到执行，处理输入验证，并调用相应的 DAL 函数来与数据库交互。

## 主要职责

-   **封装业务逻辑:** 将背景信息相关的操作（如验证、数据转换）集中处理。
-   **数据校验:** 对来自 Handler 的请求参数进行基础校验。
-   **DAL 交互:** 调用 `biz/dal/db/background_info.go` 中的函数执行数据库操作。
-   **数据转换:** 将 DAL 返回的数据库模型 (`db.BackgroundInfo`) 转换为 Service 层或 API 层所需的模型 (`background.BackgroundInfo`)。
-   **错误处理:** 捕获并记录来自 DAL 的错误，并根据需要将错误传递给上层。
-   **日志记录:** 使用 `hlog` 记录关键操作和错误信息。

## 结构体

```go
type BackgroundInfoService struct {
    ctx context.Context
    req *app.RequestContext
}
```

-   `ctx`: Golang 标准库的上下文，用于传递请求范围的数据和控制信号（如超时、取消）。
-   `req`: Hertz 框架的请求上下文，包含 HTTP 请求相关信息。

## 构造函数

```go
func NewBackgroundInfoService(ctx context.Context, req *app.RequestContext) *BackgroundInfoService
```

创建一个新的 `BackgroundInfoService` 实例。

## 主要方法

### `CreateBackgroundInfo`

-   **签名:** `func (s *BackgroundInfoService) CreateBackgroundInfo(req *background.CreateBackgroundInfoRequest) (*background.BackgroundInfo, error)`
-   **功能:** 创建一个新的背景信息。
-   **流程:**
    1.  校验 `req` 和 `req.BackgroundInfo` 是否为 `nil`。
    2.  将 `req.BackgroundInfo` (服务层模型) 转换为 `db.BackgroundInfo` (DAL 模型)。
    3.  调用 `db.CreateBackgroundInfo` 将数据存入数据库。
    4.  如果创建成功，调用 `db.GetBackgroundInfoByID` 获取完整的、包含生成 ID 和时间戳的背景信息。
    5.  将获取到的 `db.BackgroundInfo` 转换为 `background.BackgroundInfo` 并返回。
    6.  记录错误或成功日志。

### `GetBackgroundInfoByID`

-   **签名:** `func (s *BackgroundInfoService) GetBackgroundInfoByID(req *background.GetBackgroundInfoRequest) (*background.BackgroundInfo, error)`
-   **功能:** 根据 ID 获取单个背景信息。
-   **流程:**
    1.  校验 `req.Id` 是否有效 (> 0)。
    2.  调用 `db.GetBackgroundInfoByID` 从数据库查询。
    3.  如果找到，将 `db.BackgroundInfo` 转换为 `background.BackgroundInfo` 并返回。
    4.  处理 `ErrBackgroundInfoNotFound` 等错误。
    5.  记录错误或成功日志。

### `UpdateBackgroundInfo`

-   **签名:** `func (s *BackgroundInfoService) UpdateBackgroundInfo(req *background.UpdateBackgroundInfoRequest) (*background.BackgroundInfo, error)`
-   **功能:** 更新指定 ID 的背景信息。
-   **流程:**
    1.  校验 `req.Id` 和 `req.UpdateFields` 是否有效。
    2.  根据 `req.UpdateFields` (FieldMask) 构造一个 `map[string]interface{}`，只包含需要更新的字段。
    3.  添加对允许更新字段的检查 (例如，不允许直接更新 `worldview_id`)。
    4.  调用 `db.UpdateBackgroundInfo` 执行更新。
    5.  如果更新成功，调用 `db.GetBackgroundInfoByID` 获取更新后的完整信息。
    6.  将获取到的 `db.BackgroundInfo` 转换为 `background.BackgroundInfo` 并返回。
    7.  处理 `ErrBackgroundInfoNotFound` 等错误。
    8.  记录错误或成功日志。

### `DeleteBackgroundInfo`

-   **签名:** `func (s *BackgroundInfoService) DeleteBackgroundInfo(req *background.DeleteBackgroundInfoRequest) error`
-   **功能:** 根据 ID 删除背景信息。
-   **流程:**
    1.  校验 `req.Id` 是否有效 (> 0)。
    2.  调用 `db.DeleteBackgroundInfo` 执行删除操作 (硬删除)。
    3.  处理 `ErrBackgroundInfoNotFound` 等错误。
    4.  记录错误或成功日志。

### `ListBackgroundInfos`

-   **签名:** `func (s *BackgroundInfoService) ListBackgroundInfos(req *background.ListBackgroundInfosRequest) ([]*background.BackgroundInfo, int64, error)`
-   **功能:** 列出背景信息，支持过滤和分页。
-   **流程:**
    1.  校验 `req`。
    2.  处理分页参数 `Page` 和 `PageSize`，设置默认值。
    3.  处理过滤参数 `WorldviewIdFilter`, `ParentIdFilter`, `TagFilter`。
        -   特别注意 `ParentIdFilter`：如果请求中未设置该字段 (`IsSetParentIdFilter` 为 false)，则传递 `-1` 给 DAL 以表示不根据 `parent_id` 筛选。
    4.  调用 `db.ListBackgroundInfos` 从数据库查询列表和总数。
    5.  将返回的 `[]db.BackgroundInfo` 列表逐个转换为 `[]*background.BackgroundInfo`。
    6.  返回转换后的列表、总数和错误信息。
    7.  记录错误或成功日志。

## 辅助函数

### `convertDBBackgroundInfoToModel`

-   **签名:** `func convertDBBackgroundInfoToModel(dbBI *db.BackgroundInfo) *background.BackgroundInfo`
-   **功能:** 将 DAL 的 `db.BackgroundInfo` 结构体转换为 Service 层的 `background.BackgroundInfo` 结构体。

## 依赖

-   `context`, `errors`
-   `github.com/cloudwego/hertz/pkg/app`
-   `github.com/cloudwego/hertz/pkg/common/hlog`
-   `gorm.io/gorm`
-   `novelai/biz/dal/db`
-   `novelai/biz/model/background`
-   `novelai/pkg/constants`
-   `novelai/pkg/utils`

## 注意事项

-   错误处理：服务层通常会记录详细错误，但可能只向上层返回通用错误或 DAL 定义的特定错误 (如 `ErrBackgroundInfoNotFound`)。
-   验证：当前只做了基础的 `nil` 或 ID 检查，可以根据业务需求增加更复杂的验证逻辑 (例如，检查 `ParentID` 是否存在)。
-   模型转换：确保 `convertDBBackgroundInfoToModel` 函数正确映射所有需要的字段。
-   `UpdateBackgroundInfo` 中的 `FieldMask` 处理需要仔细实现，以确保只更新用户明确指定的字段。
