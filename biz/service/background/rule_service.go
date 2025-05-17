package background

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"

	"novelai/biz/dal/db"
	"novelai/biz/model/background"
	"novelai/pkg/errno"
)

// RuleService 负责处理规则相关的业务逻辑
type RuleService struct {
	ctx context.Context     // 当前上下文
	app *app.RequestContext // Hertz 的请求上下文
	mu  sync.Mutex          // 互斥锁，用于保护并发操作
}

// NewRuleService 创建一个新的 RuleService 实例
//
// 工作流程:
// 1. 接收上下文和请求上下文参数
// 2. 创建并返回 RuleService 实例，包含所需的上下文信息
//
// 参数:
//   - ctx: 业务上下文，用于日志记录和数据库操作的上下文传递
//   - appCtx: Hertz 框架的请求上下文，包含当前 HTTP 请求的相关信息
//
// 返回值:
//   - *RuleService: 新创建的 RuleService 实例，可用于处理规则相关的业务逻辑
//
// 注意事项:
//   - 该函数是工厂方法，每个请求都应创建新的 RuleService 实例
//   - 返回的服务实例包含互斥锁，可以安全地在并发环境中使用
func NewRuleService(ctx context.Context, appCtx *app.RequestContext) *RuleService {
	return &RuleService{
		ctx: ctx,
		app: appCtx,
	}
}

// convertDBRuleToModel 将数据库模型转换为服务层模型
//
// 工作流程:
// 1. 检查输入的数据库模型指针是否为 nil
// 2. 将数据库模型的字段映射到服务层模型的对应字段
//
// 参数:
//   - dbRule: 数据库层的规则实体指针，包含从数据库获取的完整规则信息
//
// 返回值:
//   - *background.Rule: 转换后的服务层规则实体，适用于 API 响应
//   - 当输入为 nil 时返回 nil
//
// 注意事项:
//   - 该函数是纯函数，不依赖服务实例的状态
//   - 实现了 DAO 层和服务层之间的数据模型转换，解耦数据访问和业务逻辑
func convertDBRuleToModel(dbRule *db.Rule) *background.Rule {
	if dbRule == nil {
		return nil
	}
	return &background.Rule{
		Id:          dbRule.ID,
		WorldviewId: dbRule.WorldviewID,
		Name:        dbRule.Name,
		Description: dbRule.Description,
		Tag:         dbRule.Tag,
		ParentId:    dbRule.ParentID,
		CreatedAt:   dbRule.CreatedAt,
		UpdatedAt:   dbRule.UpdatedAt,
	}
}

// CreateRule 创建新的规则
//
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空且名称必填
// 2. 验证所提供的世界观ID是否存在
// 3. 如果提供了ParentId，验证父规则是否存在
// 4. 验证父规则和当前规则是否属于同一世界观
// 5. 创建数据库实体并调用数据访问层执行创建操作
// 6. 返回创建成功的规则实体
//
// 参数:
//   - req: 包含要创建的规则详细内容的请求对象
//
// 返回值:
//   - *background.Rule: 新创建的规则对象，包含数据库分配的ID和时间戳
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空或缺失必要字段时返回InvalidParameterError
//     - 当引用的世界观或父规则不存在时返回InvalidParameterError
//     - 当父规则与当前规则不属于同一世界观时返回InvalidParameterError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - WorldviewId 字段是必填项，必须为正数
//   - Name 字段是必填项，不能为空
//   - ParentId 可以为0，表示没有父规则
//   - 如果提供了ParentId，则ParentId对应的规则与新创建的规则必须属于同一世界观下
//   - 创建成功后会返回完整的规则实体，包含数据库生成的ID和时间戳
func (s *RuleService) CreateRule(req *background.CreateRuleRequest) (*background.Rule, error) {
	if req == nil {
		hlog.CtxWarnf(s.ctx, "CreateRule: 请求为空")
		return nil, errno.InvalidParameterError("请求不能为空")
	}

	// 验证请求参数
	if req.WorldviewId <= 0 {
		hlog.CtxWarnf(s.ctx, "CreateRule: 无效的世界观ID: %d", req.WorldviewId)
		return nil, errno.InvalidParameterError("世界观ID必须为正数")
	}

	if req.Name == "" {
		hlog.CtxWarnf(s.ctx, "CreateRule: 名称是必填项")
		return nil, errno.InvalidParameterError("规则名称不能为空")
	}
	
	// 验证世界观是否存在
	_, err := db.GetWorldviewByID(s.ctx, req.WorldviewId)
	if err != nil {
		if errors.Is(err, db.ErrWorldviewNotFound) {
			hlog.CtxWarnf(s.ctx, "CreateRule: 指定的世界观ID %d 不存在", req.WorldviewId)
			return nil, errno.InvalidParameterError("指定的世界观不存在")
		}
		hlog.CtxErrorf(s.ctx, "CreateRule: 验证世界观ID %d 时发生错误: %v", req.WorldviewId, err)
		return nil, errno.DatabaseError(fmt.Sprintf("验证世界观失败: %v", err))
	}
	
	// 验证父规则ID是否存在(如果有)
	if req.ParentId > 0 {
		parent, err := db.GetRuleByID(s.ctx, req.ParentId)
		if err != nil {
			if errors.Is(err, db.ErrRuleNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
				hlog.CtxWarnf(s.ctx, "CreateRule: 指定的父规则ID %d 不存在", req.ParentId)
				return nil, errno.InvalidParameterError("指定的父规则不存在")
			}
			hlog.CtxErrorf(s.ctx, "CreateRule: 验证父规则ID %d 时发生错误: %v", req.ParentId, err)
			return nil, errno.DatabaseError(fmt.Sprintf("验证父规则失败: %v", err))
		}
		
		// 验证父规则是否属于同一世界观
		if parent.WorldviewID != req.WorldviewId {
			hlog.CtxWarnf(s.ctx, "CreateRule: 父规则(世界观ID=%d)与当前规则(世界观ID=%d)不属于同一世界观", parent.WorldviewID, req.WorldviewId)
			return nil, errno.InvalidParameterError("父规则必须属于同一世界观")
		}
	}

	dbRule := &db.Rule{
		WorldviewID: req.WorldviewId,
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ruleID, err := db.CreateRule(s.ctx, dbRule)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "CreateRule: 在数据库中创建规则失败: %v", err)
		return nil, errno.DatabaseError(fmt.Sprintf("创建规则失败: %v", err))
	}

	// 确保 ID 已正确设置
	if ruleID != dbRule.ID {
		dbRule.ID = ruleID
	}

	hlog.CtxInfof(s.ctx, "CreateRule: 成功创建规则，ID为: %d", dbRule.ID)
	return convertDBRuleToModel(dbRule), nil
}

