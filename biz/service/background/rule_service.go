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

// RuleService 用于管理规则相关的业务逻辑
type RuleService struct {
	ctx context.Context     // 当前上下文
	app *app.RequestContext // Hertz 的请求上下文
	mu  sync.Mutex          // 互斥锁，用于保护并发操作
}

// NewRuleService 创建 RuleService 实例
// 参数:
//   - ctx: 上下文
//   - appCtx: Hertz 请求上下文
//
// 返回:
//   - *RuleService: RuleService 实例
func NewRuleService(ctx context.Context, appCtx *app.RequestContext) *RuleService {
	return &RuleService{
		ctx: ctx,
		app: appCtx,
	}
}

// convertDBRuleToModel 将 DAL 层 Rule 结构转换为 API 模型结构
// 参数:
//   - dbRule: 数据库规则结构指针
//
// 返回:
//   - *background.Rule: 服务层规则结构
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
// 参数:
//   - req: 创建规则的请求参数，包含名称、描述、标签、父ID等
//
// 返回:
//   - *background.Rule: 创建成功后的规则信息
//   - error: 操作错误信息
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

// GetRuleByID 根据 ID 获取规则信息
// 参数:
//   - req: 获取规则的请求参数，包含规则 ID
//
// 返回:
//   - *background.Rule: 规则信息
//   - error: 操作错误信息
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

// UpdateRule 更新规则信息
// 参数:
//   - req: 更新规则的请求参数，包含规则 ID 及需要更新的字段
//
// 返回:
//   - *background.Rule: 更新后的规则信息
//   - error: 操作错误信息
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

// DeleteRule 删除规则
// 参数:
//   - req: 删除规则的请求参数，包含规则 ID
//
// 返回:
//   - error: 操作错误信息
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
// 参数:
//   - req: 列出规则的请求参数，包含过滤条件和分页参数
//
// 返回:
//   - []*background.Rule: 规则列表
//   - int64: 总记录数
//   - error: 操作错误信息
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
