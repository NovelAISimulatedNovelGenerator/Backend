// Package constants 错误常量定义
package constants

import "errors"

// 用户相关错误常量
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户名已存在")
	ErrInvalidPassword   = errors.New("密码验证失败")
	ErrCreateUserFailed  = errors.New("创建用户失败")
	ErrUpdateUserFailed  = errors.New("更新用户信息失败")
)
