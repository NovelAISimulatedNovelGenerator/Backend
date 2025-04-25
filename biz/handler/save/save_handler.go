// 自动生成的handler文件，请根据需要修改

package save

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"novelai/pkg/middleware"

	"novelai/biz/model/save"
	svc "novelai/biz/service/save"
)

// 创建保存
// CreateSave 创建保存项，完善错误处理，返回结构化响应
// 参数: ctx 上下文，c Hertz请求上下文
// 返回: JSON结构化响应（含错误码、消息、数据）
func CreateSave(ctx context.Context, c *app.RequestContext) {
	// 1. 记录请求参数，便于调试
	hlog.Debugf("[CreateSave] 请求参数: %+v", c.Request.Body())

	// 2. 绑定并校验 body 参数
	req := new(save.CreateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.CreateSaveResponse{
			Code:    400,
			Message: "参数绑定或校验失败: " + err.Error(),
		})
		return
	}
	if req.SaveName == "" || req.SaveData == "" || req.SaveType == "" {
		c.JSON(consts.StatusBadRequest, &save.CreateSaveResponse{
			Code:    400,
			Message: "缺少必需参数: save_name/save_data/save_type",
		})
		return
	}

	// 3. 解析 JWT 用户ID，类型兼容与校验
	idVal, ok := c.Get(middleware.IdentityKey)
	if !ok {
		c.JSON(consts.StatusUnauthorized, &save.CreateSaveResponse{
			Code:    401,
			Message: "未登录或Token无效",
		})
		return
	}
	var userId int64
	switch v := idVal.(type) {
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		c.JSON(consts.StatusUnauthorized, &save.CreateSaveResponse{
			Code:    401,
			Message: "无法解析用户ID（JWT类型错误）",
		})
		return
	}
	if userId <= 0 {
		c.JSON(consts.StatusUnauthorized, &save.CreateSaveResponse{
			Code:    401,
			Message: "用户ID无效",
		})
		return
	}

	// 4. 调用 service 层创建保存项，细致处理业务/数据库错误
	serviceReq := &svc.CreateSaveServiceRequest{
		UserId:          userId,
		SaveName:        req.SaveName,
		SaveDescription: req.SaveDescription,
		SaveData:        req.SaveData,
		SaveType:        req.SaveType,
	}
	serviceResp, err := svc.Create(ctx, serviceReq)
	if err != nil {
		switch err.Error() {
		case "请求参数不合法":
			c.JSON(consts.StatusBadRequest, &save.CreateSaveResponse{
				Code:    400,
				Message: "请求参数不合法",
			})
			return
		case "创建存档失败":
			c.JSON(consts.StatusInternalServerError, &save.CreateSaveResponse{
				Code:    500,
				Message: "创建存档失败",
			})
			return
		default:
			c.JSON(consts.StatusInternalServerError, &save.CreateSaveResponse{
				Code:    500,
				Message: "服务器内部错误: " + err.Error(),
			})
			return
		}
	}
	if serviceResp == nil || serviceResp.SaveId == "" {
		c.JSON(consts.StatusInternalServerError, &save.CreateSaveResponse{
			Code:    500,
			Message: "创建失败，未返回保存ID",
		})
		return
	}

	// 5. 返回成功响应
	c.JSON(consts.StatusOK, &save.CreateSaveResponse{
		Code:    200,
		Message: "创建成功",
		SaveId:  serviceResp.SaveId,
	})
}
// 流程说明：
// 1. 参数绑定失败/缺失直接 400 返回
// 2. JWT 解析失败或用户ID无效 401 返回
// 3. 调用 service 层后，细分参数、业务、数据库等错误，分别返回 400/500
// 4. 成功返回 200 及数据
// 5. 所有分支均结构化响应，便于前端统一处理
// 6. 变量作用域最小化，避免全局变量和递归，圈复杂度低于 10


