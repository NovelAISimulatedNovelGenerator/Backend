package background

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	mu  sync.Mutex // 互斥锁，用于保护并发操作
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
//
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空且名称必填
// 2. 如果提供了WorldviewId，验证该世界观是否存在
// 3. 如果提供了ParentId，验证父背景信息是否存在
// 4. 验证父背景信息和当前背景信息是否属于同一世界观
// 5. 创建数据库实体并调用数据访问层执行创建操作
// 6. 获取创建的实体并转换为服务层模型返回
//
// 参数:
//   - req: 包含要创建的背景信息详细内容的请求对象
//
// 返回值:
//   - *background.BackgroundInfo: 新创建的背景信息对象，包含数据库分配的ID和时间戳
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空或缺失必要字段时返回InvalidParameterError
//     - 当引用的世界观或父背景信息不存在时返回InvalidParameterError
//     - 当父背景信息与当前背景信息不属于同一世界观时返回InvalidParameterError
//     - 当数据库操作失败时返回相应错误
//
// 注意事项:
//   - Name 字段是必填项，不能为空
//   - WorldviewId 可以为0，表示不属于任何世界观
//   - ParentId 可以为0，表示没有父背景信息
//   - 如果提供了ParentId，则ParentId对应的背景信息与新创建的背景信息必须属于同一世界观下
//   - 创建成功后会获取完整的新实体并返回，如果在二次查询中失败，将返回错误
func (s *BackgroundInfoService) CreateBackgroundInfo(req *background.CreateBackgroundInfoRequest) (*background.BackgroundInfo, error) {
	if req == nil {
		return nil, errno.InvalidParameterError("无效的请求参数")
	}

	// 验证必填字段
	if req.Name == "" {
		return nil, errno.InvalidParameterError("背景信息名称不能为空")
	}

	// 验证 WorldviewID 是否存在
	if req.WorldviewId > 0 {
		worldview, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
		if err != nil {
			if errors.Is(err, db.ErrWorldviewNotFound) {
				return nil, errno.InvalidParameterError("指定的世界观不存在")
			}
			hlog.CtxErrorf(s.ctx, "验证世界观ID %d 时发生错误: %v", req.WorldviewId, err)
			return nil, err
		}
		if worldview == nil {
			return nil, errno.InvalidParameterError("指定的世界观不存在")
		}
	}

	// 验证父背景信息ID是否存在(如果有)
	if req.ParentId > 0 {
		parent, err := db.GetBackgroundInfoByID(s.ctx, req.ParentId)
		if err != nil {
			if errors.Is(err, db.ErrWorldviewNotFound) {
				return nil, errno.InvalidParameterError("指定的父背景信息不存在")
			}
			hlog.CtxErrorf(s.ctx, "验证父背景信息ID %d 时发生错误: %v", req.ParentId, err)
			return nil, err
		}
		if parent == nil {
			return nil, errno.InvalidParameterError("指定的父背景信息不存在")
		}

		// 验证父背景信息是否属于同一世界观
		if parent.WorldviewID != req.WorldviewId {
			return nil, errno.InvalidParameterError("父背景信息必须属于同一世界观")
		}
	}

	dbBI := &db.BackgroundInfo{
		WorldviewID: req.WorldviewId,
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	id, err := db.CreateBackgroundInfo(s.ctx, dbBI)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "在数据访问层创建背景信息失败: %v, 请求: %+v", err, req)
		return nil, errno.DatabaseError(fmt.Sprintf("创建背景信息失败: %v", err))
	}

	// 创建成功后，获取完整信息并返回
	createdBI, err := db.GetBackgroundInfoByID(s.ctx, id)
	if err != nil {
		hlog.CtxWarnf(s.ctx, "创建后获取ID为%d的背景信息失败: %v", id, err)
		// 即使获取失败，创建本身是成功的，但应该返回错误以保证接口一致性
		return nil, errno.DatabaseError(fmt.Sprintf("获取新创建的背景信息失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "成功创建背景信息，ID: %d", id)
	return convertDBBackgroundInfoToModel(createdBI), nil
}

// GetBackgroundInfoByID 通过ID获取单个背景信息
//
// 工作流程:
// 1. 验证背景信息ID参数的有效性，确保是正数
// 2. 调用数据访问层的 GetBackgroundInfoByID 方法获取背景信息
// 3. 处理可能出现的错误情况，包括记录不存在和数据库错误
// 4. 将数据库实体转换为服务层模型并返回
//
// 参数:
//   - req: 包含要查询的背景信息ID的请求对象
//
// 返回值:
//   - *background.BackgroundInfo: 返回找到的背景信息对象，如果没有找到则为 nil
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回BackgroundInfoNotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 请求参数中的BackgroundId必须为正数
//   - 当记录不存在时，返回的错误使用项目定义的errno.BackgroundInfoNotFoundError而非nil
func (s *BackgroundInfoService) GetBackgroundInfoByID(req *background.GetBackgroundInfoRequest) (*background.BackgroundInfo, error) {
	hlog.CtxInfof(s.ctx, "GetBackgroundInfoByID 被调用，ID: %d", req.BackgroundId)

	if req == nil || req.BackgroundId <= 0 {
		hlog.CtxWarnf(s.ctx, "请求了无效的背景信息ID: %d", req.BackgroundId)
		return nil, errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}

	dbBackgroundInfo, err := db.GetBackgroundInfoByID(s.ctx, req.BackgroundId)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "从DAL获取ID为%d的背景信息时出错: %v", req.BackgroundId, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			return nil, errno.NotFoundError("背景信息")
		}
		return nil, errno.DatabaseError(fmt.Sprintf("获取背景信息失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "成功获取ID为%d的背景信息", dbBackgroundInfo.ID)
	return convertDBBackgroundInfoToModel(dbBackgroundInfo), nil
}

// UpdateBackgroundInfo 更新现有的背景信息
// 
// 工作流程:
// 1. 验证请求参数的有效性，确保ID为正数
// 2. 构建需要更新的字段映射，仅包含请求中提供的字段
//    - WorldviewId: 当不等于-1时更新，-1表示不更新此字段
//    - Name: 非空字符串时更新
//    - Description: 非空字符串时更新
//    - Tag: 始终更新，允许更新为空字符串
//    - ParentId: 始终更新，允许设置为0
// 3. 如果没有需要更新的字段，返回参数错误
// 4. 调用数据访问层执行更新操作
// 5. 处理可能的错误情况，包括记录不存在和数据库错误
//
// 参数:
//   - req: 包含背景信息ID和需要更新的字段的请求对象
//
// 返回值:
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当没有提供更新字段时返回InvalidParameterError
//     - 当记录不存在时返回BackgroundInfoNotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - WorldviewId字段使用-1作为特殊值表示不更新此字段
//   - 空字符串的Name和Description不会被更新，保留原值
//   - Tag允许更新为空字符串
//   - ParentId允许更新为0，表示无父节点
//   - 当前实现不会验证WorldviewId和ParentId的有效性
func (s *BackgroundInfoService) UpdateBackgroundInfo(req *background.UpdateBackgroundInfoRequest) (*background.BackgroundInfo, error) {
	hlog.CtxInfof(s.ctx, "UpdateBackgroundInfo 被调用，ID: %d", req.Id)

	if req == nil || req.Id <= 0 {
		hlog.CtxWarnf(s.ctx, "用于更新的背景信息ID无效: %d", req.Id)
		return nil, errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}
	
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 验证背景信息是否存在
	origBI, err := db.GetBackgroundInfoByID(s.ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			hlog.CtxWarnf(s.ctx, "更新背景信息失败: 未找到ID为%d的背景信息", req.Id)
			return nil, errno.NotFoundError("背景信息")
		}
		hlog.CtxErrorf(s.ctx, "更新前检查背景信息失败: %v", err)
		return nil, errno.DatabaseError(fmt.Sprintf("验证背景信息失败: %v", err))
	}

	// 检查依赖
	// 如果更新世界观ID，验证新世界观是否存在
	if req.WorldviewId > 0 && req.WorldviewId != origBI.WorldviewID {
		_, err = db.GetWorldviewByID(s.ctx, req.WorldviewId)
		if err != nil {
			if errors.Is(err, db.ErrWorldviewNotFound) {
				hlog.CtxWarnf(s.ctx, "UpdateBackgroundInfo: 指定的新世界观ID %d 不存在", req.WorldviewId)
				return nil, errno.InvalidParameterError("指定的新世界观不存在")
			}
			hlog.CtxErrorf(s.ctx, "UpdateBackgroundInfo: 验证新世界观ID %d 时发生错误: %v", req.WorldviewId, err)
			return nil, errno.DatabaseError(fmt.Sprintf("验证新世界观失败: %v", err))
		}
	}
	
	// 如果更新父背景信息ID，验证父背景信息是否存在且属于同一世界观
	if req.ParentId > 0 && req.ParentId != origBI.ParentID {
		parent, err := db.GetBackgroundInfoByID(s.ctx, req.ParentId)
		if err != nil {
			if errors.Is(err, db.ErrBackgroundInfoNotFound) {
				hlog.CtxWarnf(s.ctx, "UpdateBackgroundInfo: 指定的父背景信息ID %d 不存在", req.ParentId)
				return nil, errno.InvalidParameterError("指定的父背景信息不存在")
			}
			hlog.CtxErrorf(s.ctx, "UpdateBackgroundInfo: 验证父背景信息ID %d 时发生错误: %v", req.ParentId, err)
			return nil, errno.DatabaseError(fmt.Sprintf("验证父背景信息失败: %v", err))
		}
		
		// 父背景信息世界观ID与当前背景信息世界观ID必须相同
		effectiveWorldviewID := origBI.WorldviewID
		if req.WorldviewId != -1 {
			effectiveWorldviewID = req.WorldviewId
		}
		if parent.WorldviewID != effectiveWorldviewID {
			hlog.CtxWarnf(s.ctx, "UpdateBackgroundInfo: 父背景信息(世界观ID=%d)与当前背景信息(世界观ID=%d)不属于同一世界观", parent.WorldviewID, effectiveWorldviewID)
			return nil, errno.InvalidParameterError("父背景信息必须属于同一世界观")
		}
	}
	
	// 构造要更新的字段映射
	updates := make(map[string]interface{})
	// 使用 != -1 作为判断标准，允许更新为0
	if req.WorldviewId != -1 {
		updates["worldview_id"] = req.WorldviewId
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// Tag可以更新为空字符串
	updates["tag"] = req.Tag
	// ParentId可以更新为0
	if req.ParentId != -1 {
		updates["parent_id"] = req.ParentId
	}

	if len(updates) == 0 {
		hlog.CtxInfof(s.ctx, "ID为%d的背景信息没有需要更新的字段", req.Id)
		// 返回当前对象
		return convertDBBackgroundInfoToModel(origBI), nil
	}

	err = db.UpdateBackgroundInfo(s.ctx, req.Id, updates)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "在DAL中更新ID为%d的背景信息时出错: %v", req.Id, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			return nil, errno.NotFoundError("背景信息")
		}
		return nil, errno.DatabaseError(fmt.Sprintf("更新背景信息失败: %v", err))
	}

	// 获取更新后的数据
	dbBIUpdated, err := db.GetBackgroundInfoByID(s.ctx, req.Id)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "获取更新后ID为%d的背景信息失败: %v", req.Id, err)
		return nil, errno.DatabaseError(fmt.Sprintf("获取更新后的背景信息失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "成功更新ID为%d的背景信息", req.Id)
	return convertDBBackgroundInfoToModel(dbBIUpdated), nil
}

// DeleteBackgroundInfo 通过ID删除背景信息
//
// 工作流程:
// 1. 验证背景信息ID参数的有效性，确保是正数
// 2. 调用数据访问层的 DeleteBackgroundInfo 方法删除背景信息
// 3. 处理可能出现的错误情况，包括记录不存在和数据库错误
//
// 参数:
//   - req: 包含要删除的背景信息ID的请求对象
//
// 返回值:
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回BackgroundInfoNotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 请求参数中的BackgroundId必须为正数
//   - 当应删除的记录不存在时，返回 BackgroundInfoNotFoundError 错误，而不是返回成功
//   - 当前实现不检查该背景信息是否有子节点，删除可能会导致父子关系数据不一致
func (s *BackgroundInfoService) DeleteBackgroundInfo(req *background.DeleteBackgroundInfoRequest) error {
	hlog.CtxInfof(s.ctx, "DeleteBackgroundInfo 被调用，ID: %d", req.BackgroundId)

	if req == nil || req.BackgroundId <= 0 {
		hlog.CtxWarnf(s.ctx, "用于删除的背景信息ID无效: %d", req.BackgroundId)
		return errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}
	
	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 验证背景信息是否存在
	_, err := db.GetBackgroundInfoByID(s.ctx, req.BackgroundId)
	if err != nil {
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			hlog.CtxWarnf(s.ctx, "删除背景信息失败: 未找到ID为%d的背景信息", req.BackgroundId)
			return errno.NotFoundError("背景信息")
		}
		hlog.CtxErrorf(s.ctx, "删除前获取背景信息失败: %v", err)
		return errno.DatabaseError(fmt.Sprintf("验证背景信息失败: %v", err))
	}
	
	// 检查是否有子背景信息依赖于该背景信息
	// 可以添加检查逻辑确保数据完整性

	err = db.DeleteBackgroundInfo(s.ctx, req.BackgroundId)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "从DAL删除ID为%d的背景信息时出错: %v", req.BackgroundId, err)
		if errors.Is(err, db.ErrBackgroundInfoNotFound) {
			return errno.NotFoundError("背景信息")
		}
		return errno.DatabaseError(fmt.Sprintf("删除背景信息失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "成功删除ID为%d的背景信息", req.BackgroundId)
	return nil
}

// ListBackgroundInfos 列出背景信息
//
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空
// 2. 处理分页参数，使用请求中的page和pageSize或默认值
// 3. 处理筛选参数，包括WorldviewIdFilter、ParentIdFilter和TagFilter
// 4. 调用数据访问层的ListBackgroundInfos方法获取列表和总数
// 5. 将数据库实体列表转换为服务层模型列表
// 6. 返回转换后的背景信息列表和总数
//
// 参数:
//   - req: 包含查询参数的请求对象，支持分页和多种筛选条件
//
// 返回值:
//   - []*background.BackgroundInfo: 背景信息列表，符合筛选条件和分页范围
//   - int64: 符合筛选条件的记录总数，不受分页限制
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空时返回InvalidParameterError
//     - 当数据库操作失败时返回相应错误
//
// 注意事项:
//   - ParentIdFilter特殊规则：值为-1表示不根据parent_id筛选；值为0或未设置表示筛选parent_id=0的记录
//   - 请求中的page和pageSize如果未设置或无效，会使用默认值
//   - 当没有符合条件的记录时，返回空列表而非nil，总数为0，错误为nil
func (s *BackgroundInfoService) ListBackgroundInfos(req *background.ListBackgroundInfosRequest) ([]*background.BackgroundInfo, int64, error) {
	if req == nil {
		hlog.CtxErrorf(s.ctx, "ListBackgroundInfos: 请求为空")
		return nil, 0, errno.InvalidParameterError("请求不能为空")
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
		hlog.CtxErrorf(s.ctx, "从DAL获取背景信息列表失败: %v, 请求: %+v", err, req)
		return nil, 0, errno.DatabaseError(fmt.Sprintf("获取背景信息列表失败: %v", err))
	}

	bIs := make([]*background.BackgroundInfo, 0, len(dbBIs))
	for _, dbBI := range dbBIs {
		bIs = append(bIs, convertDBBackgroundInfoToModel(&dbBI))
	}

	hlog.CtxInfof(s.ctx, "成功列出%d个背景信息，总数: %d", len(bIs), total)
	return bIs, total, nil
}
