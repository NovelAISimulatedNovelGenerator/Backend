// Package middleware 提供JWT中间件，统一令牌生成、验证、API签名，集成用户、权限、登录等模块
// 依赖hertz-contrib/jwt，符合Hertz最佳实践
// 只需在路由注册时调用JwtMiddleware()，即可实现登录自动生成JWT、接口自动校验，无需手写token逻辑

// Package middleware 提供 JWT 统一中间件入口，仅暴露 JwtMiddleware 与 IdentityKey
// 其余实现均隐藏于 jwt 子包，确保低耦合高内聚
package middleware

import (
	jwtImpl "novelai/pkg/middleware/jwt"
	"time"

	"github.com/hertz-contrib/jwt"
)

// IdentityKey 用户唯一标识字段，对外暴露常量
var IdentityKey = jwtImpl.IdentityKey

// JwtMiddleware 返回配置好的 JWT 中间件实例
// 只负责组装参数，具体实现隐藏于 jwt 子包
func JwtMiddleware() (*jwt.HertzJWTMiddleware, error) {
	return jwt.New(&jwt.HertzJWTMiddleware{
		Realm:           jwtImpl.JwtRealm,
		Key:             []byte(jwtImpl.JwtKey),
		Timeout:         time.Hour * jwtImpl.JwtTimeout,
		MaxRefresh:      time.Hour * jwtImpl.JwtMaxRefresh,
		IdentityKey:     jwtImpl.IdentityKey,
		PayloadFunc:     jwtImpl.PayloadFunc(),
		Authenticator:   jwtImpl.Authenticator(),
		Authorizator:    jwtImpl.Authorizator(),
		Unauthorized:    jwtImpl.Unauthorized(),
		LoginResponse:   jwtImpl.LoginResponse(),
		RefreshResponse: jwtImpl.RefreshResponse(),
	})
}

// 使用说明：
// 1. 在路由注册时：
//    jwtMw, _ := middleware.JwtMiddleware()
//    group.POST("/login", jwtMw.LoginHandler)
//    group.Use(jwtMw.MiddlewareFunc())
// 2. handler 通过 c.Get(middleware.IdentityKey) 获取用户信息
// 3. 权限、用户等可扩展 jwt 子包实现
