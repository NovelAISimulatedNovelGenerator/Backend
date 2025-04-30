// 已废弃：自定义 token 逻辑已迁移至 JWT 中间件
// 此文件保留空壳，便于后续安全移除

package token

// token_const.go 用于存放 token 相关常量，便于维护和复用。

const (
	// Token 默认过期时间（秒）
	DefaultTokenExpireSeconds int64 = 86400 // 24 小时

	// Token 签名算法
	TokenSigningMethod = "HS256"
)
