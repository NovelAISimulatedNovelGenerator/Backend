# 数据访问层（DAL）文档

## 概述

数据访问层（Data Access Layer）是NovelAI项目的数据持久化和访问核心，负责隔离业务逻辑与数据存储细节，提供统一的数据访问接口。该层基于GORM框架实现，支持多种数据库后端，并采用结构化错误处理和数据模型定义。

## 目录结构

```
/biz/dal/
├── init.go                  # 数据访问层初始化入口
└── db/                      # 数据库访问模块
    ├── init.go              # 数据库连接初始化
    ├── migration.go         # 数据库迁移管理
    ├── user.go              # 用户模型和数据操作
    ├── user_test.go         # 用户模型测试
    ├── save.go              # 保存模型和数据操作
    └── save_test.go         # 保存模型测试
```

## 初始化流程

数据访问层采用分层初始化模式：

1. **DAL层初始化**：`dal.Init()` 函数接收整体配置，并调用各个存储组件的初始化函数
2. **数据库初始化**：`db.Init()` 函数建立数据库连接，配置连接池和日志
3. **迁移管理**：`migration.go` 负责自动创建和更新数据库表结构

```go
// 初始化示例
config := &dal.Config{
    DB: &db.Config{
        DriverName: "postgres",
        DSN:        "host=localhost user=postgres password=password dbname=novelai port=5432 sslmode=disable",
        Active:     10,
        Idle:       5,
    },
}
dal.Init(config)
```

## 数据模型

### 用户模型 (User)

用户模型表示系统中的注册用户，包含身份验证和个人信息属性。

```go
type User struct {
    ID              int64          `gorm:"primaryKey;autoIncrement"`
    Username        string         `gorm:"type:varchar(64);uniqueIndex;not null"`
    Password        string         `gorm:"type:varchar(256);not null"`
    Nickname        string         `gorm:"type:varchar(64)"`
    Email           *string        `gorm:"type:varchar(128);uniqueIndex"`
    Avatar          string         `gorm:"type:varchar(256)"`
    BackgroundImage string         `gorm:"type:varchar(256)"`
    Signature       string         `gorm:"type:varchar(512)"`
    IsAdmin         bool           `gorm:"default:false"`
    Status          int8           `gorm:"default:1"`
    LastLoginTime   *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}
```

主要字段说明：
- `ID`: 用户唯一标识
- `Username`: 用户名，唯一
- `Password`: 加密存储的密码
- `Status`: 用户状态，1-正常，2-禁用
- `DeletedAt`: 软删除时间标记

### 保存模型 (Save)

保存模型表示用户的创作内容存储项，包含内容数据和元信息。

```go
type Save struct {
    ID              int64          `gorm:"primaryKey;autoIncrement"`
    UserID          int64          `gorm:"index;not null"`
    SaveID          string         `gorm:"type:varchar(64);uniqueIndex;not null"`
    SaveName        string         `gorm:"type:varchar(128);not null"`
    SaveDescription string         `gorm:"type:varchar(512)"`
    SaveData        string         `gorm:"type:text;not null"`
    SaveType        string         `gorm:"type:varchar(32);not null"`
    SaveStatus      string         `gorm:"type:varchar(16);not null"`
    CreatedAt       int64          `gorm:"autoCreateTime"`
    UpdatedAt       int64          `gorm:"autoUpdateTime"`
}
```

主要字段说明：
- `ID`: 内部唯一标识
- `UserID`: 所属用户ID
- `SaveID`: 业务层面的唯一标识
- `SaveData`: 保存的具体内容，通常是JSON格式
- `SaveType`: 保存类型，如草稿、配置等
- `SaveStatus`: 保存状态，如active、deleted等

## 数据操作接口

### 用户操作接口

```go
// 创建新用户
CreateUser(user *User) (int64, error)

// 通过用户名查询用户
QueryUserByUsername(username string) (*User, error)

// 通过ID查询用户
QueryUserByID(userID int64) (*User, error)

// 验证用户名和密码
VerifyUser(username, password string) (int64, error)

// 更新用户资料
UpdateUserProfile(user *User) error

// 更新用户密码
UpdateUserPassword(userID int64, newPassword string) error

// 删除用户（软删除）
DeleteUser(userID int64) error

// 列出用户
ListUsers(page, pageSize int) ([]User, int64, error)

// 检查用户是否存在
CheckUserExists(userID int64) (bool, error)
```

### 保存操作接口

```go
// 创建新保存项
CreateSave(save *Save) (int64, error)

// 通过ID查询保存项
QuerySaveByID(saveID int64) (*Save, error)

// 查询用户的保存项
QuerySavesByUser(userID int64, page, pageSize int) ([]Save, int64, error)

// 通过SaveID查询保存项
QuerySavesBySaveID(saveID string) (*Save, error)

// 更新保存项
UpdateSave(save *Save) error

// 删除保存项
DeleteSave(saveID int64) error

// 列出保存项
ListSaves(page, pageSize int) ([]Save, int64, error)

// 检查保存项是否存在
CheckSaveExists(saveID int64) (bool, error)
```

## 错误处理

数据访问层定义了一系列具有语义的错误常量，以便上层业务逻辑可以针对不同错误类型做出相应处理：

### 用户相关错误

```go
// 用户不存在
ErrUserNotFound = errors.New("用户不存在")

// 用户名已被占用
ErrUserAlreadyExists = errors.New("用户名已存在")

// 密码不正确
ErrInvalidPassword = errors.New("密码验证失败")

// 创建用户失败
ErrCreateUserFailed = errors.New("创建用户失败")

// 更新用户信息失败
ErrUpdateUserFailed = errors.New("更新用户信息失败")
```

### 保存相关错误

```go
// 保存项不存在
ErrSaveNotFound = errors.New("存档不存在")

// 创建保存项失败
ErrCreateSaveFailed = errors.New("创建存档失败")

// 更新保存项失败
ErrUpdateSaveFailed = errors.New("更新存档失败")
```

## 数据库支持

当前实现主要支持PostgreSQL数据库，但架构设计允许扩展支持其他数据库：

- **PostgreSQL**: 当前主要支持的数据库
- **MySQL**: 预留支持接口
- **SQLite**: 预留支持接口

## 事务处理

数据访问层支持事务操作，可以在业务层面管理事务：

```go
// 开始事务
tx := DB.Begin()

// 在事务中执行操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback() // 发生错误时回滚
    return err
}

// 提交事务
tx.Commit()
```

## 性能优化

1. **连接池管理**: 通过Active和Idle参数配置连接池大小
2. **预编译语句**: 启用PrepareStmt提高查询性能
3. **跳过默认事务**: 启用SkipDefaultTransaction提高单次查询性能
4. **OpenTracing集成**: 支持分布式追踪，便于性能监控

## 最佳实践

1. **错误处理**: 使用预定义错误常量，便于业务层识别错误类型
2. **分页查询**: 查询大量数据时总是使用分页
3. **软删除**: 使用GORM的软删除功能，而不是直接删除数据
4. **索引优化**: 对频繁查询的字段创建适当的索引
5. **参数验证**: 在调用数据库操作前验证参数有效性
6. **连接管理**: 适当配置连接池参数，避免连接耗尽或资源浪费
