# Background 模块 DAL (Data Access Layer) 文档

## 1. 概述

`background` 模块的 DAL 层负责处理与故事背景设定相关的核心实体的数据持久化和检索。它直接与数据库交互，为服务层提供统一的数据访问接口。该层主要管理以下三个实体：

*   `Worldview` (世界观)
*   `Rule` (规则设定)
*   `BackgroundInfo` (背景信息)

此文档遵循文档驱动开发 (Documentation-Driven Development) 的原则，旨在提供清晰、准确的 DAL 层参考。

## 2. 数据模型 (GORM Structs)

数据模型使用 GORM (Go Object Relational Mapper) 定义，并存储在 `biz/dal/db/` 目录下。

### 2.1. Worldview (世界观)

*   **定义文件**: `biz/dal/db/worldview.go`
*   **表名常量**: `constants.TableNameWorldview`
*   **描述**: 世界观是故事的基础框架，可以有层级关系（通过 `ParentID` 实现）。
*   **主要字段**:
    *   `ID` (int64): 主键，自增ID。
    *   `Name` (string): 世界观名称。
    *   `Description` (string): 世界观的详细描述。
    *   `Tag` (string): 标签，多个标签用英文逗号分隔，用于分类和搜索。
    *   `ParentID` (int64): 父世界观ID，0表示顶级世界观。
    *   `CreatedAt` (int64): 创建时间 (Unix时间戳秒级)，由 GORM 自动管理。
    *   `UpdatedAt` (int64): 更新时间 (Unix时间戳秒级)，由 GORM 自动管理。

### 2.2. Rule (规则设定)

*   **定义文件**: `biz/dal/db/rule.go`
*   **表名常量**: `constants.TableNameRule`
*   **描述**: 规则设定是世界观下的具体行为准则、物理法则或社会规范等。每个规则都必须属于一个世界观，并可以有层级关系。
*   **主要字段**:
    *   `ID` (int64): 主键，自增ID。
    *   `WorldviewID` (int64): 外键，关联到 `Worldview` 表的 `ID`，表示此规则所属的世界观。
    *   `Name` (string): 规则名称。
    *   `Description` (string): 规则的详细描述。
    *   `Tag` (string): 标签，多个标签用英文逗号分隔。
    *   `ParentID` (int64): 父规则ID，0表示顶级规则（在所属世界观内）。
    *   `CreatedAt` (int64): 创建时间 (Unix时间戳秒级)，由 GORM 自动管理。
    *   `UpdatedAt` (int64): 更新时间 (Unix时间戳秒级)，由 GORM 自动管理。

### 2.3. BackgroundInfo (背景信息)

*   **定义文件**: `biz/dal/db/background_info.go`
*   **表名常量**: `constants.TableNameBackgroundInfo`
*   **描述**: 背景信息是构成世界观的具体元素，如地点、角色背景、历史事件等。每个背景信息必须属于一个世界观，并可以有层级关系。
*   **主要字段**:
    *   `ID` (int64): 主键，自增ID。
    *   `WorldviewID` (int64): 外键，关联到 `Worldview` 表的 `ID`，表示此背景信息所属的世界观。
    *   `Name` (string): 背景信息名称。
    *   `Description` (string): 背景信息的详细描述。
    *   `Tag` (string): 标签，多个标签用英文逗号分隔。
    *   `ParentID` (int64): 父背景信息ID，0表示顶级背景信息（在所属世界观内）。
    *   `CreatedAt` (int64): 创建时间 (Unix时间戳秒级)，由 GORM 自动管理。
    *   `UpdatedAt` (int64): 更新时间 (Unix时间戳秒级)，由 GORM 自动管理。

## 3. 核心函数

每个实体都提供了一套标准的 CRUD (Create, Read, Update, Delete) 操作以及列表查询功能。

### 3.1. Worldview 函数

