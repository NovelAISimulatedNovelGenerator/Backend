package background

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"novelai/biz/dal/db"
	"novelai/biz/model/background"
	"novelai/pkg/errno"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// WorldviewService 负责处理世界观相关的业务逻辑
type WorldviewService struct {
	ctx context.Context
	c   *app.RequestContext
	mu  sync.Mutex // 用于保护并发操作
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
		err := errno.InvalidParameterError("请求不能为空")
		hlog.CtxErrorf(s.ctx, "创建世界观失败: %v", err)
		return nil, err
	}
	
	// 验证必填字段
	if req.Name == "" {
		err := errno.InvalidParameterError("世界观名称不能为空")
		hlog.CtxErrorf(s.ctx, "创建世界观失败: %v", err)
		return nil, err
	}

	dbWv := &db.Worldview{
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	// 加锁保护数据一致性
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 调用 DAL 层创建世界观
	// GORM 的 Create 会自动填充 ID, CreatedAt, UpdatedAt
	_, err := db.CreateWorldview(s.ctx, dbWv)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "在数据访问层创建世界观失败: %v. 请求参数: %+v", err, req)
		return nil, errno.DatabaseError(fmt.Sprintf("创建世界观失败: %v", err))
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
	if req == nil || req.WorldviewId <= 0 {
		err := errno.InvalidParameterError("请求不能为空或ID必须为正数")
		hlog.CtxErrorf(s.ctx, "通过ID获取世界观失败: %v", err)
		return nil, err
	}

	dbWv, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "未找到ID为%d的世界观: %v", req.WorldviewId, err)
			return nil, errno.NotFoundError("世界观") 
		}
		hlog.CtxErrorf(s.ctx, "从数据访问层获取ID为%d的世界观失败: %v", req.WorldviewId, err)
		return nil, errno.DatabaseError(fmt.Sprintf("获取世界观失败: %v", err))
	}

	return convertDBWorldviewToModel(dbWv), nil
}

// UpdateWorldview 更新世界观信息
// 参数:
//   - req: 更新世界观的请求参数，包含世界观 ID 及需要更新的字段
//
// 返回:
//   - error: 操作错误信息
func (s *WorldviewService) UpdateWorldview(req *background.UpdateWorldviewRequest) (*background.Worldview, error) {
	if req == nil {
		err := errno.InvalidParameterError("请求不能为空")
		hlog.CtxErrorf(s.ctx, "更新世界观失败: %v", err)
		return nil, err
	}
	
	if req.Id <= 0 {
		err := errno.InvalidParameterError("世界观ID必须为正数")
		hlog.CtxErrorf(s.ctx, "更新世界观失败: %v", err)
		return nil, err
	}

	// 加锁保护数据一致性
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 检查世界观是否存在
	_, err := db.GetWorldviewByID(s.ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "更新世界观失败: 未找到ID为%d的世界观. 错误: %v", req.Id, err)
			return nil, errno.NotFoundError("世界观")
		}
		hlog.CtxErrorf(s.ctx, "更新前检查世界观失败: %v", err)
		return nil, errno.DatabaseError(fmt.Sprintf("验证世界观失败: %v", err))
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// Tag允许更新为空字符串
	updates["tag"] = req.Tag
	// 使用 != -1 作为判断标准，允许将ParentId更新为0
	if req.ParentId != -1 {
		updates["parent_id"] = req.ParentId
	}

	if len(updates) == 0 {
		hlog.CtxInfof(s.ctx, "ID为%d的世界观没有需要更新的字段", req.Id)
		// 返回当前对象
		dbWv, _ := db.GetWorldviewByID(s.ctx, req.Id)
		return convertDBWorldviewToModel(dbWv), nil
	}

	// DAL 层的 UpdateWorldview 会自动处理 updated_at
	err = db.UpdateWorldview(s.ctx, req.Id, updates)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "在数据访问层更新ID为%d的世界观失败: %v. 更新内容: %+v", req.Id, err, updates)
		if errors.Is(err, db.ErrWorldviewNotFound) {
			return nil, errno.NotFoundError("世界观")
		}
		return nil, errno.DatabaseError(fmt.Sprintf("更新世界观失败: %v", err))
	}

	// 获取更新后的数据
	dbWvUpdated, err := db.GetWorldviewByID(s.ctx, req.Id)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "获取更新后ID为%d的世界观失败: %v", req.Id, err)
		return nil, errno.DatabaseError(fmt.Sprintf("获取更新后的世界观失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "ID为%d的世界观已成功更新", req.Id)
	return convertDBWorldviewToModel(dbWvUpdated), nil
}

// DeleteWorldview 删除世界观
// 参数:
//   - req: 删除世界观的请求参数，包含世界观 ID
//
// 返回:
//   - error: 操作错误信息
func (s *WorldviewService) DeleteWorldview(req *background.DeleteWorldviewRequest) error {
	if req == nil || req.WorldviewId <= 0 {
		hlog.CtxWarnf(s.ctx, "删除世界观: 无效请求，请求为空或世界观ID为非正数")
		return errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}
	
	// 加锁保护数据一致性
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 检查世界观是否存在
	_, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "删除世界观失败: 未找到ID为%d的世界观", req.WorldviewId)
			return errno.NotFoundError("世界观")
		}
		hlog.CtxErrorf(s.ctx, "删除前检查世界观失败: %v", err)
		return errno.DatabaseError(fmt.Sprintf("验证世界观失败: %v", err))
	}
	
	// 检查是否有依赖此世界观的规则或背景信息
	// 这里可以添加检查关联项的逻辑，当发现有相关规则时可以返回错误
	// 或者选择级联删除

	err = db.DeleteWorldview(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "删除世界观失败: 未找到ID为%d的世界观. 错误: %v", req.WorldviewId, err)
			return errno.NotFoundError("世界观")
		}
		hlog.CtxErrorf(s.ctx, "从数据访问层删除ID为%d的世界观失败: %v", req.WorldviewId, err)
		return errno.DatabaseError(fmt.Sprintf("删除世界观失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "ID为%d的世界观已成功删除", req.WorldviewId)
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
		err := errno.InvalidParameterError("请求不能为空")
		hlog.CtxErrorf(s.ctx, "获取世界观列表失败: %v", err)
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
		hlog.CtxErrorf(s.ctx, "获取世界观列表: 从数据访问层获取世界观列表失败: %v", err)
		return nil, 0, errno.DatabaseError(fmt.Sprintf("获取世界观列表失败: %v", err))
	}

	modelWorldviews := make([]*background.Worldview, 0, len(dbWorldviews))
	for _, dbWv := range dbWorldviews {
		modelWorldviews = append(modelWorldviews, convertDBWorldviewToModel(&dbWv))
	}

	return modelWorldviews, total, nil
}