// GetRuleByID 通过ID获取单个规则
//
// 工作流程:
// 1. 验证规则ID参数的有效性，确保是正数
// 2. 调用数据访问层的 GetRuleByID 方法获取规则
// 3. 处理可能出现的错误情况，包括记录不存在和数据库错误
// 4. 将数据库实体转换为服务层模型并返回
//
// 参数:
//   - req: 包含要查询的规则ID的请求对象
//
// 返回值:
//   - *background.Rule: 返回找到的规则对象，如果没有找到则为 nil
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回NotFoundError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 请求参数中的RuleId必须为正数
//   - 当记录不存在时，返回的错误使用项目定义的errno.NotFoundError而非nil
func (s *RuleService) GetRuleByID(req *background.GetRuleRequest) (*background.Rule, error) {
	if req == nil || req.RuleId <= 0 {
		hlog.CtxWarnf(s.ctx, "GetRuleByID: 无效的请求或规则ID: %v", req)
		return nil, errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}

	dbRule, err := db.GetRuleByID(s.ctx, req.RuleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "GetRuleByID: ID为%d的规则未找到: %v", req.RuleId, err)
			return nil, errno.NotFoundError("规则")
		}
		hlog.CtxErrorf(s.ctx, "GetRuleByID: 从数据库获取规则失败: %v", err)
		return nil, errno.DatabaseError(fmt.Sprintf("获取规则失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "GetRuleByID: 成功获取ID为%d的规则", dbRule.ID)
	return convertDBRuleToModel(dbRule), nil
}

// UpdateRule 更新现有的规则
//
// 工作流程:
// 1. 验证请求参数的有效性，确保ID为正数
// 2. 验证要更新的规则是否存在
// 3. 如果更新世界观ID，验证新世界观是否存在
// 4. 验证新父规则ID(如果有)是否存在，并属于同一世界观
// 5. 构建需要更新的字段映射，仅包含请求中提供的字段
// 6. 调用数据访问层更新规则信息
// 7. 获取更新后的实体并返回
//
// 参数:
//   - req: 包含规则ID和需要更新的字段的请求对象
//
// 返回值:
//   - *background.Rule: 更新后的规则对象，包含最新的信息
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空或ID无效时返回InvalidParameterError
//     - 当规则不存在时返回NotFoundError
//     - 当引用的世界观或父规则不存在时返回InvalidParameterError
//     - 当父规则与当前规则不属于同一世界观时返回InvalidParameterError
//     - 当数据库操作失败时返回DatabaseError
//
// 注意事项:
//   - 当没有任何字段需要更新时，会直接返回当前规则对象
//   - Name, Description 仅在非空时才会更新
//   - Tag 始终会被更新，包括更新为空字符串
//   - ParentId 不等于-1时才会更新，允许更新为0(表示没有父规则)
//   - 更新后会重新查询数据获取完整的更新后信息
func (s *RuleService) UpdateRule(req *background.UpdateRuleRequest) (*background.Rule, error) {
	if req == nil || req.Id <= 0 {
		hlog.CtxWarnf(s.ctx, "UpdateRule: 无效的请求或规则ID: %v", req)
		return nil, errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}

	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 检查规则是否存在
	origRule, err := db.GetRuleByID(s.ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "UpdateRule: 未找到ID为%d的规则进行更新: %v", req.Id, err)
			return nil, errno.NotFoundError("规则")
		}
		hlog.CtxErrorf(s.ctx, "UpdateRule: 更新前获取规则失败: %v", err)
		return nil, errno.DatabaseError(fmt.Sprintf("验证规则失败: %v", err))
	}
	
	// 如果更新世界观ID，验证新世界观是否存在
	if req.WorldviewId > 0 && req.WorldviewId != origRule.WorldviewID {
		_, err = db.GetWorldviewByID(s.ctx, req.WorldviewId)
		if err != nil {
			if errors.Is(err, db.ErrWorldviewNotFound) {
				hlog.CtxWarnf(s.ctx, "UpdateRule: 指定的新世界观ID %d 不存在", req.WorldviewId)
				return nil, errno.InvalidParameterError("指定的新世界观不存在")
			}
			hlog.CtxErrorf(s.ctx, "UpdateRule: 验证新世界观ID %d 时发生错误: %v", req.WorldviewId, err)
			return nil, errno.DatabaseError(fmt.Sprintf("验证新世界观失败: %v", err))
		}
	}
	
	// 如果更新父规则ID，验证父规则是否存在且属于同一世界观
	if req.ParentId > 0 && req.ParentId != origRule.ParentID {
		parent, err := db.GetRuleByID(s.ctx, req.ParentId)
		if err != nil {
			if errors.Is(err, db.ErrRuleNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
				hlog.CtxWarnf(s.ctx, "UpdateRule: 指定的父规则ID %d 不存在", req.ParentId)
				return nil, errno.InvalidParameterError("指定的父规则不存在")
			}
			hlog.CtxErrorf(s.ctx, "UpdateRule: 验证父规则ID %d 时发生错误: %v", req.ParentId, err)
			return nil, errno.DatabaseError(fmt.Sprintf("验证父规则失败: %v", err))
		}
		
		// 父规则世界观ID与当前规则世界观ID必须相同
		// 值得注意的是，如果同时更新世界观ID和父规则ID，我们应该使用新的世界观ID进行比较
		effectiveWorldviewID := origRule.WorldviewID
		if req.WorldviewId > 0 {
			effectiveWorldviewID = req.WorldviewId
		}
		if parent.WorldviewID != effectiveWorldviewID {
			hlog.CtxWarnf(s.ctx, "UpdateRule: 父规则(世界观ID=%d)与当前规则(世界观ID=%d)不属于同一世界观", parent.WorldviewID, effectiveWorldviewID)
			return nil, errno.InvalidParameterError("父规则必须属于同一世界观")
		}
	}

	updates := make(map[string]interface{})
	if req.WorldviewId > 0 {
		updates["worldview_id"] = req.WorldviewId
	}
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
		hlog.CtxInfof(s.ctx, "UpdateRule: ID为%d的规则没有字段需要更新", req.Id)
		// 返回当前对象
		return convertDBRuleToModel(origRule), nil
	}

	if err := db.UpdateRule(s.ctx, req.Id, updates); err != nil {
		hlog.CtxErrorf(s.ctx, "UpdateRule: 在数据库中更新ID为%d的规则失败: %v", req.Id, err)
		if errors.Is(err, db.ErrRuleNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.NotFoundError("规则")
		}
		return nil, errno.DatabaseError(fmt.Sprintf("更新规则失败: %v", err))
	}

	dbRuleUpdated, err := db.GetRuleByID(s.ctx, req.Id)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "UpdateRule: 获取更新后ID为%d的规则失败: %v", req.Id, err)
		return nil, errno.DatabaseError(fmt.Sprintf("获取更新后的规则失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "UpdateRule: 成功更新ID为%d的规则", req.Id)
	return convertDBRuleToModel(dbRuleUpdated), nil
}

// DeleteRule 通过ID删除规则
// 
// 工作流程:
// 1. 验证规则ID参数的有效性，确保是正数
// 2. 调用数据访问层的 DeleteRule 方法删除规则
// 3. 处理可能出现的错误情况，包括记录不存在和数据库错误
// 
// 参数:
//   - req: 包含要删除的规则ID的请求对象
// 
// 返回值:
//   - error: 操作过程中的错误，成功时返回nil
//     - 当ID无效时返回InvalidParameterError
//     - 当记录不存在时返回NotFoundError
//     - 当数据库操作失败时返回DatabaseError
// 
// 注意事项:
//   - 请求参数中的RuleId必须为正数
//   - 当应删除的记录不存在时，返回 NotFoundError 错误，而不是返回成功
//   - 当前实现不检查该规则是否有子规则，删除可能会导致父子关系数据不一致
func (s *RuleService) DeleteRule(req *background.DeleteRuleRequest) error {
	if req == nil || req.RuleId <= 0 {
		hlog.CtxWarnf(s.ctx, "DeleteRule: 无效的请求或规则ID: %v", req)
		return errno.InvalidParameterError("请求不能为空或ID必须为正数")
	}

	// 加锁保护并发安全
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 检查规则是否存在
	_, err := db.GetRuleByID(s.ctx, req.RuleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "DeleteRule: 未找到ID为%d的规则进行删除: %v", req.RuleId, err)
			return errno.NotFoundError("规则")
		}
		hlog.CtxErrorf(s.ctx, "DeleteRule: 删除前获取规则失败: %v", err)
		return errno.DatabaseError(fmt.Sprintf("验证规则失败: %v", err))
	}
	
	// 根据需要，检查是否有子规则依赖于该规则
	// 可以添加代码检查更复杂的依赖关系
	// 例如：
	// childRules, _, err := db.ListRules(s.ctx, rule.WorldviewID, rule.ID, "", 1, 1)
	// if err == nil && len(childRules) > 0 {
	//     return errno.InvalidParameterError("无法删除存在子规则的规则")
	// }

	if err := db.DeleteRule(s.ctx, req.RuleId); err != nil {
		hlog.CtxErrorf(s.ctx, "DeleteRule: 在数据库中删除ID为%d的规则失败: %v", req.RuleId, err)
		return errno.DatabaseError(fmt.Sprintf("删除规则失败: %v", err))
	}

	hlog.CtxInfof(s.ctx, "DeleteRule: 成功删除ID为%d的规则", req.RuleId)
	return nil
}

