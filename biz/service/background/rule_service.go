package background

import (
	"context"
	"errors"

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
		hlog.CtxWarnf(s.ctx, "CreateRule: request is nil")
		return nil, errno.InvalidParameterError("Request is nil")
	}

	// 验证请求参数，例如 WorldviewID 是否存在等（根据实际业务需求添加）
	if req.WorldviewId <= 0 {
		hlog.CtxWarnf(s.ctx, "CreateRule: invalid WorldviewId: %d", req.WorldviewId)
		return nil, errno.InvalidParameterError("Invalid WorldviewId")
	}

	if req.Name == "" {
		hlog.CtxWarnf(s.ctx, "CreateRule: name is required")
		return nil, errno.InvalidParameterError("Rule name is required")
	}

	dbRule := &db.Rule{
		WorldviewID: req.WorldviewId,
		Name:        req.Name,
		Description: req.Description,
		Tag:         req.Tag,
		ParentID:    req.ParentId,
	}

	ruleID, err := db.CreateRule(s.ctx, dbRule)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "CreateRule: failed to create rule in DB: %v", err)
		return nil, errno.DatabaseError("Failed to create rule")
	}
	
	// 确保 ID 已正确设置
	if ruleID != dbRule.ID {
		dbRule.ID = ruleID
	}

	hlog.CtxInfof(s.ctx, "CreateRule: successfully created rule with ID: %d", dbRule.ID)
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
		hlog.CtxWarnf(s.ctx, "GetRuleByID: invalid request or RuleId: %v", req)
		return nil, errno.InvalidParameterError("Invalid rule ID")
	}

	dbRule, err := db.GetRuleByID(s.ctx, req.RuleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "GetRuleByID: rule with ID %d not found: %v", req.RuleId, err)
			return nil, errno.NotFoundError("Rule")
		}
		hlog.CtxErrorf(s.ctx, "GetRuleByID: failed to get rule from DB: %v", err)
		return nil, errno.DatabaseError("Failed to retrieve rule")
	}

	hlog.CtxInfof(s.ctx, "GetRuleByID: successfully retrieved rule with ID: %d", dbRule.ID)
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
		hlog.CtxWarnf(s.ctx, "UpdateRule: invalid request or RuleId: %v", req)
		return nil, errno.InvalidParameterError("Invalid rule ID")
	}

	// 检查规则是否存在
	_, err := db.GetRuleByID(s.ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "UpdateRule: rule with ID %d not found for update: %v", req.Id, err)
			return nil, errno.NotFoundError("Rule")
		}
		hlog.CtxErrorf(s.ctx, "UpdateRule: failed to get rule before update: %v", err)
		return nil, errno.DatabaseError("Failed to verify rule before update")
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
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}
	// 注意：ParentId为0可能是有效值（表示顶级规则），所以可能需要特别处理
	// 这里我们总是更新 parent_id
	updates["parent_id"] = req.ParentId

	if len(updates) == 0 {
		hlog.CtxInfof(s.ctx, "UpdateRule: no fields to update for rule ID %d", req.Id)
		// 根据业务需求，可以选择返回错误或直接返回原对象
		dbRule, _ := db.GetRuleByID(s.ctx, req.Id) // 重新获取以确保数据最新
		return convertDBRuleToModel(dbRule), nil
	}

	if err := db.UpdateRule(s.ctx, req.Id, updates); err != nil {
		hlog.CtxErrorf(s.ctx, "UpdateRule: failed to update rule in DB for ID %d: %v", req.Id, err)
		return nil, errno.DatabaseError("Failed to update rule")
	}

	dbRuleUpdated, err := db.GetRuleByID(s.ctx, req.Id)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "UpdateRule: failed to retrieve updated rule with ID %d: %v", req.Id, err)
		return nil, errno.DatabaseError("Failed to retrieve rule after update")
	}

	hlog.CtxInfof(s.ctx, "UpdateRule: successfully updated rule with ID: %d", req.Id)
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
		hlog.CtxWarnf(s.ctx, "DeleteRule: invalid request or RuleId: %v", req)
		return errno.InvalidParameterError("Invalid rule ID")
	}

	// 检查规则是否存在
	_, err := db.GetRuleByID(s.ctx, req.RuleId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, db.ErrRuleNotFound) {
			hlog.CtxWarnf(s.ctx, "DeleteRule: rule with ID %d not found for deletion: %v", req.RuleId, err)
			return errno.NotFoundError("Rule")
		}
		hlog.CtxErrorf(s.ctx, "DeleteRule: failed to get rule before deletion: %v", err)
		return errno.DatabaseError("Failed to verify rule before deletion")
	}

	if err := db.DeleteRule(s.ctx, req.RuleId); err != nil {
		hlog.CtxErrorf(s.ctx, "DeleteRule: failed to delete rule in DB for ID %d: %v", req.RuleId, err)
		return errno.DatabaseError("Failed to delete rule")
	}

	hlog.CtxInfof(s.ctx, "DeleteRule: successfully deleted rule with ID: %d", req.RuleId)
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
		err := errno.InvalidParameterError("ListRules request cannot be nil")
		hlog.CtxErrorf(s.ctx, "ListRules failed: %v", err)
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
		err := errno.DatabaseError("Failed to list rules")
		hlog.CtxErrorf(s.ctx, "ListRules: failed to list rules from DB: %v, request: %+v", err, req)
		return nil, 0, err
	}

	modelRules := make([]*background.Rule, 0, len(dbRules))
	for i := range dbRules {
		modelRules = append(modelRules, convertDBRuleToModel(&dbRules[i]))
	}

	hlog.CtxInfof(s.ctx, "ListRules: successfully retrieved %d rules, total: %d", len(modelRules), total)
	return modelRules, total, nil
}
