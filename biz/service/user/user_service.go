/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package user

import (
	"context"

	"novelai/biz/dal/db"
	"novelai/biz/model/user"
	"novelai/pkg/utils/crypto"

	"github.com/cloudwego/hertz/pkg/app"
)

// UserService 用户服务结构体
// 负责处理所有与用户相关的业务逻辑
type UserService struct {
	ctx context.Context
	c   *app.RequestContext
}

// NewUserService 创建用户服务实例
// 参数:
//   - ctx: 上下文
//   - c: 请求上下文
//
// 返回:
//   - *UserService: 用户服务实例
func NewUserService(ctx context.Context, c *app.RequestContext) *UserService {
	return &UserService{ctx: ctx, c: c}
}

// generatePasswordHash 生成MD5密码哈希（调用通用加密模块）
// 参数: password 明文密码
// 返回: 加密后的字符串（32位小写MD5）
func generatePasswordHash(password string) string {
	return crypto.HashPassword(password)
}

// Register 处理用户注册业务逻辑（已重构，token 交由 JWT 中间件统一生成）
// 参数:
//   - req: 注册请求
// 返回:
//   - userId: 用户ID
//   - error: 操作错误信息
func (s *UserService) Register(req *user.RegisterRequest) (userId int64, err error) {
	// 检查用户名是否已存在
	existUser, err := db.QueryUserByUsername(req.Username)
	if err != nil && err != db.ErrUserNotFound {
		return 0, err
	}
	if existUser != nil {
		return 0, db.ErrUserAlreadyExists
	}

	// 密码加密
	passwordHash := generatePasswordHash(req.Password)

	// 创建用户记录
	newUser := &db.User{
		Username: req.Username,
		Password: passwordHash, // service层统一加密
		Nickname: req.Nickname,
		Email:    req.Email,
		Status:   0, // 默认状态：正常
	}

	// 调用数据库层创建用户
	userId, err = db.CreateUser(newUser)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

// Login 处理用户登录业务逻辑（已重构，token 交由 JWT 中间件统一生成）
// 参数:
//   - req: 登录请求
// 返回:
//   - userId: 用户ID
//   - error: 操作错误信息
func (s *UserService) Login(req *user.LoginRequest) (userId int64, err error) {
	// 调用数据库层验证用户名和密码
	// 注意：密码已在handler层加密
	userId, err = db.VerifyUser(req.Username, req.Password)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

// GetUserInfo 获取用户信息
// 参数:
//   - userId: 目标用户ID
//
// 返回:
//   - *user.User: 用户信息
//   - error: 操作错误信息
func (s *UserService) GetUserInfo(userId int64) (*user.User, error) {
	// 查询用户基本信息
	dbUser, err := db.QueryUserByID(userId)
	if err != nil {
		return nil, err
	}

	// 构建用户对象
	userInfo := &user.User{
		Id:        dbUser.ID,
		Username:  dbUser.Username,
		Nickname:  dbUser.Nickname,
		Email:     dbUser.Email,
		Avatar:    dbUser.Avatar,
		Status:    dbUser.Status,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}

	return userInfo, nil
}

// UpdateUserProfile 更新用户资料
// 参数:
//   - userId: 用户ID
//   - req: 更新请求
//
// 返回:
//   - error: 操作错误信息
func (s *UserService) UpdateUserProfile(userId int64, req *user.UpdateUserRequest) error {
	// 首先检查用户是否存在
	exists, err := db.CheckUserExists(userId)
	if err != nil {
		return err
	}
	if !exists {
		return db.ErrUserNotFound
	}

	// 构建更新对象
	updateUser := &db.User{
		ID:       userId,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Email:    req.Email,
	}

	// 调用数据库层更新用户资料
	return db.UpdateUserProfile(updateUser)
}

// UpdateUserPassword 更新用户密码
// 参数:
//   - userId: 用户ID
//   - oldPassword: 旧密码(已加密)
//   - newPassword: 新密码(已加密)
//
// 返回:
//   - error: 操作错误信息
func (s *UserService) UpdateUserPassword(userId int64, oldPassword, newPassword string) error {
	// 首先查询用户信息
	user, err := db.QueryUserByID(userId)
	if err != nil {
		return err
	}

	// 验证旧密码是否正确
	_, err = db.VerifyUser(user.Username, oldPassword)
	if err != nil {
		return err
	}

	// 调用数据库层更新密码
	return db.UpdateUserPassword(userId, newPassword)
}

// DeleteUser 软删除用户
// 参数: userId
// 返回: error
func (s *UserService) DeleteUser(userId int64) error {
	// 检查用户是否存在
	exists, err := db.CheckUserExists(userId)
	if err != nil {
		return err
	}
	if !exists {
		return db.ErrUserNotFound
	}
	// 执行软删除
	return db.DeleteUser(userId)
}

// ListUsers 获取用户列表
// 参数:
//   - page: 页码
//   - pageSize: 每页记录数
//
// 返回:
//   - []*user.User: 用户列表
//   - int64: 总记录数
//   - error: 操作错误信息
func (s *UserService) ListUsers(page, pageSize int) ([]*user.User, int64, error) {
	// 调用数据库层获取用户列表
	dbUsers, total, err := db.ListUsers(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为API响应格式
	users := make([]*user.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		userInfo := &user.User{
			Id:        dbUser.ID,
			Username:  dbUser.Username,
			Nickname:  dbUser.Nickname,
			Email:     dbUser.Email,
			Avatar:    dbUser.Avatar,
			Status:    dbUser.Status,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		}

		users = append(users, userInfo)
	}

	return users, total, nil
}
