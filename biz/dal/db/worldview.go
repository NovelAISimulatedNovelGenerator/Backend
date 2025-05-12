package db

import (
	"context"
	"errors"
	"time"

	"novelai/pkg/constants"

	"gorm.io/gorm"
)

// Worldview相关错误定义
var (
	ErrWorldviewNotFound      = errors.New("世界观不存在")
	ErrCreateWorldviewFailed  = errors.New("创建世界观失败")
	ErrUpdateWorldviewFailed  = errors.New("更新世界观失败")
	ErrDeleteWorldviewFailed  = errors.New("删除世界观失败")
	ErrListWorldviewsFailed = errors.New("列出世界观失败")
)

// Worldview 世界观模型定义
// 对应 idl/background.proto 中的 Worldview 消息
// 字段说明：
//   - ID: 世界观ID
//   - Name: 世界观名称
//   - Description: 世界观详细描述
//   - Tag: 标签，多个标签用英文逗号分隔
//   - ParentID: 父世界观ID，0表示主世界观 (顶级世界观)
//   - CreatedAt: 创建时间（unix时间戳）
//   - UpdatedAt: 更新时间（unix时间戳）
type Worldview struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Tag         string `gorm:"type:varchar(255)" json:"tag"`
	ParentID    int64  `gorm:"index" json:"parent_id"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 返回世界观表名
func (Worldview) TableName() string {
	return constants.TableNameWorldview
}

// CreateWorldview 创建新世界观
// 参数:
//   - ctx: 上下文
//   - wv: 世界观信息结构体指针
// 返回:
//   - int64: 创建成功返回世界观ID
//   - error: 操作错误信息
func CreateWorldview(ctx context.Context, wv *Worldview) (int64, error) {
	if wv == nil {
		return 0, ErrCreateWorldviewFailed
	}
	result := DB.WithContext(ctx).Create(wv)
	if result.Error != nil {
		return 0, errors.Join(ErrCreateWorldviewFailed, result.Error)
	}
	return wv.ID, nil
}

// GetWorldviewByID 通过ID查询世界观信息
// 参数:
//   - ctx: 上下文
//   - id: 世界观ID
// 返回:
//   - *Worldview: 世界观信息
//   - error: 操作错误信息
func GetWorldviewByID(ctx context.Context, id int64) (*Worldview, error) {
	var wv Worldview
	result := DB.WithContext(ctx).Where("id = ?", id).First(&wv)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrWorldviewNotFound
		}
		return nil, result.Error
	}
	return &wv, nil
}

// UpdateWorldview 更新世界观信息
// 参数:
//   - ctx: 上下文
//   - id: 要更新的世界观ID
//   - updates: 包含更新字段的map
// 返回:
//   - error: 操作错误信息
func UpdateWorldview(ctx context.Context, id int64, updates map[string]interface{}) error {
	if id == 0 {
		return ErrWorldviewNotFound // 或者 ErrUpdateWorldviewFailed 加上具体原因
	}
	if updates == nil {
		return ErrUpdateWorldviewFailed
	}
	// 确保更新时间被设置
	updates["updated_at"] = time.Now().Unix()

	result := DB.WithContext(ctx).Model(&Worldview{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.Join(ErrUpdateWorldviewFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrWorldviewNotFound // 可能记录不存在
	}
	return nil
}

// DeleteWorldview 删除世界观 (硬删除)
// 参数:
//   - ctx: 上下文
//   - id: 世界观ID
// 返回:
//   - error: 操作错误信息
func DeleteWorldview(ctx context.Context, id int64) error {
	if id == 0 {
		return ErrWorldviewNotFound
	}
	result := DB.WithContext(ctx).Delete(&Worldview{}, id)
	if result.Error != nil {
		return errors.Join(ErrDeleteWorldviewFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrWorldviewNotFound
	}
	return nil
}

// ListWorldviews 列出世界观，支持分页和过滤
// 参数:
//   - ctx: 上下文
//   - parentIDFilter: 父世界观ID筛选 (0表示顶级, -1或不传表示不筛选parent_id)
//   - tagFilter: 标签筛选 (部分匹配)
//   - page: 页码 (从1开始)
//   - pageSize: 每页数量
// 返回:
//   - []Worldview: 世界观列表
//   - int64: 总记录数
//   - error: 操作错误信息
func ListWorldviews(ctx context.Context, parentIDFilter int64, tagFilter string, page, pageSize int) ([]Worldview, int64, error) {
	var worldviews []Worldview
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // 默认每页10条
	}

	dbQuery := DB.WithContext(ctx).Model(&Worldview{})

	if parentIDFilter != -1 { // -1 表示不根据 parent_id 筛选
		dbQuery = dbQuery.Where("parent_id = ?", parentIDFilter)
	}

	if tagFilter != "" {
		dbQuery = dbQuery.Where("tag LIKE ?", "%"+tagFilter+"%")
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, errors.Join(ErrListWorldviewsFailed, err)
	}

	offset := (page - 1) * pageSize
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&worldviews).Error; err != nil {
		return nil, 0, errors.Join(ErrListWorldviewsFailed, err)
	}

	return worldviews, total, nil
}