// GetSave 获取指定保存项，完善错误处理，返回结构化响应
// 参数: ctx 上下文，c Hertz请求上下文
// 返回: JSON结构化响应（含错误码、消息、数据）
func GetSave(ctx context.Context, c *app.RequestContext) {
	// 1. 记录请求参数，便于调试
	hlog.Debugf("[GetSave] 请求参数: %+v", c.Request.Body())
	hlog.Debugf("[GetSave] query参数: %+v", c.QueryArgs().QueryString())

	// 2. 绑定并校验 query 参数
	req := new(save.GetSaveRequest)
	if err := c.BindQuery(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.GetSaveResponse{
			Code:    400,
			Message: "参数绑定失败: " + err.Error(),
		})
		return
	}
	if req.SaveId == "" {
		c.JSON(consts.StatusBadRequest, &save.GetSaveResponse{
			Code:    400,
			Message: "缺少必需参数: save_id",
		})
		return
	}

	// 3. 解析 JWT 用户ID，类型兼容与校验
	idVal, ok := c.Get(middleware.IdentityKey)
	if !ok {
		c.JSON(consts.StatusUnauthorized, &save.GetSaveResponse{
			Code:    401,
			Message: "未登录或Token无效",
		})
		return
	}
	var userId int64
	switch v := idVal.(type) {
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		c.JSON(consts.StatusUnauthorized, &save.GetSaveResponse{
			Code:    401,
			Message: "无法解析用户ID（JWT类型错误）",
		})
		return
	}
	if userId <= 0 {
		c.JSON(consts.StatusUnauthorized, &save.GetSaveResponse{
			Code:    401,
			Message: "用户ID无效",
		})
		return
	}

	// 4. 调用 service 层获取保存项，细致处理业务/数据库错误
	serviceReq := &svc.GetSaveServiceRequest{
		UserId: userId,
		SaveId: req.SaveId,
	}
	serviceResp, err := svc.Get(ctx, serviceReq)
	if err != nil {
		// 业务/数据库错误细分
		switch err.Error() {
		case "请求参数不合法":
			c.JSON(consts.StatusBadRequest, &save.GetSaveResponse{
				Code:    400,
				Message: "请求参数不合法",
			})
			return
		case "存档不存在":
			c.JSON(consts.StatusNotFound, &save.GetSaveResponse{
				Code:    404,
				Message: "保存项不存在",
			})
			return
		default:
			c.JSON(consts.StatusInternalServerError, &save.GetSaveResponse{
				Code:    500,
				Message: "服务器内部错误: " + err.Error(),
			})
			return
		}
	}
	if serviceResp == nil || serviceResp.Save == nil {
		c.JSON(consts.StatusNotFound, &save.GetSaveResponse{
			Code:    404,
			Message: "保存项不存在",
		})
		return
	}

	// 5. 返回成功响应
	c.JSON(consts.StatusOK, &save.GetSaveResponse{
		Code:    200,
		Message: "获取成功",
		Save:    serviceResp.Save,
	})
}
// 流程说明：
// 1. 参数绑定失败/缺失直接 400 返回
// 2. JWT 解析失败或用户ID无效 401 返回
// 3. 调用 service 层后，细分参数、业务、数据库等错误，分别返回 400/404/500
// 4. 成功返回 200 及数据
// 5. 所有分支均结构化响应，便于前端统一处理
// 6. 变量作用域最小化，避免全局变量和递归，圈复杂度低于 10


