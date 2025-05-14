package background

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"novelai/biz/dal/db"
	"novelai/biz/model/background"
	"novelai/pkg/errno"
)

// BackgroundInfoService 负责处理 BackgroundInfo 相关的业务逻辑
type BackgroundInfoService struct {
	ctx context.Context
	req *app.RequestContext
}

// NewBackgroundInfoService 创建一个新的 BackgroundInfoService 实例
func NewBackgroundInfoService(ctx context.Context, req *app.RequestContext) *BackgroundInfoService {
	return &BackgroundInfoService{
		ctx: ctx,
		req: req,
	}
}

// convertDBBackgroundInfoToModel 将数据库模型转换为服务层模型
func convertDBBackgroundInfoToModel(dbBI *db.BackgroundInfo) *background.BackgroundInfo {
	if dbBI == nil {
		return nil
	}
	return &background.BackgroundInfo{
		Id:          dbBI.ID,
		WorldviewId: dbBI.WorldviewID,
		Name:        dbBI.Name,
		Description: dbBI.Description,
		Tag:         dbBI.Tag,
		ParentId:    dbBI.ParentID,
		CreatedAt:   dbBI.CreatedAt,
		UpdatedAt:   dbBI.UpdatedAt,
	}
}

// CreateBackgroundInfo 创建新的背景信息
func (s *BackgroundInfoService) CreateBackgroundInfo(req *background.CreateBackgroundInfoRequest) (*background.BackgroundInfo, error) {
	if req == nil {
		return nil, errno.InvalidParameterError("无效的请求参数")
	}

	// TODO: 可以添加更多验证逻辑，例如 WorldviewID 是否存在

	dbBI := &db.BackgroundInfo{
		WorldviewID: req.WorldviewId,
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	id, err := db.CreateBackgroundInfo(s.ctx, dbBI)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Failed to create background info in DAL: %v, request: %+v", err, req)
		return nil, err // 返回原始 DAL 错误
	}

	// 创建成功后，获取完整信息并返回
	createdBI, err := db.GetBackgroundInfoByID(s.ctx, id)
	if err != nil {
		hlog.CtxWarnf(s.ctx, "Failed to get created background info by ID %d after creation: %v", id, err)
		// 即使获取失败，创建本身是成功的，可以考虑返回基础信息或者 nil
		// 这里选择返回 nil 和错误，因为无法提供完整的响应模型
		return nil, err
	}

	hlog.CtxInfof(s.ctx, "Successfully created background info with ID: %d", id)
	return convertDBBackgroundInfoToModel(createdBI), nil
}

// GetBackgroundInfoByID retrieves a single BackgroundInfo by its ID.
func (s *BackgroundInfoService) GetBackgroundInfoByID(req *background.GetBackgroundInfoRequest) (*background.BackgroundInfo, error) {
	hlog.CtxInfof(s.ctx, "GetBackgroundInfoByID called with ID: %d", req.BackgroundId)

	if req.BackgroundId <= 0 {
		hlog.CtxWarnf(s.ctx, "Invalid BackgroundInfo ID requested: %d", req.BackgroundId)
		return nil, errno.InvalidParameterError("BackgroundInfo ID must be positive")
	}

	dbBackgroundInfo, err := db.GetBackgroundInfoByID(s.ctx, req.BackgroundId)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Error retrieving BackgroundInfo with ID %d from DAL: %v", req.BackgroundId, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			return nil, errno.BackgroundInfoNotFoundError
		}
		return nil, errno.DatabaseError("Failed to get BackgroundInfo")
	}

	hlog.CtxInfof(s.ctx, "Successfully retrieved BackgroundInfo with ID: %d", dbBackgroundInfo.ID)
	return convertDBBackgroundInfoToModel(dbBackgroundInfo), nil
}

