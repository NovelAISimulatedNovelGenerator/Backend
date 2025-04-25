// type.go
// JWT 相关类型定义
package jwt

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