// 更新保存
// UpdateSave 更新保存项，完善错误处理，返回结构化响应
// 参数: ctx 上下文，c Hertz请求上下文
// 返回: JSON结构化响应（含错误码、消息、数据）
func UpdateSave(ctx context.Context, c *app.RequestContext) {
	// 1. 记录请求参数，便于调试
	hlog.Debugf("[UpdateSave] 请求参数: %+v", c.Request.Body())

	// 2. 绑定并校验 body 参数
	req := new(save.UpdateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.UpdateSaveResponse{
			Code:    400,
			Message: "参数绑定或校验失败: " + err.Error(),
		})
		return
	}
	if req.SaveId == "" {
		c.JSON(consts.StatusBadRequest, &save.UpdateSaveResponse{
			Code:    400,
			Message: "缺少必需参数: save_id",
		})
		return
	}

	// 3. 解析 JWT 用户ID，类型兼容与校验
	idVal, ok := c.Get(middleware.IdentityKey)
	if !ok {
		c.JSON(consts.StatusUnauthorized, &save.UpdateSaveResponse{
			Code:    401,
			Message: "未登录或Token无效",
		})
		return
	}
	var userId int64
	switch v := idVal.(type) {
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		c.JSON(consts.StatusUnauthorized, &save.UpdateSaveResponse{
			Code:    401,
			Message: "无法解析用户ID（JWT类型错误）",
		})
		return
	}
	if userId <= 0 {
		c.JSON(consts.StatusUnauthorized, &save.UpdateSaveResponse{
			Code:    401,
			Message: "用户ID无效",
		})
		return
	}

	// 4. 调用 service 层更新保存项，细致处理业务/数据库错误
	serviceReq := &svc.UpdateSaveServiceRequest{
		UserId:          userId,
		SaveId:          req.SaveId,
		SaveName:        req.SaveName,
		SaveDescription: req.SaveDescription,
		SaveData:        req.SaveData,
	}
	_, err := svc.Update(ctx, serviceReq)
	if err != nil {
		switch err.Error() {
		case "请求参数不合法":
			c.JSON(consts.StatusBadRequest, &save.UpdateSaveResponse{
				Code:    400,
				Message: "请求参数不合法",
			})
			return
		case "存档不存在":
			c.JSON(consts.StatusNotFound, &save.UpdateSaveResponse{
				Code:    404,
				Message: "保存项不存在",
			})
			return
		case "更新存档失败":
			c.JSON(consts.StatusInternalServerError, &save.UpdateSaveResponse{
				Code:    500,
				Message: "更新存档失败",
			})
			return
		default:
			c.JSON(consts.StatusInternalServerError, &save.UpdateSaveResponse{
				Code:    500,
				Message: "服务器内部错误: " + err.Error(),
			})
			return
		}
	}

	// 5. 返回成功响应
	c.JSON(consts.StatusOK, &save.UpdateSaveResponse{
		Code:    200,
		Message: "更新成功",
	})
}
// 流程说明：
// 1. 参数绑定失败/缺失直接 400 返回
// 2. JWT 解析失败或用户ID无效 401 返回
// 3. 调用 service 层后，细分参数、业务、数据库等错误，分别返回 400/404/500
// 4. 成功返回 200 及数据
// 5. 所有分支均结构化响应，便于前端统一处理
// 6. 变量作用域最小化，避免全局变量和递归，圈复杂度低于 10


// 删除保存
// DeleteSave 删除保存项，完善错误处理，返回结构化响应
// 参数: ctx 上下文，c Hertz请求上下文
// 返回: JSON结构化响应（含错误码、消息、数据）
func DeleteSave(ctx context.Context, c *app.RequestContext) {
	// 1. 记录请求参数，便于调试
	hlog.Debugf("[DeleteSave] 请求参数: %+v", c.Request.Body())

	// 2. 绑定并校验 body 参数
	req := new(save.DeleteSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.DeleteSaveResponse{
			Code:    400,
			Message: "参数绑定或校验失败: " + err.Error(),
		})
		return
	}
	if req.SaveId == "" {
		c.JSON(consts.StatusBadRequest, &save.DeleteSaveResponse{
			Code:    400,
			Message: "缺少必需参数: save_id",
		})
		return
	}

	// 3. 解析 JWT 用户ID，类型兼容与校验
	idVal, ok := c.Get(middleware.IdentityKey)
	if !ok {
		c.JSON(consts.StatusUnauthorized, &save.DeleteSaveResponse{
			Code:    401,
			Message: "未登录或Token无效",
		})
		return
	}
	var userId int64
	switch v := idVal.(type) {
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		c.JSON(consts.StatusUnauthorized, &save.DeleteSaveResponse{
			Code:    401,
			Message: "无法解析用户ID（JWT类型错误）",
		})
		return
	}
	if userId <= 0 {
		c.JSON(consts.StatusUnauthorized, &save.DeleteSaveResponse{
			Code:    401,
			Message: "用户ID无效",
		})
		return
	}

	// 4. 调用 service 层删除保存项，细致处理业务/数据库错误
	serviceReq := &svc.DeleteSaveServiceRequest{
		UserId: userId,
		SaveId: req.SaveId,
	}
	_, err := svc.Delete(ctx, serviceReq)
	if err != nil {
		switch err.Error() {
		case "请求参数不合法":
			c.JSON(consts.StatusBadRequest, &save.DeleteSaveResponse{
				Code:    400,
				Message: "请求参数不合法",
			})
			return
		case "存档不存在":
			c.JSON(consts.StatusNotFound, &save.DeleteSaveResponse{
				Code:    404,
				Message: "保存项不存在",
			})
			return
		default:
			c.JSON(consts.StatusInternalServerError, &save.DeleteSaveResponse{
				Code:    500,
				Message: "服务器内部错误: " + err.Error(),
			})
			return
		}
	}

	// 5. 返回成功响应
	c.JSON(consts.StatusOK, &save.DeleteSaveResponse{
		Code:    200,
		Message: "删除成功",
	})
}
// 流程说明：
// 1. 参数绑定失败/缺失直接 400 返回
// 2. JWT 解析失败或用户ID无效 401 返回
// 3. 调用 service 层后，细分参数、业务、数据库等错误，分别返回 400/404/500
// 4. 成功返回 200 及数据
// 5. 所有分支均结构化响应，便于前端统一处理
// 6. 变量作用域最小化，避免全局变量和递归，圈复杂度低于 10


