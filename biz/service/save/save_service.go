// save_service.go 封装保存相关所有业务逻辑（创建、获取、更新、删除、列表），参数为结构体，返回值为结果和 error
// 变量作用域最小化，函数无递归，圈复杂度低于 10，注释完整，便于单元测试
package save

import (
	"context"
	"errors"
	"fmt"
	db "novelai/biz/dal/db"
	"novelai/biz/model/save"
	"time"
)

// CreateSaveServiceRequest 创建保存业务参数
// 包含用户ID、保存名称、描述、数据等
// 仅用于 service 层，便于扩展和单元测试
type CreateSaveServiceRequest struct {
	UserId          int64  // 用户ID
	SaveName        string // 保存名称
	SaveDescription string // 保存描述
	SaveData        string // 保存数据
	SaveType        string // 保存类型
}

// CreateSaveServiceResponse 创建保存业务返回值
// 包含保存ID等信息
// 仅用于 service 层
type CreateSaveServiceResponse struct {
	SaveId string // 保存ID
}

// Create 创建保存业务逻辑，返回保存ID和错误
// ctx: 上下文，req: 创建请求参数
// 返回: 创建结果和错误
// Create 创建保存业务逻辑，返回保存ID和错误
// ctx: 上下文，req: 创建请求参数
// 返回: 创建结果和错误
func Create(ctx context.Context, req *CreateSaveServiceRequest) (*CreateSaveServiceResponse, error) {
	// 参数校验
	if req == nil {
		return nil, ErrInvalidRequest
	}
	if req.UserId <= 0 || req.SaveName == "" || req.SaveData == "" || req.SaveType == "" {
		return nil, ErrInvalidRequest
	}
	// 构造 db.Save
	dbSave := &db.Save{
		UserID:          req.UserId,
		SaveID:          generateSaveID(req.UserId), // 生成唯一ID
		SaveName:        req.SaveName,
		SaveDescription: req.SaveDescription,
		SaveData:        req.SaveData,
		SaveType:        req.SaveType,
		SaveStatus:      "active",
		CreatedAt:       nowUnix(),
		UpdatedAt:       nowUnix(),
	}
	_, err := db.CreateSave(dbSave)
	if err != nil {
		return nil, err
	}
	return &CreateSaveServiceResponse{SaveId: dbSave.SaveID}, nil
}

// generateSaveID 生成唯一的保存ID（可根据实际需求替换为更复杂算法）
func generateSaveID(userID int64) string {
	return fmt.Sprintf("save-%d-%d", userID, nowUnixNano())
}

// nowUnix 获取当前unix时间戳
func nowUnix() int64 {
	return time.Now().Unix()
}

// nowUnixNano 获取当前纳秒时间戳
func nowUnixNano() int64 {
	return time.Now().UnixNano()
}

// ErrInvalidRequest 非法参数错误
var ErrInvalidRequest = errors.New("请求参数不合法")

// GetSaveServiceRequest 获取保存业务参数
// 包含用户ID、保存ID
// 仅用于 service 层，便于扩展和单元测试
type GetSaveServiceRequest struct {
	UserId int64  // 用户ID
	SaveId string // 保存ID
}

// GetSaveServiceResponse 获取保存业务返回值
// 包含保存项详细信息
// 仅用于 service 层
type GetSaveServiceResponse struct {
	Save *save.Save // 保存项
}

// Get 获取保存业务逻辑，返回保存项和错误
// ctx: 上下文，req: 获取请求参数
// 返回: 获取结果和错误
// Get 获取保存业务逻辑，返回保存项和错误
// ctx: 上下文，req: 获取请求参数
// 返回: 获取结果和错误
func Get(ctx context.Context, req *GetSaveServiceRequest) (*GetSaveServiceResponse, error) {
	if req == nil || req.UserId <= 0 || req.SaveId == "" {
		return nil, ErrInvalidRequest
	}
	dbSave, err := querySaveBySaveID(req.SaveId)
	if err != nil {
		return nil, err
	}
	if dbSave.UserID != req.UserId {
		return nil, db.ErrSaveNotFound
	}
	// 转换为 model/save.Save
	modelSave := &save.Save{
		Id:              dbSave.ID,
		UserId:          dbSave.UserID,
		SaveId:          dbSave.SaveID,
		SaveName:        dbSave.SaveName,
		SaveDescription: dbSave.SaveDescription,
		SaveData:        dbSave.SaveData,
		SaveType:        dbSave.SaveType,
		SaveStatus:      dbSave.SaveStatus,
		CreatedAt:       dbSave.CreatedAt,
		UpdatedAt:       dbSave.UpdatedAt,
	}
	return &GetSaveServiceResponse{Save: modelSave}, nil
}

