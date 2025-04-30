// 已废弃：自定义 token 逻辑已迁移至 JWT 中间件
// 此文件保留空壳，便于后续安全移除

package crypto

import (
	"errors"
	"time"

	tokenpkg "novelai/pkg/utils/crypto/token"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateToken 生成 JWT Token
// 参数 claimsData: 业务自定义数据（map[string]interface{}）
// 参数 secret: 签名密钥
// 参数 expireSeconds: token 有效期（秒）
// 返回 token 字符串和错误信息
func GenerateToken(claimsData map[string]interface{}, secret string, expireSeconds int64) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("签名密钥不能为空")
	}
	if expireSeconds <= 0 {
		expireSeconds = tokenpkg.DefaultTokenExpireSeconds
	}
	claims := tokenpkg.CustomClaims{
		Data: claimsData,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod(tokenpkg.TokenSigningMethod), claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// ParseToken 解析并校验 token
// 参数 tokenStr: token 字符串
// 参数 secret: 签名密钥
// 返回 claims 数据和错误信息
func ParseToken(tokenStr string, secret string) (map[string]interface{}, error) {
	if len(tokenStr) == 0 {
		return nil, errors.New("token 不能为空")
	}
	if len(secret) == 0 {
		return nil, errors.New("签名密钥不能为空")
	}
	claims := &tokenpkg.CustomClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != tokenpkg.TokenSigningMethod {
			return nil, errors.New("签名算法不匹配")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims.Data, nil
}

// IsTokenValid 校验 token 是否有效
// 参数 tokenStr: token 字符串
// 参数 secret: 签名密钥
// 返回是否有效和错误信息
func IsTokenValid(tokenStr string, secret string) (bool, error) {
	if len(tokenStr) == 0 || len(secret) == 0 {
		return false, errors.New("token 或密钥不能为空")
	}
	claims := &tokenpkg.CustomClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != tokenpkg.TokenSigningMethod {
			return nil, errors.New("签名算法不匹配")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return false, errors.New("token 已过期")
	}
	return true, nil
}

// RefreshToken 刷新 token（生成新 token，过期时间重置）
// 参数 tokenStr: 旧 token 字符串
// 参数 secret: 签名密钥
// 参数 expireSeconds: 新 token 有效期（秒）
// 返回新 token 字符串和错误信息
func RefreshToken(tokenStr string, secret string, expireSeconds int64) (string, error) {
	if len(tokenStr) == 0 || len(secret) == 0 {
		return "", errors.New("token 或密钥不能为空")
	}
	claims := &tokenpkg.CustomClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != tokenpkg.TokenSigningMethod {
			return nil, errors.New("签名算法不匹配")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	// 重新生成 token，重置过期时间
	return GenerateToken(claims.Data, secret, expireSeconds)
}
