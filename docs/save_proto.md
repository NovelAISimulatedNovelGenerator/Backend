# Save Proto 文档

## 概述

`save.proto` 定义了小说AI生成系统中的保存功能服务，用于管理用户的创作内容和配置。该服务通过Hertz框架的hz工具生成代码实现。

## 数据模型

### Save（保存项）

```protobuf
message Save {
  int64 id = 1;                     // 保存项ID
  int64 user_id = 2;                // 用户ID
  string save_id = 3;               // 保存项唯一标识符
  string save_name = 4;             // 保存项名称
  string save_description = 5;      // 保存项描述
  string save_data = 6;             // 保存的具体内容（如JSON字符串）
  string save_type = 7;             // 保存类型（如草稿、配置等）
  string save_status = 8;           // 保存状态（如active、deleted等）
  int64 created_at = 9;             // 创建时间（unix时间戳）
  int64 updated_at = 10;            // 更新时间（unix时间戳）
}
```

## 服务接口

### SaveService

```protobuf
service SaveService {
  // 创建保存
  rpc CreateSave(CreateSaveRequest) returns (CreateSaveResponse) {}
  
  // 获取保存
  rpc GetSave(GetSaveRequest) returns (GetSaveResponse) {}
  
  // 更新保存
  rpc UpdateSave(UpdateSaveRequest) returns (UpdateSaveResponse) {}
  
  // 删除保存
  rpc DeleteSave(DeleteSaveRequest) returns (DeleteSaveResponse) {}
  
  // 列出用户保存
  rpc ListSaves(ListSavesRequest) returns (ListSavesResponse) {}
}
```

## 请求和响应消息

### 创建保存项

**请求**: `CreateSaveRequest`
```protobuf
message CreateSaveRequest {
  int64 user_id = 1;                 // 用户ID
  string token = 2;                  // 用户认证令牌
  string save_name = 3;              // 保存项名称
  string save_description = 4;       // 保存项描述
  string save_data = 5;              // 保存的具体内容
  string save_type = 6;              // 保存类型
}
```

**响应**: `CreateSaveResponse`
```protobuf
message CreateSaveResponse {
  int32 code = 1;                    // 状态码：0-成功，非0-失败
  string message = 2;                // 响应消息
  string save_id = 3;                // 保存项ID
}
```

### 获取保存项

**请求**: `GetSaveRequest`
```protobuf
message GetSaveRequest {
  int64 user_id = 1;                 // 用户ID
  string token = 2;                  // 用户认证令牌
  string save_id = 3;                // 保存项ID
}
```

**响应**: `GetSaveResponse`
```protobuf
message GetSaveResponse {
  int32 code = 1;                    // 状态码：0-成功，非0-失败
  string message = 2;                // 响应消息
  Save save = 3;                     // 保存项信息
}
```

### 更新保存项

**请求**: `UpdateSaveRequest`
```protobuf
message UpdateSaveRequest {
  int64 user_id = 1;                 // 用户ID
  string token = 2;                  // 用户认证令牌
  string save_id = 3;                // 保存项ID
  string save_name = 4;              // 保存项名称
  string save_description = 5;       // 保存项描述
  string save_data = 6;              // 保存的具体内容
  string save_status = 7;            // 保存状态
}
```

**响应**: `UpdateSaveResponse`
```protobuf
message UpdateSaveResponse {
  int32 code = 1;                    // 状态码：0-成功，非0-失败
  string message = 2;                // 响应消息
}
```

### 删除保存项

**请求**: `DeleteSaveRequest`
```protobuf
message DeleteSaveRequest {
  int64 user_id = 1;                 // 用户ID
  string token = 2;                  // 用户认证令牌
  string save_id = 3;                // 保存项ID
}
```

**响应**: `DeleteSaveResponse`
```protobuf
message DeleteSaveResponse {
  int32 code = 1;                    // 状态码：0-成功，非0-失败
  string message = 2;                // 响应消息
}
```

### 列出保存项

**请求**: `ListSavesRequest`
```protobuf
message ListSavesRequest {
  int64 user_id = 1;                 // 用户ID
  string token = 2;                  // 用户认证令牌
  string save_type = 3;              // 保存类型（可选）
  int32 page = 4;                    // 页码，从1开始
  int32 page_size = 5;               // 每页数量
}
```

**响应**: `ListSavesResponse`
```protobuf
message ListSavesResponse {
  int32 code = 1;                    // 状态码：0-成功，非0-失败
  string message = 2;                // 响应消息
  repeated Save saves = 3;           // 保存项列表
  int32 total = 4;                   // 总数量
}
```

## 设计特点

1. **用户认证**: 所有操作都需要用户ID和认证令牌，确保数据安全
2. **分类支持**: 通过save_type字段支持不同类型的保存项
3. **状态管理**: 通过save_status字段管理保存项的生命周期
4. **分页查询**: 列表查询支持分页功能
5. **筛选功能**: 支持按保存类型筛选查询结果
6. **统一响应**: 所有响应都包含code和message字段，遵循一致的响应格式

## 使用注意事项

1. save_data字段通常存储JSON格式的字符串，需要在应用层进行序列化和反序列化
2. 客户端应妥善保管token，所有请求都需要验证token有效性
3. 更新操作只会修改请求中指定的字段，未指定的字段保持不变
4. 列表查询时，建议根据实际需求设置合理的page_size