// querySaveBySaveID 通过保存唯一标识符查询存档，直接调用 dal 层接口
func querySaveBySaveID(saveID string) (*db.Save, error) {
	return db.QuerySavesBySaveID(saveID)
}

// UpdateSaveServiceRequest 更新保存业务参数
// 包含用户ID、保存ID、名称、描述、数据等
// 仅用于 service 层，便于扩展和单元测试
type UpdateSaveServiceRequest struct {
	UserId          int64  // 用户ID
	SaveId          string // 保存ID
	SaveName        string // 保存名称
	SaveDescription string // 保存描述
	SaveData        string // 保存数据
	SaveType        string // 保存类型
}

// UpdateSaveServiceResponse 更新保存业务返回值
// 仅用于 service 层
type UpdateSaveServiceResponse struct {
}

// Update 更新保存业务逻辑，返回错误
// ctx: 上下文，req: 更新请求参数
// 返回: 更新结果和错误
// Update 更新保存业务逻辑，返回错误
// ctx: 上下文，req: 更新请求参数
// 返回: 更新结果和错误
func Update(ctx context.Context, req *UpdateSaveServiceRequest) (*UpdateSaveServiceResponse, error) {
	if req == nil || req.UserId <= 0 || req.SaveId == "" {
		return nil, ErrInvalidRequest
	}
	dbSave, err := querySaveBySaveID(req.SaveId)
	if err != nil {
		return nil, err
	}
	if dbSave.UserID != req.UserId {
		return nil, db.ErrSaveNotFound
	}
	dbSave.SaveName = req.SaveName
	dbSave.SaveDescription = req.SaveDescription
	dbSave.SaveData = req.SaveData
	dbSave.SaveType = req.SaveType
	dbSave.UpdatedAt = nowUnix()
	err = db.UpdateSave(dbSave)
	if err != nil {
		return nil, err
	}
	return &UpdateSaveServiceResponse{}, nil
}

// DeleteSaveServiceRequest 删除保存业务参数
// 包含用户ID、保存ID
// 仅用于 service 层，便于扩展和单元测试
type DeleteSaveServiceRequest struct {
	UserId int64  // 用户ID
	SaveId string // 保存ID
}

// DeleteSaveServiceResponse 删除保存业务返回值
// 仅用于 service 层
type DeleteSaveServiceResponse struct {
}

// Delete 删除保存业务逻辑，返回错误
// ctx: 上下文，req: 删除请求参数
// 返回: 删除结果和错误
// Delete 删除保存业务逻辑，返回错误
// ctx: 上下文，req: 删除请求参数
// 返回: 删除结果和错误
func Delete(ctx context.Context, req *DeleteSaveServiceRequest) (*DeleteSaveServiceResponse, error) {
	if req == nil || req.UserId <= 0 || req.SaveId == "" {
		return nil, ErrInvalidRequest
	}
	dbSave, err := querySaveBySaveID(req.SaveId)
	if err != nil {
		return nil, err
	}
	if dbSave.UserID != req.UserId {
		return nil, db.ErrSaveNotFound
	}
	err = db.DeleteSave(dbSave.ID)
	if err != nil {
		return nil, err
	}
	return &DeleteSaveServiceResponse{}, nil
}

// ListSavesServiceRequest 列出保存业务参数
// 包含用户ID、分页参数等
// 仅用于 service 层，便于扩展和单元测试
type ListSavesServiceRequest struct {
	UserId   int64 // 用户ID
	Page     int   // 页码
	PageSize int   // 每页数量
}

// ListSavesServiceResponse 列出保存业务返回值
// 包含保存项列表和总数
// 仅用于 service 层
type ListSavesServiceResponse struct {
	Saves []*save.Save // 保存项列表
	Total int          // 总数
}

// List 列出保存业务逻辑，返回保存项列表和错误
// ctx: 上下文，req: 列出请求参数
// 返回: 列表结果和错误
func List(ctx context.Context, req *ListSavesServiceRequest) (*ListSavesServiceResponse, error) {
	if req == nil || req.UserId <= 0 || req.Page < 1 || req.PageSize < 1 {
		return nil, ErrInvalidRequest
	}
	dbSaves, total, err := db.QuerySavesByUser(req.UserId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	modelSaves := make([]*save.Save, 0, len(dbSaves))
	for _, s := range dbSaves {
		modelSaves = append(modelSaves, &save.Save{
			Id:              s.ID,
			UserId:          s.UserID,
			SaveId:          s.SaveID,
			SaveName:        s.SaveName,
			SaveDescription: s.SaveDescription,
			SaveData:        s.SaveData,
			SaveType:        s.SaveType,
			SaveStatus:      s.SaveStatus,
			CreatedAt:       s.CreatedAt,
			UpdatedAt:       s.UpdatedAt,
		})
	}
	return &ListSavesServiceResponse{Saves: modelSaves, Total: int(total)}, nil
}
