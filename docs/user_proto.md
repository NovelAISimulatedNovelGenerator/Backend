# User Proto 文档

## 概述

`user.proto` 定义了小说AI生成系统中的用户管理服务，包括用户注册、登录、信息获取与更新等功能。该服务通过Hertz框架的hz工具生成代码实现。

## 数据模型

### User（用户）

```protobuf
message User {
    int64 id = 1;                    // 用户ID
    string username = 2;             // 用户名
    string nickname = 3;             // 昵称
    string avatar = 4;               // 头像URL
    string email = 5;                // 电子邮箱
    int32 status = 6;                // 用户状态：0-正常，1-禁用
    int64 created_at = 7;            // 创建时间（Unix时间戳）
    int64 updated_at = 8;            // 更新时间（Unix时间戳）
}
```

## 服务接口

### UserService

```protobuf
service UserService {
    // 用户注册
    rpc Register(RegisterRequest) returns (RegisterResponse) {}
    
    // 用户登录
    rpc Login(LoginRequest) returns (LoginResponse) {}
    
    // 获取用户信息
    rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
    
    // 更新用户信息
    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {}
}
```

## 请求和响应消息

### 用户注册

**请求**: `RegisterRequest`
```protobuf
message RegisterRequest {
    string username = 1;             // 用户名
    string password = 2;             // 密码
    string nickname = 3;             // 昵称
    string email = 4;                // 电子邮箱
}
```

**响应**: `RegisterResponse`
```protobuf
message RegisterResponse {
    int32 code = 1;                  // 状态码：0-成功，非0-失败
    string message = 2;              // 响应消息
    int64 user_id = 3;               // 用户ID
    string token = 4;                // 用户认证令牌
}
```

### 用户登录

**请求**: `LoginRequest`
```protobuf
message LoginRequest {
    string username = 1;             // 用户名
    string password = 2;             // 密码
}
```

**响应**: `LoginResponse`
```protobuf
message LoginResponse {
    int32 code = 1;                  // 状态码：0-成功，非0-失败
    string message = 2;              // 响应消息
    int64 user_id = 3;               // 用户ID
    string token = 4;                // 用户认证令牌
}
```

### 获取用户信息

**请求**: `GetUserRequest`
```protobuf
message GetUserRequest {
    int64 user_id = 1;               // 用户ID
    string token = 2;                // 用户认证令牌
}
```

**响应**: `GetUserResponse`
```protobuf
message GetUserResponse {
    int32 code = 1;                  // 状态码：0-成功，非0-失败
    string message = 2;              // 响应消息
    User user = 3;                   // 用户信息
}
```

### 更新用户信息

**请求**: `UpdateUserRequest`
```protobuf
message UpdateUserRequest {
    int64 user_id = 1;               // 用户ID
    string token = 2;                // 用户认证令牌
    string nickname = 3;             // 昵称
    string avatar = 4;               // 头像URL
    string email = 5;                // 电子邮箱
}
```

**响应**: `UpdateUserResponse`
```protobuf
message UpdateUserResponse {
    int32 code = 1;                  // 状态码：0-成功，非0-失败
    string message = 2;              // 响应消息
}
```

## 设计特点

1. **认证机制**: 用户登录后获取token，后续操作需使用该token进行认证
2. **安全设计**: 敏感操作需要提供用户ID和认证令牌双重验证
3. **状态管理**: 支持用户状态标记，可用于禁用问题账户
4. **统一响应**: 所有响应都包含code和message字段，遵循一致的响应格式
5. **信息保护**: 用户密码仅在注册和登录请求中传输，不在响应中返回

## 使用注意事项

1. 客户端应对用户密码进行加密或哈希处理后再发送
2. 服务端应对密码进行安全存储，如使用加盐哈希等技术
3. token应设置合理的过期时间，并实现刷新机制
4. 用户头像URL应经过合法性验证，防止XSS攻击
5. 邮箱信息应进行格式验证和唯一性检查
6. 实现时应考虑添加密码修改、找回密码等扩展功能
