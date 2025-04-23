/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"errors"
	"time"

	"novelai/pkg/constants"

	"gorm.io/gorm"
)

// 存档相关错误定义
var (
	ErrSaveNotFound     = errors.New("存档不存在")
	ErrCreateSaveFailed = errors.New("创建存档失败")
	ErrUpdateSaveFailed = errors.New("更新存档失败")
)

// Save 存档模型定义
// 表示用户的保存项，包含保存内容、类型、状态等信息
// 字段说明：
//   - ID: 保存项ID
//   - UserID: 用户ID
//   - SaveID: 保存项唯一标识符
//   - SaveName: 保存项名称
//   - SaveDescription: 保存项描述
//   - SaveData: 保存的具体内容（如JSON字符串）
//   - SaveType: 保存类型（如草稿、配置等）
//   - SaveStatus: 保存状态（如active、deleted等）
//   - CreatedAt: 创建时间（unix时间戳）
//   - UpdatedAt: 更新时间（unix时间戳）
type Save struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`                      // 保存项ID
	UserID          int64          `gorm:"index;not null" json:"user_id"`                          // 用户ID
	SaveID          string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"save_id"`    // 保存项唯一标识符
	SaveName        string         `gorm:"type:varchar(128);not null" json:"save_name"`             // 保存项名称
	SaveDescription string         `gorm:"type:varchar(512)" json:"save_description"`               // 保存项描述
	SaveData        string         `gorm:"type:text;not null" json:"save_data"`                     // 保存的具体内容
	SaveType        string         `gorm:"type:varchar(32);not null" json:"save_type"`              // 保存类型
	SaveStatus      string         `gorm:"type:varchar(16);not null" json:"save_status"`            // 保存状态
	CreatedAt       int64          `gorm:"autoCreateTime" json:"created_at"`                        // 创建时间(unix时间戳)
	UpdatedAt       int64          `gorm:"autoUpdateTime" json:"updated_at"`                        // 更新时间(unix时间戳)
}

// TableName 返回存档表名
func (Save) TableName() string {
	return constants.TableNameSave
}

// CreateSave 创建新存档
// 参数:
//   - save: 存档信息结构体指针
//
// 返回:
//   - int64: 创建成功返回存档ID
//   - error: 操作错误信息
func CreateSave(save *Save) (int64, error) {
	if save == nil {
		return 0, ErrCreateSaveFailed
	}
	if err := DB.Create(save).Error; err != nil {
		return 0, ErrCreateSaveFailed
	}
	return save.ID, nil
}

// QuerySaveByID 通过存档ID查询存档信息
// 参数:
//   - saveID: 存档ID
//
// 返回:
//   - *Save: 存档信息
//   - error: 操作错误信息
func QuerySaveByID(saveID int64) (*Save, error) {
	var save Save
	if err := DB.Where("id = ?", saveID).First(&save).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSaveNotFound
		}
		return nil, err
	}
	return &save, nil
}

// QuerySavesByUser 根据用户ID获取该用户所有存档，支持分页
// 参数:
//   - userID: 用户ID
//   - page: 页码（从1开始）
//   - pageSize: 每页记录数
//
// 返回:
//   - []Save: 存档列表
//   - int64: 总记录数
//   - error: 操作错误信息
func QuerySavesByUser(userID int64, page, pageSize int) ([]Save, int64, error) {
	var saves []Save
	var total int64
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	db := DB.Model(&Save{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&saves).Error; err != nil {
		return nil, 0, err
	}
	return saves, total, nil
}

// UpdateSave 更新存档内容
// 参数:
//   - save: 包含更新内容的存档结构体，必须有ID
//
// 返回:
//   - error: 操作错误信息
func UpdateSave(save *Save) error {
	if save == nil || save.ID == 0 {
		return ErrUpdateSaveFailed
	}
	m := map[string]interface{}{
		"save_name":        save.SaveName,
		"save_description": save.SaveDescription,
		"save_data":        save.SaveData,
		"save_type":        save.SaveType,
		"save_status":      save.SaveStatus,
		"updated_at":       time.Now().Unix(),
	}
	if err := DB.Model(&Save{}).Where("id = ?", save.ID).Updates(m).Error; err != nil {
		return ErrUpdateSaveFailed
	}
	return nil
}

// DeleteSave 删除存档（软删除）
// 参数:
//   - saveID: 存档ID
//
// 返回:
//   - error: 操作错误信息
func DeleteSave(saveID int64) error {
	if saveID == 0 {
		return ErrSaveNotFound
	}
	if err := DB.Where("id = ?", saveID).Delete(&Save{}).Error; err != nil {
		return err
	}
	return nil
}

// ListSaves 获取所有存档（支持分页）
// 参数:
//   - page: 页码
//   - pageSize: 每页记录数
//
// 返回:
//   - []Save: 存档列表
//   - int64: 总记录数
//   - error: 操作错误信息
func ListSaves(page, pageSize int) ([]Save, int64, error) {
	var saves []Save
	var total int64
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	db := DB.Model(&Save{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&saves).Error; err != nil {
		return nil, 0, err
	}
	return saves, total, nil
}

// CheckSaveExists 检查存档是否存在
// 参数:
//   - saveID: 存档ID
//
// 返回:
//   - bool: 是否存在
//   - error: 操作错误信息
func CheckSaveExists(saveID int64) (bool, error) {
	var count int64
	if err := DB.Model(&Save{}).Where("id = ?", saveID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