*   `CreateWorldview(ctx context.Context, wv *Worldview) (int64, error)`: 创建新的世界观，返回新创建的世界观ID。
*   `GetWorldviewByID(ctx context.Context, id int64) (*Worldview, error)`: 根据ID查询世界观信息。
*   `UpdateWorldview(ctx context.Context, id int64, updates map[string]interface{}) error`: 更新指定ID的世界观信息。
*   `DeleteWorldview(ctx context.Context, id int64) error`: 删除指定ID的世界观（硬删除）。
*   `ListWorldviews(ctx context.Context, parentIDFilter int64, tagFilter string, page, pageSize int) ([]Worldview, int64, error)`: 列出世界观，支持按父ID和标签过滤，并进行分页。返回世界观列表、总记录数和错误信息。

### 3.2. Rule 函数

*   `CreateRule(ctx context.Context, r *Rule) (int64, error)`: 创建新的规则，返回新创建的规则ID。
*   `GetRuleByID(ctx context.Context, id int64) (*Rule, error)`: 根据ID查询规则信息。
*   `UpdateRule(ctx context.Context, id int64, updates map[string]interface{}) error`: 更新指定ID的规则信息。
*   `DeleteRule(ctx context.Context, id int64) error`: 删除指定ID的规则（硬删除）。
*   `ListRules(ctx context.Context, worldviewIDFilter int64, parentIDFilter int64, tagFilter string, page, pageSize int) ([]Rule, int64, error)`: 列出规则，支持按世界观ID、父ID和标签过滤，并进行分页。

### 3.3. BackgroundInfo 函数

*   `CreateBackgroundInfo(ctx context.Context, bi *BackgroundInfo) (int64, error)`: 创建新的背景信息，返回新创建的背景信息ID。
*   `GetBackgroundInfoByID(ctx context.Context, id int64) (*BackgroundInfo, error)`: 根据ID查询背景信息。
*   `UpdateBackgroundInfo(ctx context.Context, id int64, updates map[string]interface{}) error`: 更新指定ID的背景信息。
*   `DeleteBackgroundInfo(ctx context.Context, id int64) error`: 删除指定ID的背景信息（硬删除）。
*   `ListBackgroundInfos(ctx context.Context, worldviewIDFilter int64, parentIDFilter int64, tagFilter string, page, pageSize int) ([]BackgroundInfo, int64, error)`: 列出背景信息，支持按世界观ID、父ID和标签过滤，并进行分页。

## 4. 错误处理

DAL 层为每个实体定义了特定的错误类型，便于服务层进行精细化的错误处理。这些错误在各自的 Go 文件中定义（例如 `biz/dal/db/worldview.go` 中的 `ErrWorldviewNotFound`）。

*   `ErrWorldviewNotFound`, `ErrCreateWorldviewFailed`, etc.
*   `ErrRuleNotFound`, `ErrCreateRuleFailed`, etc.
*   `ErrBackgroundInfoNotFound`, `ErrCreateBackgroundInfoFailed`, etc.

## 5. 技术栈

*   **ORM**: [GORM](https://gorm.io/) - Go 语言的 ORM 库。
*   **测试数据库**: [SQLite](https://www.sqlite.org/) (in-memory) - 用于单元测试，通过 `gorm.io/driver/sqlite`驱动连接。

## 6. 单元测试

DAL 层的每个公开函数都配有单元测试，以确保其功能的正确性和稳定性。测试用例覆盖了所有 CRUD 操作和列表查询功能，包括各种边界条件和过滤条件。

*   测试文件与源文件位于同一目录，并以 `_test.go` 结尾 (例如 `worldview_test.go`)。
*   测试使用了 `github.com/stretchr/testify/assert` 包进行断言。
*   测试数据库使用内存中的 SQLite，确保测试的独立性和速度。

## 7. 未来展望与维护

本文档应随代码的迭代而持续更新。任何对数据模型、核心函数或错误处理的重大变更，都应及时反映在文档中，以保持其准确性和参考价值。
