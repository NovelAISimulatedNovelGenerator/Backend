# Background Proto 文档

## 概述

`background.proto` 定义了小说AI生成系统中的背景信息服务，包括世界观、规则和背景信息三个主要模型及其相关的CRUD操作接口。这些接口由Hertz框架的hz工具生成代码实现。

## 数据模型

### Worldview（世界观）

```protobuf
message Worldview {
    int64 id = 1;                               // 世界观ID
    string name = 2;                            // 世界观名称
    string description = 3;                     // 世界观详细描述
    string tag = 4;                             // 标签，多个标签用英文逗号分隔
    int64 parent_id = 5;                        // 父世界观ID，0表示主世界观 (顶级世界观)
    google.protobuf.Timestamp created_at = 6;   // 创建时间
    google.protobuf.Timestamp updated_at = 7;   // 更新时间
}
```

### Rule（规则）

```protobuf
message Rule {
    int64 id = 1;                               // 规则ID
    int64 worldview_id = 2;                     // 所属世界观ID
    string name = 3;                            // 规则名称
    string description = 4;                     // 规则详细描述
    string tag = 5;                             // 标签，多个标签用英文逗号分隔
    int64 parent_id = 6;                        // 父规则ID，0表示主规则 (顶级规则)
    google.protobuf.Timestamp created_at = 7;   // 创建时间
    google.protobuf.Timestamp updated_at = 8;   // 更新时间
}
```

### BackgroundInfo（背景信息）

```protobuf
message BackgroundInfo {
    int64 id = 1;                               // 背景ID
    int64 worldview_id = 2;                     // 所属世界观ID
    string name = 3;                            // 背景名称
    string description = 4;                     // 背景详细描述
    string tag = 5;                             // 标签，多个标签用英文逗号分隔
    int64 parent_id = 6;                        // 父背景ID，0表示主背景 (顶级背景)
    google.protobuf.Timestamp created_at = 7;   // 创建时间
    google.protobuf.Timestamp updated_at = 8;   // 更新时间
}
```

## 服务接口

### BackgroundService

```protobuf
service BackgroundService {
    // Worldview RPCs
    rpc CreateWorldview(CreateWorldviewRequest) returns (CreateWorldviewResponse);
    rpc GetWorldview(GetWorldviewRequest) returns (GetWorldviewResponse);
    rpc UpdateWorldview(UpdateWorldviewRequest) returns (UpdateWorldviewResponse);
    rpc DeleteWorldview(DeleteWorldviewRequest) returns (DeleteWorldviewResponse);
    rpc ListWorldviews(ListWorldviewsRequest) returns (ListWorldviewsResponse);
    
    // Rule RPCs
    rpc CreateRule(CreateRuleRequest) returns (CreateRuleResponse);
    rpc GetRule(GetRuleRequest) returns (GetRuleResponse);
    rpc UpdateRule(UpdateRuleRequest) returns (UpdateRuleResponse);
    rpc DeleteRule(DeleteRuleRequest) returns (DeleteRuleResponse);
    rpc ListRules(ListRulesRequest) returns (ListRulesResponse);
    
    // BackgroundInfo RPCs
    rpc CreateBackgroundInfo(CreateBackgroundInfoRequest) returns (CreateBackgroundInfoResponse);
    rpc GetBackgroundInfo(GetBackgroundInfoRequest) returns (GetBackgroundInfoResponse);
    rpc UpdateBackgroundInfo(UpdateBackgroundInfoRequest) returns (UpdateBackgroundInfoResponse);
    rpc DeleteBackgroundInfo(DeleteBackgroundInfoRequest) returns (DeleteBackgroundInfoResponse);
    rpc ListBackgroundInfos(ListBackgroundInfosRequest) returns (ListBackgroundInfosResponse);
}
```

## 请求和响应消息

### Worldview 操作

#### 创建世界观
- **请求**: `CreateWorldviewRequest` - 包含世界观名称、描述、标签和父世界观ID
- **响应**: `CreateWorldviewResponse` - 返回状态码、消息和创建的世界观信息

#### 获取世界观
- **请求**: `GetWorldviewRequest` - 指定要获取的世界观ID
- **响应**: `GetWorldviewResponse` - 返回状态码、消息和世界观信息

#### 更新世界观
- **请求**: `UpdateWorldviewRequest` - 指定要更新的世界观ID和新的属性值
- **响应**: `UpdateWorldviewResponse` - 返回状态码和消息

#### 删除世界观
- **请求**: `DeleteWorldviewRequest` - 指定要删除的世界观ID
- **响应**: `DeleteWorldviewResponse` - 返回状态码和消息

#### 列出世界观
- **请求**: `ListWorldviewsRequest` - 包含父世界观ID筛选、标签筛选和分页参数
- **响应**: `ListWorldviewsResponse` - 返回状态码、消息、世界观列表和总数量

### Rule 操作

类似的请求和响应结构适用于Rule（规则）的CRUD操作。

### BackgroundInfo 操作

类似的请求和响应结构适用于BackgroundInfo（背景信息）的CRUD操作。

## 设计特点

1. **层级结构**: 所有模型都支持父子关系，通过parent_id形成层次结构
2. **标签系统**: 通过tag字段支持灵活的分类和搜索
3. **分页查询**: 列表查询支持分页功能
4. **筛选功能**: 提供多种筛选条件的列表查询
5. **统一响应**: 所有响应都包含code和message字段，遵循一致的响应格式

## 使用注意事项

1. 创建子项目（子世界观、子规则、子背景）时，需要指定正确的parent_id
2. 删除操作应考虑级联删除子项目的处理
3. 更新操作只会修改请求中指定的字段，未指定的字段保持不变
4. 标签字段使用英文逗号分隔多个标签
