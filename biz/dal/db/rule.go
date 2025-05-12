package db

import (
	"context"
	"errors"
	"time"

	"novelai/pkg/constants"

	"gorm.io/gorm"
)

// Rule相关错误定义
var (
	ErrRuleNotFound      = errors.New("规则不存在")
	ErrCreateRuleFailed  = errors.New("创建规则失败")
	ErrUpdateRuleFailed  = errors.New("更新规则失败")
	ErrDeleteRuleFailed  = errors.New("删除规则失败")
	ErrListRulesFailed   = errors.New("列出规则失败")
)

// Rule 规则模型定义
// 对应 idl/background.proto 中的 Rule 消息
// 字段说明：
//   - ID: 规则ID
//   - WorldviewID: 所属世界观ID
//   - Name: 规则名称
//   - Description: 规则详细描述
//   - Tag: 标签，多个标签用英文逗号分隔
//   - ParentID: 父规则ID，0表示主规则 (顶级规则)
//   - CreatedAt: 创建时间（unix时间戳）
//   - UpdatedAt: 更新时间（unix时间戳）
type Rule struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	WorldviewID int64  `gorm:"index;not null" json:"worldview_id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Tag         string `gorm:"type:varchar(255)" json:"tag"`
	ParentID    int64  `gorm:"index" json:"parent_id"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 返回规则表名
func (Rule) TableName() string {
	return constants.TableNameRule
}

// CreateRule 创建新规则
// 参数:
//   - ctx: 上下文
//   - r: 规则信息结构体指针
// 返回:
//   - int64: 创建成功返回规则ID
//   - error: 操作错误信息
func CreateRule(ctx context.Context, r *Rule) (int64, error) {
	if r == nil {
		return 0, ErrCreateRuleFailed
	}
	result := DB.WithContext(ctx).Create(r)
	if result.Error != nil {
		return 0, errors.Join(ErrCreateRuleFailed, result.Error)
	}
	return r.ID, nil
}

// GetRuleByID 通过ID查询规则信息
// 参数:
//   - ctx: 上下文
//   - id: 规则ID
// 返回:
//   - *Rule: 规则信息
//   - error: 操作错误信息
func GetRuleByID(ctx context.Context, id int64) (*Rule, error) {
	var r Rule
	result := DB.WithContext(ctx).Where("id = ?", id).First(&r)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, result.Error
	}
	return &r, nil
}

// UpdateRule 更新规则信息
// 参数:
//   - ctx: 上下文
//   - id: 要更新的规则ID
//   - updates: 包含更新字段的map
// 返回:
//   - error: 操作错误信息
func UpdateRule(ctx context.Context, id int64, updates map[string]interface{}) error {
	if id == 0 {
		return ErrRuleNotFound
	}
	if updates == nil {
		return ErrUpdateRuleFailed
	}
	updates["updated_at"] = time.Now().Unix()

	result := DB.WithContext(ctx).Model(&Rule{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.Join(ErrUpdateRuleFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}

// DeleteRule 删除规则 (硬删除)
// 参数:
//   - ctx: 上下文
//   - id: 规则ID
// 返回:
//   - error: 操作错误信息
func DeleteRule(ctx context.Context, id int64) error {
	if id == 0 {
		return ErrRuleNotFound
	}
	result := DB.WithContext(ctx).Delete(&Rule{}, id)
	if result.Error != nil {
		return errors.Join(ErrDeleteRuleFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRuleNotFound
	}
	return nil
}

// ListRules 列出规则，支持分页和过滤
// 参数:
//   - ctx: 上下文
//   - worldviewIDFilter: 所属世界观ID筛选 (可选, 0或不传表示不筛选)
//   - parentIDFilter: 父规则ID筛选 (可选, 0表示顶级, -1或不传表示不筛选parent_id)
//   - tagFilter: 标签筛选
//   - page: 页码 (从1开始)
//   - pageSize: 每页数量
// 返回:
//   - []Rule: 规则列表
//   - int64: 总记录数
//   - error: 操作错误信息
func ListRules(ctx context.Context, worldviewIDFilter int64, parentIDFilter int64, tagFilter string, page, pageSize int) ([]Rule, int64, error) {
	var rules []Rule
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	dbQuery := DB.WithContext(ctx).Model(&Rule{})

	if worldviewIDFilter != 0 {
		dbQuery = dbQuery.Where("worldview_id = ?", worldviewIDFilter)
	}
	if parentIDFilter != -1 { // -1 表示不根据 parent_id 筛选
		dbQuery = dbQuery.Where("parent_id = ?", parentIDFilter)
	}
	if tagFilter != "" {
		dbQuery = dbQuery.Where("tag LIKE ?", "%"+tagFilter+"%")
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, errors.Join(ErrListRulesFailed, err)
	}

	offset := (page - 1) * pageSize
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error; err != nil {
		return nil, 0, errors.Join(ErrListRulesFailed, err)
	}

	return rules, total, nil
}
