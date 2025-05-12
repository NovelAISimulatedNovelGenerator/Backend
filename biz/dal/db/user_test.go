/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 测试初始化函数，使用SQLite内存数据库
func setupTestDB(t *testing.T) {
	var err error
	// 使用SQLite内存数据库进行测试
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试环境静默日志
	})
	assert.NoError(t, err, "初始化测试数据库失败")

	// 自动创建表结构
	err = DB.AutoMigrate(&User{})
	assert.NoError(t, err, "自动迁移用户表失败")
	
	// 确保每次测试都从空表开始
	DB.Exec("DELETE FROM " + TableNameUser)
}

// 创建测试用户
func createTestUser(t *testing.T) *User {
	// 使用时间戳确保用户名唯一
	timestamp := time.Now().UnixNano()
	username := "testuser" + string(rune(timestamp%26+'a'))
	email := username + "@example.com"
	
	user := &User{
		Username: username,
		Password: "password123",
		Nickname: "测试用户",
		Email:    email,
		Avatar:   "https://example.com/avatar.jpg",
		Status:   0,
		// Unix时间戳（毫秒）
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	id, err := CreateUser(user)
	assert.NoError(t, err, "创建测试用户失败")
	assert.Greater(t, id, int64(0), "用户ID应大于0")
	
	// 重新查询确保有效数据
	createdUser, err := QueryUserByID(id)
	assert.NoError(t, err, "查询创建的用户失败")
	return createdUser
}

// TestCreateUser 测试用户创建
func TestCreateUser(t *testing.T) {
	setupTestDB(t)

	// 测试正常创建用户
	email := "user1@example.com"
	user := &User{
		Username:  "user1",
		Password:  "pass123",
		Nickname:  "用户1",
		Email:     email,
		Status:    0,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	id, err := CreateUser(user)
	assert.NoError(t, err, "创建用户失败")
	assert.Greater(t, id, int64(0), "用户ID应大于0")

	// 测试创建重复用户名
	duplicateUser := &User{
		Username: "user1",
		Password: "anotherpass",
		Nickname: "重复用户",
	}

	_, err = CreateUser(duplicateUser)
	assert.Error(t, err, "应检测到重复用户名")
	assert.Equal(t, ErrUserAlreadyExists, err, "错误类型应为ErrUserAlreadyExists")
}

// TestQueryUserByUsername 测试通过用户名查询
func TestQueryUserByUsername(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 测试查询存在的用户
	user, err := QueryUserByUsername(originalUser.Username)
	assert.NoError(t, err, "查询存在的用户失败")
	assert.Equal(t, originalUser.Username, user.Username, "用户名应匹配")
	assert.Equal(t, originalUser.Email, user.Email, "邮箱应匹配")

	// 测试查询不存在的用户
	_, err = QueryUserByUsername("nonexistentuser")
	assert.Error(t, err, "查询不存在的用户应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")
}

// TestQueryUserByID 测试通过ID查询
func TestQueryUserByID(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 测试查询存在的用户ID
	user, err := QueryUserByID(originalUser.ID)
	assert.NoError(t, err, "查询存在的用户ID失败")
	assert.Equal(t, originalUser.ID, user.ID, "用户ID应匹配")
	assert.Equal(t, originalUser.Username, user.Username, "用户名应匹配")

	// 测试查询不存在的用户ID
	_, err = QueryUserByID(9999)
	assert.Error(t, err, "查询不存在的用户ID应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")
}

// TestVerifyUser 测试用户验证
func TestVerifyUser(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 测试正确凭据
	id, err := VerifyUser(originalUser.Username, originalUser.Password)
	assert.NoError(t, err, "验证正确凭据失败")
	assert.Equal(t, originalUser.ID, id, "返回的用户ID应匹配")

	// 测试错误密码
	_, err = VerifyUser(originalUser.Username, "wrongpassword")
	assert.Error(t, err, "验证错误密码应返回错误")
	assert.Equal(t, ErrInvalidPassword, err, "错误类型应为ErrInvalidPassword")

	// 测试不存在的用户
	_, err = VerifyUser("nonexistentuser", "anypassword")
	assert.Error(t, err, "验证不存在的用户应返回错误")
	assert.Equal(t, ErrInvalidPassword, err, "错误类型应为ErrInvalidPassword")
}

// TestUpdateUserProfile 测试更新用户资料
func TestUpdateUserProfile(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 更新用户资料
	updatedUser := &User{
		ID:       originalUser.ID,
		Nickname: "更新后的昵称",
		Avatar:   "https://example.com/new-avatar.jpg",
		Email:    "updated_" + originalUser.Email,
		Status:   originalUser.Status,
	}

	err := UpdateUserProfile(updatedUser)
	assert.NoError(t, err, "更新用户资料失败")

	// 验证更新结果
	user, err := QueryUserByID(originalUser.ID)
	assert.NoError(t, err, "查询更新后的用户失败")
	assert.Equal(t, updatedUser.Nickname, user.Nickname, "昵称应已更新")
	assert.Equal(t, updatedUser.Avatar, user.Avatar, "头像应已更新")
	assert.Equal(t, updatedUser.Email, user.Email, "邮箱应已更新")
	assert.Equal(t, updatedUser.Status, user.Status, "状态应已更新")
	assert.Equal(t, originalUser.Username, user.Username, "用户名不应变化")
	assert.Equal(t, originalUser.Password, user.Password, "密码不应变化")

	// 测试更新不存在的用户
	nonExistentUser := &User{
		ID:       9999,
		Nickname: "不存在的用户",
	}
	err = UpdateUserProfile(nonExistentUser)
	assert.Error(t, err, "更新不存在的用户应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")
}

// TestUpdateUserPassword 测试更新用户密码
func TestUpdateUserPassword(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)
	newPassword := "newpassword123"

	// 更新密码
	err := UpdateUserPassword(originalUser.ID, newPassword)
	assert.NoError(t, err, "更新密码失败")

	// 验证新密码
	id, err := VerifyUser(originalUser.Username, newPassword)
	assert.NoError(t, err, "验证新密码失败")
	assert.Equal(t, originalUser.ID, id, "验证后的用户ID应匹配")

	// 测试更新不存在的用户密码
	err = UpdateUserPassword(9999, "anypassword")
	assert.Error(t, err, "更新不存在的用户密码应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")
}

// TestDeleteUser 测试删除用户
func TestDeleteUser(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 删除用户
	err := DeleteUser(originalUser.ID)
	assert.NoError(t, err, "删除用户失败")

	// 验证用户已删除
	_, err = QueryUserByID(originalUser.ID)
	assert.Error(t, err, "查询已删除的用户应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")

	// 测试删除不存在的用户
	err = DeleteUser(9999)
	assert.Error(t, err, "删除不存在的用户应返回错误")
	assert.Equal(t, ErrUserNotFound, err, "错误类型应为ErrUserNotFound")
}

// TestListUsers 测试用户列表查询
func TestListUsers(t *testing.T) {
	setupTestDB(t)
	
	// 创建多个测试用户
	for i := 0; i < 10; i++ {
		timestamp := time.Now().UnixNano() + int64(i)
		username := "listuser" + string(rune('a'+i)) + string(rune(timestamp%10+'0'))
		email := username + "@example.com"
		user := &User{
			Username: username,
			Password: "password",
			Nickname: "列表用户" + string(rune('0'+i)),
			Email:    email + "_" + strconv.Itoa(i),
			Status:   0,
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		}
		_, err := CreateUser(user)
		assert.NoError(t, err, "创建测试用户失败")
	}

	// 测试第一页
	users, total, err := ListUsers(1, 5)
	assert.NoError(t, err, "获取用户列表失败")
	assert.Equal(t, int64(10), total, "总用户数应为10")
	assert.Len(t, users, 5, "第一页应有5条记录")

	// 测试第二页
	users, total, err = ListUsers(2, 5)
	assert.NoError(t, err, "获取第二页用户列表失败")
	assert.Equal(t, int64(10), total, "总用户数应为10")
	assert.Len(t, users, 5, "第二页应有5条记录")

	// 测试超出范围的页码
	users, total, err = ListUsers(3, 5)
	assert.NoError(t, err, "获取超出范围的页码应成功但无数据")
	assert.Equal(t, int64(10), total, "总用户数应为10")
	assert.Len(t, users, 0, "超出范围的页码应返回空列表")
}

// TestCheckUserExists 测试检查用户是否存在
func TestCheckUserExists(t *testing.T) {
	setupTestDB(t)
	originalUser := createTestUser(t)

	// 测试存在的用户
	exists, err := CheckUserExists(originalUser.ID)
	assert.NoError(t, err, "检查存在的用户失败")
	assert.True(t, exists, "存在的用户应返回true")

	// 测试不存在的用户
	exists, err = CheckUserExists(9999)
	assert.NoError(t, err, "检查不存在的用户失败")
	assert.False(t, exists, "不存在的用户应返回false")
}
