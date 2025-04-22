#!/bin/bash

# save服务代码生成脚本
# 根据save.proto文件生成对应的handler、service、model和路由代码

# 确保脚本在任何错误时退出
set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
echo "项目根目录: $PROJECT_ROOT"

# 确保目录存在
mkdir -p "$PROJECT_ROOT/biz/handler/save"
mkdir -p "$PROJECT_ROOT/biz/service/save"
mkdir -p "$PROJECT_ROOT/biz/model/save"
mkdir -p "$PROJECT_ROOT/biz/router/save"

# proto文件路径
PROTO_FILE="$PROJECT_ROOT/idl/save.proto"
SERVICE_NAME="save"
PACKAGE_NAME="save"

echo "正在处理: save.proto"

# 使用hz工具生成代码
echo "生成 $SERVICE_NAME 服务的代码..."
cd "$PROJECT_ROOT"

# 生成model代码
hz model \
    --idl "idl/save.proto" \
    --model_dir "biz/model/save"

# 创建handler基本结构
echo "创建 $SERVICE_NAME 服务的handler基本结构..."

# 为handler创建基本文件
HANDLER_DIR="$PROJECT_ROOT/biz/handler/save"
if [ ! -f "$HANDLER_DIR/${SERVICE_NAME}_handler.go" ]; then
    cat > "$HANDLER_DIR/${SERVICE_NAME}_handler.go" << EOF
// 自动生成的handler文件，请根据需要修改

package $PACKAGE_NAME

import (
	"context"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	
	"novelai/biz/model/$PACKAGE_NAME"
)

// 创建保存
func CreateSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${PACKAGE_NAME}.CreateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${PACKAGE_NAME}.CreateSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现创建保存逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${PACKAGE_NAME}.CreateSaveResponse{
		Code:    0,
		Message: "创建成功",
		SaveId:  "save_123456", // 示例保存ID
	})
}

// 获取保存
func GetSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${PACKAGE_NAME}.GetSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${PACKAGE_NAME}.GetSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现获取保存逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${PACKAGE_NAME}.GetSaveResponse{
		Code:    0,
		Message: "获取成功",
		Save: &${PACKAGE_NAME}.Save{
			Id:              1,
			UserId:          req.UserId,
			SaveId:          req.SaveId,
			SaveName:        "示例保存",
			SaveDescription: "这是一个示例保存项",
			SaveData:        "{\"content\":\"示例数据内容\"}",
			SaveType:        "draft",
			SaveStatus:      "active",
			CreatedAt:       1714406400, // 示例时间戳
			UpdatedAt:       1714406400, // 示例时间戳
		},
	})
}

// 更新保存
func UpdateSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${PACKAGE_NAME}.UpdateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${PACKAGE_NAME}.UpdateSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现更新保存逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${PACKAGE_NAME}.UpdateSaveResponse{
		Code:    0,
		Message: "更新成功",
	})
}

// 删除保存
func DeleteSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${PACKAGE_NAME}.DeleteSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${PACKAGE_NAME}.DeleteSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现删除保存逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${PACKAGE_NAME}.DeleteSaveResponse{
		Code:    0,
		Message: "删除成功",
	})
}

// 列出用户保存
func ListSaves(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(${PACKAGE_NAME}.ListSavesRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &${PACKAGE_NAME}.ListSavesResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// TODO: 实现列出用户保存逻辑
	
	// 返回成功响应
	c.JSON(consts.StatusOK, &${PACKAGE_NAME}.ListSavesResponse{
		Code:    0,
		Message: "获取成功",
		Saves: []*${PACKAGE_NAME}.Save{
			{
				Id:              1,
				UserId:          req.UserId,
				SaveId:          "save_123456",
				SaveName:        "示例保存1",
				SaveDescription: "这是一个示例保存项1",
				SaveData:        "{\"content\":\"示例数据内容1\"}",
				SaveType:        req.SaveType,
				SaveStatus:      "active",
				CreatedAt:       1714406400,
				UpdatedAt:       1714406400,
			},
			{
				Id:              2,
				UserId:          req.UserId,
				SaveId:          "save_234567",
				SaveName:        "示例保存2",
				SaveDescription: "这是一个示例保存项2",
				SaveData:        "{\"content\":\"示例数据内容2\"}",
				SaveType:        req.SaveType,
				SaveStatus:      "active",
				CreatedAt:       1714406500,
				UpdatedAt:       1714406500,
			},
		},
		Total: 2,
	})
}
EOF
fi

# 为routes创建基本文件
ROUTER_DIR="$PROJECT_ROOT/biz/router/save"
if [ ! -f "$ROUTER_DIR/${SERVICE_NAME}_router.go" ]; then
    cat > "$ROUTER_DIR/${SERVICE_NAME}_router.go" << EOF
// 自动生成的路由文件，请根据需要修改

package $PACKAGE_NAME

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	
	"novelai/biz/handler/$PACKAGE_NAME"
)

// 注册保存相关路由
func RegisterRoutes(r *server.Hertz) {
	saveGroup := r.Group("/api/save")
	{
		saveGroup.POST("/create", handler.CreateSave)
		saveGroup.GET("/get", handler.GetSave)
		saveGroup.PUT("/update", handler.UpdateSave)
		saveGroup.DELETE("/delete", handler.DeleteSave)
		saveGroup.GET("/list", handler.ListSaves)
	}
}
EOF
fi

echo "$SERVICE_NAME 服务代码生成完成"
echo "脚本执行完毕"

# 为脚本添加执行权限
chmod +x "$0"
