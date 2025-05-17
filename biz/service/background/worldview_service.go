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
	ctx context.Context     // 上下文环境，用于数据库操作和日志记录
	c   *app.RequestContext // Hertz 框架的请求上下文
	mu  sync.Mutex          // 互斥锁，用于保护并发操作
}

// NewWorldviewService 创建一个新的 WorldviewService 实例
//
// 工作流程:
// 1. 接收上下文和请求上下文参数
// 2. 创建并返回 WorldviewService 实例，包含所需的上下文信息
//
// 参数:
//   - ctx: 业务上下文，用于日志记录和数据库操作的上下文传递
//   - c: Hertz 框架的请求上下文，包含当前 HTTP 请求的相关信息
//
// 返回值:
//   - *WorldviewService: 新创建的 WorldviewService 实例，可用于处理世界观相关的业务逻辑
//
// 注意事项:
//   - 该函数是工厂方法，每个请求都应创建新的 WorldviewService 实例
//   - 返回的服务实例包含互斥锁，可以安全地在并发环境中使用
func NewWorldviewService(ctx context.Context, c *app.RequestContext) *WorldviewService {
	return &WorldviewService{ctx: ctx, c: c}
}

// convertDBWorldviewToModel 将数据库模型转换为服务层模型
//
// 工作流程:
// 1. 检查输入的数据库模型指针是否为 nil
// 2. 将数据库模型的字段映射到服务层模型的对应字段
//
// 参数:
//   - dbWv: 数据库层的世界观实体指针，包含从数据库获取的完整世界观信息
//
// 返回值:
//   - *background.Worldview: 转换后的服务层世界观实体，适用于 API 响应
//   - 当输入为 nil 时返回 nil
//
// 注意事项:
//   - 该函数是纯函数，不依赖服务实例的状态
//   - 实现了 DAO 层和服务层之间的数据模型转换，解耦数据访问和业务逻辑
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
//
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空且名称必填
// 2. 创建数据库实体并设置相关字段
// 3. 加锁保护并发安全
// 4. 调用数据访问层的 CreateWorldview 方法执行创建操作
// 5. 返回创建成功的世界观实体，包含数据库生成的 ID 和时间戳
//
// 参数:
//   - req: 包含要创建的世界观详细内容的请求对象
//
// 返回值:
//   - *background.Worldview: 新创建的世界观对象，包含数据库分配的ID和时间戳
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空或缺失必要字段时返回InvalidParameterError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - Name 字段是必填项，不能为空
//   - ParentId 可以为0，表示没有父世界观
//   - 创建成功后直接返回完整的世界观实体，包含 GORM 自动填充的 ID, CreatedAt, UpdatedAt
//   - 使用互斥锁保护创建操作，确保数据一致性
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

// GetWorldviewByID 通过ID获取单个世界观
//
// 工作流程:
// 1. 验证世界观ID参数的有效性，确保是正数
// 2. 调用数据访问层的 GetWorldviewByID 方法获取世界观
// 3. 处理可能出现的错误情况，包括记录不存在和数据库错误
// 4. 将数据库实体转换为服务层模型并返回
//
// 参数:
//   - req: 包含要查询的世界观ID的请求对象
//
// 返回值:
//   - *background.Worldview: 返回找到的世界观对象，如果没有找到则为 nil
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回NotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 请求参数中的WorldviewId必须为正数
//   - 当记录不存在时，返回的错误使用项目定义的errno.NotFoundError而非nil
//   - 该方法不需要加锁，因为它只执行只读操作
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

// UpdateWorldview 更新现有的世界观
// 
// 工作流程:
// 1. 验证请求参数的有效性，确保ID为正数
// 2. 构建需要更新的字段映射，仅包含请求中提供的字段
//    - Name: 非空字符串时更新
//    - Description: 非空字符串时更新
//    - Tag: 始终更新，允许更新为空字符串
//    - ParentId: 当不等于-1时更新，允许更新为0
// 3. 调用数据访问层更新世界观信息
// 4. 获取更新后的实体并返回
//
// 参数:
//   - req: 包含世界观ID和需要更新的字段的请求对象
//
// 返回值:
//   - *background.Worldview: 更新后的世界观对象，包含最新的信息
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空或ID无效时返回InvalidParameterError
//     - 当世界观不存在时返回NotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 当没有任何字段需要更新时，会直接返回当前世界观对象
//   - Name, Description 仅在非空时才会更新
//   - Tag 始终会被更新，包括更新为空字符串
//   - ParentId 不等于-1时才会更新，允许更新为0(表示没有父世界观)
//   - 更新后会重新查询数据库获取更新后的完整信息
//   - 使用互斥锁保护更新操作，确保数据一致性
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

