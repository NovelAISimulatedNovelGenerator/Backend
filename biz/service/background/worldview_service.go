package background

import (
	"context"
	"errors"
	"fmt"

	"novelai/biz/dal/db"
	"novelai/biz/model/background"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// WorldviewService 负责处理世界观相关的业务逻辑
type WorldviewService struct {
	ctx context.Context
	c   *app.RequestContext
}

// NewWorldviewService 创建 WorldviewService 实例
// 参数:
//   - ctx: 上下文
//   - c: Hertz 请求上下文
//
// 返回:
//   - *WorldviewService: WorldviewService 实例
func NewWorldviewService(ctx context.Context, c *app.RequestContext) *WorldviewService {
	return &WorldviewService{ctx: ctx, c: c}
}

// convertDBWorldviewToModel 将 DAL 层 Worldview 结构转换为 API 模型结构
func convertDBWorldviewToModel(dbWv *db.Worldview) *background.Worldview {
	if dbWv == nil {
		return nil
	}
	return &background.Worldview{
		Id:          dbWv.ID,
		Name:        dbWv.Name,
		Description: dbWv.Description,
		Tag:         dbWv.Tag,
		ParentId:    dbWv.ParentID,
		CreatedAt:   dbWv.CreatedAt,
		UpdatedAt:   dbWv.UpdatedAt,
	}
}

// CreateWorldview 创建新的世界观
// 参数:
//   - req: 创建世界观的请求参数，包含名称、描述、标签、父ID等
//
// 返回:
//   - *background.Worldview: 创建成功后的世界观信息
//   - error: 操作错误信息
func (s *WorldviewService) CreateWorldview(req *background.CreateWorldviewRequest) (*background.Worldview, error) {
	if req == nil {
		err := errors.New("CreateWorldview request cannot be nil")
		hlog.CtxErrorf(s.ctx, "CreateWorldview failed: %v", err)
		return nil, err
	}

	dbWv := &db.Worldview{
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	// 调用 DAL 层创建世界观
	// GORM 的 Create 会自动填充 ID, CreatedAt, UpdatedAt
	_, err := db.CreateWorldview(s.ctx, dbWv)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Failed to create worldview in DAL: %v. Request: %+v", err, req)
		return nil, err // 直接返回 DAL 层的错误，可能需要包装
	}

	// dbWv 已经被 GORM 更新了 ID, CreatedAt, UpdatedAt
	return convertDBWorldviewToModel(dbWv), nil
}

// GetWorldviewByID 根据 ID 获取世界观信息
// 参数:
//   - req: 获取世界观的请求参数，包含世界观 ID
//
// 返回:
//   - *background.Worldview: 世界观信息
//   - error: 操作错误信息
func (s *WorldviewService) GetWorldviewByID(req *background.GetWorldviewRequest) (*background.Worldview, error) {
	if req == nil || req.WorldviewId == 0 {
		err := errors.New("GetWorldviewByID request cannot be nil or ID cannot be zero")
		hlog.CtxErrorf(s.ctx, "GetWorldviewByID failed: %v", err)
		return nil, err
	}

	dbWv, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "Worldview not found for ID %d: %v", req.WorldviewId, err)
			return nil, err // 可以自定义更友好的错误信息或直接透传
		}
		hlog.CtxErrorf(s.ctx, "Failed to get worldview by ID %d from DAL: %v", req.WorldviewId, err)
		return nil, err
	}

	return convertDBWorldviewToModel(dbWv), nil
}

// UpdateWorldview 更新世界观信息
// 参数:
//   - req: 更新世界观的请求参数，包含世界观 ID 及需要更新的字段
//
// 返回:
//   - error: 操作错误信息
func (s *WorldviewService) UpdateWorldview(req *background.UpdateWorldviewRequest) error {
	if req == nil {
		err := errors.New("UpdateWorldview request cannot be nil")
		hlog.CtxErrorf(s.ctx, "UpdateWorldview failed: %v", err)
		return err
	}

	// 检查世界观是否存在
	_, err := db.GetWorldviewByID(s.ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "UpdateWorldview failed: worldview with ID %d not found. Error: %v", req.Id, err)
		}
		return err // 返回原始错误，调用方可以判断是否是 NotFound
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}
	if req.ParentId != 0 {
		updates["parent_id"] = req.ParentId
	}

	if len(updates) == 0 {
		hlog.CtxInfof(s.ctx, "No fields to update for worldview ID %d", req.Id)
		return nil // 没有需要更新的字段
	}

	// DAL 层的 UpdateWorldview 会自动处理 updated_at
	err = db.UpdateWorldview(s.ctx, req.Id, updates)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "Failed to update worldview ID %d in DAL: %v. Updates: %+v", req.Id, err, updates)
		return err
	}

	hlog.CtxInfof(s.ctx, "Worldview ID %d updated successfully.", req.Id)
	return nil
}

// DeleteWorldview 删除世界观
// 参数:
//   - req: 删除世界观的请求参数，包含世界观 ID
//
// 返回:
//   - error: 操作错误信息
func (s *WorldviewService) DeleteWorldview(req *background.DeleteWorldviewRequest) error {
	if req == nil || req.WorldviewId == 0 {
		hlog.CtxWarnf(s.ctx, "DeleteWorldview: invalid request, req is nil or WorldviewId is 0")
		return errors.New("invalid request")
	}

	err := db.DeleteWorldview(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "DeleteWorldview failed: worldview with ID %d not found. Error: %v", req.WorldviewId, err)
		}
		hlog.CtxErrorf(s.ctx, "Failed to delete worldview ID %d from DAL: %v", req.WorldviewId, err)
		return err
	}

	hlog.CtxInfof(s.ctx, "Worldview ID %d deleted successfully.", req.WorldviewId)
	return nil
}

// ListWorldviews 列出世界观，支持分页和过滤
// 参数:
//   - req: 列出世界观的请求参数，包含过滤条件和分页参数
//
// 返回:
//   - []*background.Worldview: 世界观列表
//   - int64: 总记录数
//   - error: 操作错误信息
func (s *WorldviewService) ListWorldviews(req *background.ListWorldviewsRequest) ([]*background.Worldview, int64, error) {
	if req == nil {
		err := errors.New("ListWorldviews request cannot be nil")
		hlog.CtxErrorf(s.ctx, "ListWorldviews failed: %v", err)
		return nil, 0, err
	}

	page := req.GetPage()
	if page <= 0 {
		page = 1 // Default page is 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10 // Default page size is 10
	}

	// Assuming the request struct doesn't support ParentId filtering.
	// Pass -1 to DAL assuming it means 'no parent filter'. Verify this assumption.
	parentID := int64(-1)
	// Revert to using GetTagFilter(), assuming this getter exists.
	tagFilter := req.GetTagFilter()

	// 调用 DAL 层获取世界观列表和总数
	dbWorldviews, total, err := db.ListWorldviews(s.ctx, parentID, tagFilter, int(page), int(pageSize)) // Pass params directly
	if err != nil {
		hlog.CtxErrorf(s.ctx, "ListWorldviews: failed to list worldviews from DAL: %v", err)
		return nil, 0, fmt.Errorf("failed to list worldviews: %w", err)
	}

	modelWorldviews := make([]*background.Worldview, 0, len(dbWorldviews))
	for _, dbWv := range dbWorldviews {
		modelWorldviews = append(modelWorldviews, convertDBWorldviewToModel(&dbWv))
	}

	return modelWorldviews, total, nil
}