// 列出用户保存
// ListSaves 列出用户保存项，完善错误处理，返回结构化响应
// 参数: ctx 上下文，c Hertz请求上下文
// 返回: JSON结构化响应（含错误码、消息、数据）
func ListSaves(ctx context.Context, c *app.RequestContext) {
	// 1. 记录请求参数，便于调试
	hlog.Debugf("[ListSaves] 请求参数: %+v", c.Request.Body())

	// 2. 绑定并校验 body 参数
	req := new(save.ListSavesRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.ListSavesResponse{
			Code:    400,
			Message: "参数绑定或校验失败: " + err.Error(),
		})
		return
	}
	if req.Page <= 0 || req.PageSize <= 0 {
		c.JSON(consts.StatusBadRequest, &save.ListSavesResponse{
			Code:    400,
			Message: "分页参数非法",
		})
		return
	}

	// 3. 解析 JWT 用户ID，类型兼容与校验
	idVal, ok := c.Get(middleware.IdentityKey)
	if !ok {
		c.JSON(consts.StatusUnauthorized, &save.ListSavesResponse{
			Code:    401,
			Message: "未登录或Token无效",
		})
		return
	}
	var userId int64
	switch v := idVal.(type) {
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		c.JSON(consts.StatusUnauthorized, &save.ListSavesResponse{
			Code:    401,
			Message: "无法解析用户ID（JWT类型错误）",
		})
		return
	}
	if userId <= 0 {
		c.JSON(consts.StatusUnauthorized, &save.ListSavesResponse{
			Code:    401,
			Message: "用户ID无效",
		})
		return
	}

	// 4. 调用 service 层列出保存项，细致处理业务/数据库错误
	serviceReq := &svc.ListSavesServiceRequest{
		UserId:   userId,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	serviceResp, err := svc.List(ctx, serviceReq)
	if err != nil {
		switch err.Error() {
		case "请求参数不合法":
			c.JSON(consts.StatusBadRequest, &save.ListSavesResponse{
				Code:    400,
				Message: "请求参数不合法",
			})
			return
		default:
			c.JSON(consts.StatusInternalServerError, &save.ListSavesResponse{
				Code:    500,
				Message: "服务器内部错误: " + err.Error(),
			})
			return
		}
	}
	if serviceResp == nil {
		c.JSON(consts.StatusInternalServerError, &save.ListSavesResponse{
			Code:    500,
			Message: "获取失败，未返回数据",
		})
		return
	}

	// 5. 返回成功响应
	c.JSON(consts.StatusOK, &save.ListSavesResponse{
		Code:    200,
		Message: "获取成功",
		Saves:   serviceResp.Saves,
		Total:   int32(serviceResp.Total),
	})
}
// 流程说明：
// 1. 参数绑定失败/缺失直接 400 返回
// 2. JWT 解析失败或用户ID无效 401 返回
// 3. 调用 service 层后，细分参数、业务、数据库等错误，分别返回 400/500
// 4. 成功返回 200 及数据
// 5. 所有分支均结构化响应，便于前端统一处理
// 6. 变量作用域最小化，避免全局变量和递归，圈复杂度低于 10

