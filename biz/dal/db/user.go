/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// 用户相关错误定义
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户名已存在")
	ErrInvalidPassword   = errors.New("密码验证失败")
	ErrCreateUserFailed  = errors.New("创建用户失败")
	ErrUpdateUserFailed  = errors.New("更新用户信息失败")
)

// TableName 用户表名常量
const TableNameUser = "users"

// User 用户模型定义
// 包含用户基本信息及偏好设置
type User struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`                    // 用户唯一标识
	Username        string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"` // 用户名，唯一
	Password        string         `gorm:"type:varchar(256);not null" json:"-"`                   // 密码，安全起见不返回给客户端
	Nickname        string         `gorm:"type:varchar(64)" json:"nickname"`                      // 昵称
	Email           *string        `gorm:"type:varchar(128);uniqueIndex" json:"email"`            // 邮箱
	Avatar          string         `gorm:"type:varchar(256)" json:"avatar"`                       // 头像URL
	BackgroundImage string         `gorm:"type:varchar(256)" json:"background_image"`             // 背景图片URL
	Signature       string         `gorm:"type:varchar(512)" json:"signature"`                    // 个人签名
	IsAdmin         bool           `gorm:"default:false" json:"is_admin"`                         // 是否管理员
	Status          int8           `gorm:"default:1" json:"status"`                               // 状态：1-正常，2-禁用
	LastLoginTime   *time.Time     `json:"last_login_time"`                                       // 最后登录时间
	CreatedAt       time.Time      `json:"created_at"`                                            // 创建时间
	UpdatedAt       time.Time      `json:"updated_at"`                                            // 更新时间
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
}

// TableName 返回用户表名
func (User) TableName() string {
	return TableNameUser
}

// CreateUser 创建新用户
// 参数:
//   - user: 用户信息结构体指针
//
// 返回:
//   - int64: 创建成功返回用户ID
//   - error: 操作错误信息
func CreateUser(user *User) (int64, error) {
	// 检查用户名是否已存在
	var count int64
	if err := DB.Model(&User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrUserAlreadyExists
	}

	// 创建用户记录
	if err := DB.Create(user).Error; err != nil {
		// 处理数据库唯一性冲突（多数据库兼容）
		// Postgres: SQLSTATE 23505；SQLite: code 2067；MySQL: 1062
		errMsg := err.Error()
		if errMsg != "" {
			if contains(errMsg, "duplicate key value") && contains(errMsg, "unique constraint") {
				return 0, ErrUserAlreadyExists // Postgres
			}
			//if contains(errMsg, "UNIQUE constraint failed") {
			//	return 0, ErrUserAlreadyExists // SQLite
			//}
			//if contains(errMsg, "Duplicate entry") {
			//	return 0, ErrUserAlreadyExists // MySQL
			//}
		}
		return 0, ErrCreateUserFailed
	}
	return user.ID, nil

}

// contains 判断字符串是否包含子串
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}

// QueryUserByUsername 通过用户名查询用户信息
// 参数:
//   - username: 用户名
//
// 返回:
//   - *User: 用户信息
//   - error: 操作错误信息
func QueryUserByUsername(username string) (*User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// QueryUserByID 通过用户ID查询用户信息
// 参数:
//   - userID: 用户ID
//
// 返回:
//   - *User: 用户信息
//   - error: 操作错误信息
func QueryUserByID(userID int64) (*User, error) {
	var user User
	result := DB.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// VerifyUser 验证用户名和密码
// 参数:
//   - username: 用户名
//   - password: 密码
//
// 返回:
//   - int64: 验证成功返回用户ID
//   - error: 操作错误信息
func VerifyUser(username, password string) (int64, error) {
	var user User
	result := DB.Where("username = ? AND password = ?", username, password).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrInvalidPassword
		}
		return 0, result.Error
	}

	// 更新最后登录时间
	now := time.Now()
	DB.Model(&user).UpdateColumn("last_login_time", now)

	return user.ID, nil
}

// UpdateUserProfile 更新用户资料
// 参数:
//   - user: 包含更新信息的用户结构体
//
// 返回:
//   - error: 操作错误信息
func UpdateUserProfile(user *User) error {
	// 只允许更新特定字段
	result := DB.Model(&User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"nickname":         user.Nickname,
		"avatar":           user.Avatar,
		"background_image": user.BackgroundImage,
		"signature":        user.Signature,
	})

	if result.Error != nil {
		return ErrUpdateUserFailed
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserPassword 更新用户密码
// 参数:
//   - userID: 用户ID
//   - newPassword: 新密码
//
// 返回:
//   - error: 操作错误信息
func UpdateUserPassword(userID int64, newPassword string) error {
	result := DB.Model(&User{}).Where("id = ?", userID).Update("password", newPassword)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser 删除用户（软删除）
// 参数:
//   - userID: 用户ID
//
// 返回:
//   - error: 操作错误信息
func DeleteUser(userID int64) error {
	result := DB.Delete(&User{}, userID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers 获取用户列表，支持分页
// 参数:
//   - page: 页码
//   - pageSize: 每页记录数
//
// 返回:
//   - []User: 用户列表
//   - int64: 总记录数
//   - error: 操作错误信息
func ListUsers(page, pageSize int) ([]User, int64, error) {
	var users []User
	var total int64

	// 计算总记录数
	if err := DB.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	offset := (page - 1) * pageSize
	if err := DB.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CheckUserExists 检查用户是否存在
// 参数:
//   - userID: 用户ID
//
// 返回:
//   - bool: 用户是否存在
//   - error: 操作错误信息
func CheckUserExists(userID int64) (bool, error) {
	var count int64
	if err := DB.Model(&User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