// UpdateBackgroundInfo updates an existing BackgroundInfo.
func (s *BackgroundInfoService) UpdateBackgroundInfo(req *background.UpdateBackgroundInfoRequest) error {
	hlog.CtxInfof(s.ctx, "UpdateBackgroundInfo called for ID: %d", req.Id)

	if req.Id <= 0 {
		hlog.CtxWarnf(s.ctx, "Invalid BackgroundInfo ID for update: %d", req.Id)
		return errno.InvalidParameterError("BackgroundInfo ID must be positive for update")
	}

	// Construct the map of fields to update.
	// Assumes DAL handles zero/empty values appropriately (e.g., skips update for them).
	updates := make(map[string]interface{})
	// We no longer use FieldMask. Include fields if they are present in the request.
	// The presence check logic might need adjustment based on how optional fields are handled (e.g., pointers vs value types).
	// For simple value types, we might update even if the value is the zero value (e.g., updating ParentId to 0).
	// This assumes the request struct directly holds the values to update.
	if req.WorldviewId > 0 { // Assuming WorldviewId 0 is invalid or not meant for update here
		updates["worldview_id"] = req.WorldviewId
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// Tag can potentially be updated to an empty string
	updates["tag"] = req.Tag
	// ParentId can be updated to 0
	updates["parent_id"] = req.ParentId

	if len(updates) == 0 {
		hlog.CtxWarnf(s.ctx, "UpdateBackgroundInfo called for ID %d with no fields to update", req.Id)
		return errno.InvalidParameterError("No fields provided for update")
	}

	err := db.UpdateBackgroundInfo(s.ctx, req.Id, updates)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Error updating BackgroundInfo with ID %d in DAL: %v", req.Id, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			return errno.BackgroundInfoNotFoundError
		}
		return errno.DatabaseError("Failed to update BackgroundInfo")
	}

	hlog.CtxInfof(s.ctx, "Successfully updated BackgroundInfo with ID: %d", req.Id)
	return nil
}

// DeleteBackgroundInfo deletes a BackgroundInfo by its ID.
func (s *BackgroundInfoService) DeleteBackgroundInfo(req *background.DeleteBackgroundInfoRequest) error {
	hlog.CtxInfof(s.ctx, "DeleteBackgroundInfo called for ID: %d", req.BackgroundId)

	if req.BackgroundId <= 0 {
		hlog.CtxWarnf(s.ctx, "Invalid BackgroundInfo ID for deletion: %d", req.BackgroundId)
		return errno.InvalidParameterError("BackgroundInfo ID must be positive for deletion")
	}

	err := db.DeleteBackgroundInfo(s.ctx, req.BackgroundId)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Error deleting BackgroundInfo with ID %d from DAL: %v", req.BackgroundId, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			// Consider if deleting a non-existent record is an error or idempotent.
			// Returning NotFoundError seems appropriate.
			return errno.BackgroundInfoNotFoundError
		}
		return errno.DatabaseError("Failed to delete BackgroundInfo")
	}

	hlog.CtxInfof(s.ctx, "Successfully deleted BackgroundInfo with ID: %d", req.BackgroundId)
	return nil
}

// ListBackgroundInfos 列出背景信息
func (s *BackgroundInfoService) ListBackgroundInfos(req *background.ListBackgroundInfosRequest) ([]*background.BackgroundInfo, int64, error) {
	if req == nil {
		return nil, 0, errno.InvalidParameterError("无效的请求参数")
	}

	page := req.GetPage()
	pageSize := req.GetPageSize()

	// 直接传递来自请求的 ParentIdFilter。
	// DAL 层约定: parentIDFilter = -1 表示不根据 parent_id 筛选。
	// 如果客户端希望不按 parent_id 筛选，应显式传递 -1。
	// 如果客户端传递 0 或省略该字段 (默认为 0)，将筛选 parent_id = 0 的记录。
	parentIDFilter := req.ParentIdFilter

	dbBIs, total, err := db.ListBackgroundInfos(
		s.ctx,
		req.WorldviewIdFilter,
		parentIDFilter,
		req.TagFilter,
		page,
		pageSize,
	)

	if err != nil {
		hlog.CtxErrorf(s.ctx, "Failed to list background infos from DAL: %v, request: %+v", err, req)
		return nil, 0, err
	}

	bIs := make([]*background.BackgroundInfo, 0, len(dbBIs))
	for _, dbBI := range dbBIs {
		bIs = append(bIs, convertDBBackgroundInfoToModel(&dbBI))
	}

	hlog.CtxInfof(s.ctx, "Successfully listed %d background infos, total: %d", len(bIs), total)
	return bIs, total, nil
}
