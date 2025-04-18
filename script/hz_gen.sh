#!/bin/bash

# 用于生成Hertz框架代码的脚本
# 该脚本根据idl目录下的proto文件生成对应的handler、service、model和路由代码

# 确保脚本在任何错误时退出
set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
echo "项目根目录: $PROJECT_ROOT"

# 确保目录存在
mkdir -p "$PROJECT_ROOT/biz/handler/user"
mkdir -p "$PROJECT_ROOT/biz/service/user"
mkdir -p "$PROJECT_ROOT/biz/model/user"

# 检查idl目录下的proto文件
for proto_file in "$PROJECT_ROOT/idl"/*.proto; do
    if [ -f "$proto_file" ]; then
        filename=$(basename "$proto_file")
        service_name="${filename%.*}"
        
        echo "正在处理: $filename"
        
        # 获取包名
        package_name=$(grep "^package" "$proto_file" | awk '{print $2}' | tr -d ';')
        
        # 使用hz工具生成代码
        echo "生成 $service_name 服务的代码..."
        cd "$PROJECT_ROOT"
        
        # 生成model代码
        hz model \
            --idl "idl/$filename" \
            --model_dir "biz/model/$package_name"
        
        # 手动创建handler和service的基本结构
        echo "创建 $service_name 服务的handler和service基本结构..."
        
        # 为handler创建基本文件
        HANDLER_DIR="$PROJECT_ROOT/biz/handler/$package_name"
        if [ ! -f "$HANDLER_DIR/${service_name}_handler.go" ]; then
            cat > "$HANDLER_DIR/${service_name}_handler.go" << EOF
// 自动生成的handler文件，请根据需要修改

package $package_name

import (
	"context"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	
	"novelai/biz/model/$package_name"
)

// 用户注册
func Register(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${package_name}.RegisterRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${package_name}.RegisterResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现注册逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${package_name}.RegisterResponse{
		Code:    0,
		Message: "注册成功",
		UserId:  1001, // 示例用户ID
		Token:   "sample_token", // 示例token
	})
}

// 用户登录
func Login(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${package_name}.LoginRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${package_name}.LoginResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现登录逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${package_name}.LoginResponse{
		Code:    0,
		Message: "登录成功",
		UserId:  1001, // 示例用户ID
		Token:   "sample_token", // 示例token
	})
}

// 获取用户信息
func GetUser(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${package_name}.GetUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${package_name}.GetUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现获取用户信息逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${package_name}.GetUserResponse{
		Code:    0,
		Message: "获取成功",
		User: &${package_name}.User{
			Id:       1001,
			Username: "testuser",
			Nickname: "测试用户",
			Avatar:   "https://example.com/avatar.jpg",
			Email:    "test@example.com",
			Status:   0,
		},
	})
}

// 更新用户信息
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${package_name}.UpdateUserRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${package_name}.UpdateUserResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现更新用户信息逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${package_name}.UpdateUserResponse{
		Code:    0,
		Message: "更新成功",
	})
}
EOF
        fi
        
        # 为routes创建基本文件
        mkdir -p "$PROJECT_ROOT/biz/router/$package_name"
        ROUTER_DIR="$PROJECT_ROOT/biz/router/$package_name"
        if [ ! -f "$ROUTER_DIR/${service_name}_router.go" ]; then
            cat > "$ROUTER_DIR/${service_name}_router.go" << EOF
// 自动生成的路由文件，请根据需要修改

package $package_name

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	
	"novelai/biz/handler/$package_name"
)

// 注册用户相关路由
func RegisterRoutes(r *server.Hertz) {
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
		userGroup.GET("/info", handler.GetUser)
		userGroup.PUT("/update", handler.UpdateUser)
	}
}
EOF
        fi
        
        echo "$service_name 服务代码生成完成"
    fi
done

echo "所有服务代码生成完成"
