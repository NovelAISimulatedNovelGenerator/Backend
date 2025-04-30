// auth.go
// JWT 认证、授权、响应相关实现
package jwt

import (
	"context"

	"time"

	"novelai/biz/dal/db"
	userpb "novelai/biz/model/user"
	"novelai/pkg/constants"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
	"novelai/pkg/utils/crypto"
)

// Authenticator 返回 JWT Authenticator 实现
// 用于 hertz-contrib/jwt 中间件配置，负责登录认证逻辑
// 返回一个闭包，签名为 func(ctx context.Context, c *app.RequestContext) (interface{}, error)
// ctx: 上下文，c: hertz 请求上下文
// 返回值：认证通过时返回用户相关数据（如 user_id），否则返回错误
func Authenticator() func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
	return authenticator
}

// authenticator 登录认证实现
// 1. 解析请求体，获取用户名和密码
// 2. 对密码进行 MD5 哈希
// 3. 调用 db.VerifyUser 校验用户名密码
// 4. 校验通过返回用户 user_id，失败返回错误
func authenticator(ctx context.Context, c *app.RequestContext) (interface{}, error) {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return nil, jwt.ErrMissingLoginValues
	}
	req.Password = crypto.HashPassword(req.Password)
	userId, err := db.VerifyUser(req.Username, req.Password)
	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	c.Set(IdentityKey, userId)
	return map[string]interface{}{IdentityKey: userId}, nil
}


// PayloadFunc 返回 JWT PayloadFunc 实现
// 用于 hertz-contrib/jwt 中间件配置，负责生成 JWT token 的 claims 数据
// 返回一个闭包，签名为 func(data interface{}) jwt.MapClaims
// data: 登录认证通过后的用户数据
// 返回值：JWT claims，包含用户唯一标识、权限等
func PayloadFunc() func(data interface{}) jwt.MapClaims {
	return payloadFunc
}

// payloadFunc JWT claims 生成实现
// 1. 若 data 为 map[string]interface{}，则提取 IdentityKey 和 role 字段
// 2. 返回 jwt.MapClaims，供 JWT token 使用
func payloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		return jwt.MapClaims{
			IdentityKey: v[IdentityKey],
			"role":      v["role"],
		}
	}
	return jwt.MapClaims{}
}

// Unauthorized 返回 JWT Unauthorized 实现
// 用于 hertz-contrib/jwt 中间件配置，未授权时响应
// 返回一个闭包，签名为 func(ctx context.Context, c *app.RequestContext, code int, message string)
// ctx: 上下文，c: hertz 请求上下文，code: 状态码，message: 错误信息
func Unauthorized() func(ctx context.Context, c *app.RequestContext, code int, message string) {
	return unauthorized
}

// unauthorized 未授权响应实现
// 1. 返回 JSON 格式的错误信息，包含 code 和 message 字段
func unauthorized(ctx context.Context, c *app.RequestContext, code int, message string) {
	c.JSON(constants.StatusUnauthorized, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

// LoginResponse 返回 JWT LoginResponse 实现
// 用于 hertz-contrib/jwt 中间件配置，登录成功时响应
// 返回一个闭包，签名为 func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time)
// ctx: 上下文，c: hertz 请求上下文，code: 状态码，token: JWT token，expire: 过期时间
func LoginResponse() func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
	return loginResponse
}

// loginResponse 登录成功响应实现
// 1. 从 context 获取 user_id
// 2. 返回 LoginResponse 结构体，包含 code、message、user_id、token
func loginResponse(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
	idVal, _ := c.Get(IdentityKey)
	userId := idVal.(int64)
	resp := &userpb.LoginResponse{
		Code:    constants.StatusOK,
		Message: "登录成功",
		UserId:  userId,
		Token:   token,
	}
	c.JSON(constants.StatusOK, resp)
}

// RefreshResponse 返回 JWT RefreshResponse 实现
// 用于 hertz-contrib/jwt 中间件配置，刷新 token 时响应
// 返回一个闭包，签名为 func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time)
// ctx: 上下文，c: hertz 请求上下文，code: 状态码，token: JWT token，expire: 过期时间
func RefreshResponse() func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
	return refreshResponse
}

// refreshResponse 刷新 token 响应实现
// 1. 返回 JSON 格式的成功信息，包含 code、message、token 字段
func refreshResponse(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
	c.JSON(constants.StatusOK, map[string]interface{}{
		"code":    constants.StatusOK,
		"message": "刷新成功",
		"token":   token,
	})
}
