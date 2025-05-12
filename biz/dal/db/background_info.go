package db

import (
	"context"
	"errors"
	"time"

	"novelai/pkg/constants"

	"gorm.io/gorm"
)

// BackgroundInfo相关错误定义
var (
	ErrBackgroundInfoNotFound     = errors.New("背景信息不存在")
	ErrCreateBackgroundInfoFailed = errors.New("创建背景信息失败")
	ErrUpdateBackgroundInfoFailed = errors.New("更新背景信息失败")
	ErrDeleteBackgroundInfoFailed = errors.New("删除背景信息失败")
	ErrListBackgroundInfosFailed  = errors.New("列出背景信息失败")
)

// BackgroundInfo 背景信息模型定义
// 对应 idl/background.proto 中的 BackgroundInfo 消息
// 字段说明：
//   - ID: 背景ID
//   - WorldviewID: 所属世界观ID
//   - Name: 背景名称
//   - Description: 背景详细描述
//   - Tag: 标签，多个标签用英文逗号分隔
//   - ParentID: 父背景ID，0表示主背景 (顶级背景)
//   - CreatedAt: 创建时间（unix时间戳）
//   - UpdatedAt: 更新时间（unix时间戳）
type BackgroundInfo struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	WorldviewID int64  `gorm:"index;not null" json:"worldview_id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Tag         string `gorm:"type:varchar(255)" json:"tag"`
	ParentID    int64  `gorm:"index" json:"parent_id"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 返回背景信息表名
func (BackgroundInfo) TableName() string {
	return constants.TableNameBackgroundInfo
}

// CreateBackgroundInfo 创建新背景信息
// 参数:
//   - ctx: 上下文
//   - bi: 背景信息结构体指针
// 返回:
//   - int64: 创建成功返回背景信息ID
//   - error: 操作错误信息
func CreateBackgroundInfo(ctx context.Context, bi *BackgroundInfo) (int64, error) {
	if bi == nil {
		return 0, ErrCreateBackgroundInfoFailed
	}
	result := DB.WithContext(ctx).Create(bi)
	if result.Error != nil {
		return 0, errors.Join(ErrCreateBackgroundInfoFailed, result.Error)
	}
	return bi.ID, nil
}

// GetBackgroundInfoByID 通过ID查询背景信息
// 参数:
//   - ctx: 上下文
//   - id: 背景信息ID
// 返回:
//   - *BackgroundInfo: 背景信息
//   - error: 操作错误信息
func GetBackgroundInfoByID(ctx context.Context, id int64) (*BackgroundInfo, error) {
	var bi BackgroundInfo
	result := DB.WithContext(ctx).Where("id = ?", id).First(&bi)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrBackgroundInfoNotFound
		}
		return nil, result.Error
	}
	return &bi, nil
}

// UpdateBackgroundInfo 更新背景信息
// 参数:
//   - ctx: 上下文
//   - id: 要更新的背景信息ID
//   - updates: 包含更新字段的map
// 返回:
//   - error: 操作错误信息
func UpdateBackgroundInfo(ctx context.Context, id int64, updates map[string]interface{}) error {
	if id == 0 {
		return ErrBackgroundInfoNotFound
	}
	if updates == nil {
		return ErrUpdateBackgroundInfoFailed
	}
	updates["updated_at"] = time.Now().Unix()

	result := DB.WithContext(ctx).Model(&BackgroundInfo{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.Join(ErrUpdateBackgroundInfoFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrBackgroundInfoNotFound
	}
	return nil
}

// DeleteBackgroundInfo 删除背景信息 (硬删除)
// 参数:
//   - ctx: 上下文
//   - id: 背景信息ID
// 返回:
//   - error: 操作错误信息
func DeleteBackgroundInfo(ctx context.Context, id int64) error {
	if id == 0 {
		return ErrBackgroundInfoNotFound
	}
	result := DB.WithContext(ctx).Delete(&BackgroundInfo{}, id)
	if result.Error != nil {
		return errors.Join(ErrDeleteBackgroundInfoFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrBackgroundInfoNotFound
	}
	return nil
}

// ListBackgroundInfos 列出背景信息，支持分页和过滤
// 参数:
//   - ctx: 上下文
//   - worldviewIDFilter: 所属世界观ID筛选 (可选, 0或不传表示不筛选)
//   - parentIDFilter: 父背景ID筛选 (可选, 0表示顶级, -1或不传表示不筛选parent_id)
//   - tagFilter: 标签筛选
//   - page: 页码 (从1开始)
//   - pageSize: 每页数量
// 返回:
//   - []BackgroundInfo: 背景信息列表
//   - int64: 总记录数
//   - error: 操作错误信息
func ListBackgroundInfos(ctx context.Context, worldviewIDFilter int64, parentIDFilter int64, tagFilter string, page, pageSize int) ([]BackgroundInfo, int64, error) {
	var bis []BackgroundInfo
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	dbQuery := DB.WithContext(ctx).Model(&BackgroundInfo{})

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
		return nil, 0, errors.Join(ErrListBackgroundInfosFailed, err)
	}

	offset := (page - 1) * pageSize
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&bis).Error; err != nil {
		return nil, 0, errors.Join(ErrListBackgroundInfosFailed, err)
	}

	return bis, total, nil
}
