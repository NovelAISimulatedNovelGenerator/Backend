package token

// token_const.go 用于存放 token 相关常量，便于维护和复用。

const (
	// Token 默认过期时间（秒）
	DefaultTokenExpireSeconds int64 = 86400 // 24 小时

	// Token 签名算法
	TokenSigningMethod = "HS256"
)
