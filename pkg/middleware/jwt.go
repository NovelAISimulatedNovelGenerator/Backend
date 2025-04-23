// Package middleware 提供JWT中间件，统一令牌生成、验证、API签名，集成用户、权限、登录等模块
// 依赖hertz-contrib/jwt，符合Hertz最佳实践
// 只需在路由注册时调用JwtMiddleware()，即可实现登录自动生成JWT、接口自动校验，无需手写token逻辑

package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

// IdentityKey JWT中保存的用户唯一标识
const IdentityKey = "user_id"

// JwtMiddleware 返回配置好的JWT中间件实例
// 用于自动生成、校验JWT令牌，实现API签名和权限集成
func JwtMiddleware() (*jwt.HertzJWTMiddleware, error) {
	return jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       "novelai zone", // 认证领域
		Key:         []byte("change-this-to-a-strong-secret-key"), // 签名密钥，需配置为不可逆安全字符串
		Timeout:     time.Hour * 24, // token有效期
		MaxRefresh:  time.Hour * 24, // 刷新时间
		IdentityKey: IdentityKey,
		// PayloadFunc 定义JWT载荷内容，可扩展用户ID、权限等
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(map[string]interface{}); ok {
				return jwt.MapClaims{
					IdentityKey: v[IdentityKey],
					"role":      v["role"], // 可扩展权限字段
				}
			}
			return jwt.MapClaims{}
		},
		// Authenticator 登录认证逻辑，集成用户模块
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.Bind(&req); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}
			// TODO: 调用用户模块校验用户名和密码，返回用户ID和权限
			// 例如：user, err := db.VerifyUser(req.Username, req.Password)
			// if err != nil { return nil, jwt.ErrFailedAuthentication }
			// return map[string]interface{}{IdentityKey: user.ID, "role": user.Role}, nil
			return map[string]interface{}{IdentityKey: 1, "role": "user"}, nil // 示例
		},
		// Authorizator 权限校验逻辑，可扩展
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			// 可根据data中的role字段实现权限控制
			return true // 示例：全部通过
		},
		// Unauthorized 未授权响应
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(401, map[string]interface{}{
				"code":    code,
				"message": message,
			})
		},
	})
}

// 使用说明：
// 1. 在路由注册时：
//    jwtMw, _ := middleware.JwtMiddleware()
//    group.POST("/login", jwtMw.LoginHandler) // 登录自动生成token
//    group.Use(jwtMw.MiddlewareFunc())         // 受保护接口自动校验token
// 2. handler通过c.Get(middleware.IdentityKey)获取用户信息
// 3. 权限、用户等模块可扩展PayloadFunc/Authorizator实现更复杂业务
