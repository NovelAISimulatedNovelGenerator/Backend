// 自动生成的handler文件，请根据需要修改

package save

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"novelai/biz/model/save"
	svc "novelai/biz/service/save"
)

// 创建保存
func CreateSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(save.CreateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.CreateSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 调用 service 层创建保存逻辑
	serviceReq := &svc.CreateSaveServiceRequest{
		UserId:          req.UserId,
		SaveName:        req.SaveName,
		SaveDescription: req.SaveDescription,
		SaveData:        req.SaveData,
		SaveType:        req.SaveType,
	}
	serviceResp, err := svc.Create(ctx, serviceReq)
	if err != nil {
		// 业务错误处理
		c.JSON(consts.StatusInternalServerError, &save.CreateSaveResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	// 返回成功响应
	c.JSON(consts.StatusOK, &save.CreateSaveResponse{
		Code:    0,
		Message: "创建成功",
		SaveId:  serviceResp.SaveId,
	})
}

// 获取保存
func GetSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(save.GetSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.GetSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 调用 service 层获取保存逻辑
	serviceReq := &svc.GetSaveServiceRequest{
		UserId: req.UserId,
		SaveId: req.SaveId,
	}
	serviceResp, err := svc.Get(ctx, serviceReq)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, &save.GetSaveResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	if serviceResp.Save == nil {
		c.JSON(consts.StatusNotFound, &save.GetSaveResponse{
			Code:    404,
			Message: "保存项不存在",
		})
		return
	}
	// 返回成功响应
	c.JSON(consts.StatusOK, &save.GetSaveResponse{
		Code:    0,
		Message: "获取成功",
		Save:    serviceResp.Save,
	})
}

// 更新保存
func UpdateSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(save.UpdateSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.UpdateSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 调用 service 层更新保存逻辑
	serviceReq := &svc.UpdateSaveServiceRequest{
		UserId:          req.UserId,
		SaveId:          req.SaveId,
		SaveName:        req.SaveName,
		SaveDescription: req.SaveDescription,
		SaveData:        req.SaveData,
	} // 只传递 service 层定义的字段
	_, err := svc.Update(ctx, serviceReq)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, &save.UpdateSaveResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	// 返回成功响应
	c.JSON(consts.StatusOK, &save.UpdateSaveResponse{
		Code:    0,
		Message: "更新成功",
	})
}

// 删除保存
func DeleteSave(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(save.DeleteSaveRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.DeleteSaveResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 调用 service 层删除保存逻辑
	serviceReq := &svc.DeleteSaveServiceRequest{
		UserId: req.UserId,
		SaveId: req.SaveId,
	}
	_, err := svc.Delete(ctx, serviceReq)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, &save.DeleteSaveResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	// 返回成功响应
	c.JSON(consts.StatusOK, &save.DeleteSaveResponse{
		Code:    0,
		Message: "删除成功",
	})
}

// 列出用户保存
func ListSaves(ctx context.Context, c *app.RequestContext) {
	// 获取请求参数
	req := new(save.ListSavesRequest)
	if err := c.BindAndValidate(req); err != nil {
		c.JSON(consts.StatusBadRequest, &save.ListSavesResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 调用 service 层列出保存逻辑
	serviceReq := &svc.ListSavesServiceRequest{
		UserId:   req.UserId,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	} // int32转int
	serviceResp, err := svc.List(ctx, serviceReq)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, &save.ListSavesResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	// 返回成功响应
	c.JSON(consts.StatusOK, &save.ListSavesResponse{
		Code:    0,
		Message: "获取成功",
		Saves:   serviceResp.Saves,
		Total:   int32(serviceResp.Total),
	}) // int转int32
}
