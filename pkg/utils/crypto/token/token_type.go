package token

import (
	"github.com/golang-jwt/jwt/v4"
)

// CustomClaims 定义通用的 JWT Claims 结构体，便于扩展。
type CustomClaims struct {
	Data map[string]interface{} `json:"data"`
	jwt.RegisteredClaims
}
