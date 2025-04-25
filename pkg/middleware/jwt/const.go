// const.go
// JWT 相关常量定义
package jwt

const (
	JwtRealm      = "novelai zone"
	JwtKey        = "change-this-to-a-strong-secret-key"
	JwtTimeout    = 24 // 单位：小时
	JwtMaxRefresh = 24 // 单位：小时
	IdentityKey   = "user_id" // 必须大写导出，供外部访问
)