// DeleteWorldview 通过ID删除世界观
// 
// 工作流程:
// 1. 验证世界观ID参数的有效性，确保是正数
// 2. 检查要删除的世界观是否存在
// 3. 调用数据访问层的 DeleteWorldview 方法删除世界观
// 4. 处理可能出现的错误情况，包括记录不存在和数据库错误
// 
// 参数:
//   - req: 包含要删除的世界观ID的请求对象
// 
// 返回值:
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回NotFoundError
//     - 当数据库操作失败时返回DatabaseError
// 
// 注意事项:
//   - 请求参数中的WorldviewId必须为正数
//   - 当应删除的记录不存在时，返回 NotFoundError 错误，而不是返回成功
//   - 删除前会验证世界观是否存在，以确保返回一致的错误信息
//   - 当前实现不检查该世界观下是否有关联的数据，删除可能会导致数据不一致
//   - 使用互斥锁保护删除操作，确保数据一致性
func (s *WorldviewService) DeleteWorldview(req *background.DeleteWorldviewRequest) error {
	if req == nil || req.WorldviewId <= 0 {
		err := errno.InvalidParameterError("请求不能为空或ID必须为正数")
		hlog.CtxErrorf(s.ctx, "删除世界观失败: %v", err)
		return err
	}

	// 加锁保护数据一致性
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证世界观是否存在
	_, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "删除世界观失败: 未找到ID为%d的世界观. 错误: %v", req.WorldviewId, err)
			return errno.NotFoundError("世界观")
		}
		hlog.CtxErrorf(s.ctx, "删除前检查世界观失败: %v", err)
		return errno.DatabaseError(fmt.Sprintf("验证世界观失败: %v", err))
	}

	// 执行删除操作
	err = db.DeleteWorldview(s.ctx, req.WorldviewId)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "在数据访问层删除ID为%d的世界观失败: %v", req.WorldviewId, err)
		if errors.Is(err, db.ErrWorldviewNotFound) {
			return errno.NotFoundError("世界观")
		}
		return errno.DatabaseError(fmt.Sprintf("删除世界观失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "ID为%d的世界观已成功删除", req.WorldviewId)
	return nil
}

// ListWorldviews 列出世界观
// 
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空
// 2. 处理分页参数，使用请求中的page和pageSize或默认值
// 3. 处理筛选参数，包括ParentID和TagFilter
// 4. 调用数据访问层的ListWorldviews方法获取列表和总数
// 5. 将数据库实体列表转换为服务层模型列表
// 6. 返回转换后的世界观列表和总数
// 
// 参数:
//   - req: 包含查询参数的请求对象，支持分页和多种筛选条件
// 
// 返回值:
//   - []*background.Worldview: 世界观列表，符合筛选条件和分页范围
//   - int64: 符合筛选条件的记录总数，不受分页限制
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空时返回InvalidParameterError
//     - 当数据库操作失败时返回相应错误
// 
// 注意事项:
//   - 请求中的Page和PageSize如果未设置或无效，会使用默认值
//   - ParentID默认为-1，表示不根据父世界观进行筛选
//   - TagFilter用于根据标签进行筛选
//   - 当没有符合条件的记录时，返回空列表而非nil，总数为0，错误为nil
//   - 该方法不需要加锁，因为它只执行只读操作
func (s *WorldviewService) ListWorldviews(req *background.ListWorldviewsRequest) ([]*background.Worldview, int64, error) {
	if req == nil {
		err := errno.InvalidParameterError("请求不能为空")
		hlog.CtxErrorf(s.ctx, "获取世界观失败: %v", err)
		return nil, 0, err
	}

	// 处理分页参数
	page := req.GetPage()
	if page <= 0 {
		page = 1 // 默认第一页
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10 // 默认每页 10 条
	}

	// 处理筛选参数
	// 默认不根据父ID进行筛选
	parentID := int64(-1)
	tagFilter := req.GetTagFilter()

	// 调用 DAL 层获取数据
	dbWorldviews, total, err := db.ListWorldviews(s.ctx, parentID, tagFilter, int(page), int(pageSize))
	if err != nil {
		hlog.CtxErrorf(s.ctx, "从数据访问层获取世界观列表失败: %v", err)
		return nil, 0, errno.DatabaseError(fmt.Sprintf("获取世界观列表失败: %v", err))
	}

	// 转换为 API 模型
	modelWorldviews := make([]*background.Worldview, 0, len(dbWorldviews))
	for _, dbWv := range dbWorldviews {
		modelWorldviews = append(modelWorldviews, convertDBWorldviewToModel(&dbWv))
	}

	hlog.CtxInfof(s.ctx, "成功获取世界观列表，总数: %d, 当前页: %d, 每页数量: %d", total, page, pageSize)
	return modelWorldviews, total, nil
}