// ListRules 列出规则，支持分页和过滤
// 
// 工作流程:
// 1. 验证请求参数的有效性，确保请求不为空
// 2. 处理分页参数，使用请求中的page和pageSize或默认值
// 3. 处理筛选参数，包括WorldviewId、IsEnabled和Name
// 4. 调用数据访问层的ListRules方法获取列表和总数
// 5. 将数据库实体列表转换为服务层模型列表
// 6. 返回转换后的规则列表和总数
// 
// 参数:
//   - req: 包含查询参数的请求对象，支持分页和多种筛选条件
// 
// 返回值:
//   - []*background.Rule: 规则列表，符合筛选条件和分页范围
//   - int64: 符合筛选条件的记录总数，不受分页限制
//   - error: 操作过程中的错误，成功时返回nil
//     - 当请求为空时返回InvalidParameterError
//     - 当数据库操作失败时返回相应错误
// 
// 注意事项:
//   - 请求中的page和pageSize如果未设置或无效，会使用默认值
//   - 当没有符合条件的记录时，返回空列表而非nil，总数为0，错误为nil
//   - 世界观ID、是否启用和名称筛选可以组合使用，形成AND逻辑
func (s *RuleService) ListRules(req *background.ListRulesRequest) ([]*background.Rule, int64, error) {
	if req == nil {
		err := errno.InvalidParameterError("请求不能为空")
		hlog.CtxErrorf(s.ctx, "ListRules失败: %v", err)
		return nil, 0, err
	}

	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1 // Default page is 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Default page size is 10
	}

	var worldviewIDFilter int64 = 0 // 默认不筛选
	if req.WorldviewIdFilter > 0 {
		worldviewIDFilter = req.WorldviewIdFilter
	}

	var parentIDFilter int64 = -1 // 默认不筛选
	if req.ParentIdFilter >= 0 {
		parentIDFilter = req.ParentIdFilter
	}

	tagFilter := req.TagFilter

	dbRules, total, err := db.ListRules(s.ctx, worldviewIDFilter, parentIDFilter, tagFilter, int(page), int(pageSize))
	if err != nil {
		hlog.CtxErrorf(s.ctx, "ListRules: 从数据库列出规则失败: %v, 请求: %+v", err, req)
		return nil, 0, errno.DatabaseError(fmt.Sprintf("获取规则列表失败: %v", err))
	}

	modelRules := make([]*background.Rule, 0, len(dbRules))
	for i := range dbRules {
		modelRules = append(modelRules, convertDBRuleToModel(&dbRules[i]))
	}

	hlog.CtxInfof(s.ctx, "ListRules: 成功获取%d条规则，总计: %d", len(modelRules), total)
	return modelRules, total, nil
}
