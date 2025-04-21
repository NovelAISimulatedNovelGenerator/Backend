// http_status.go
// HTTP状态码常量定义，统一管理所有业务响应的HTTP状态码
// 便于维护和扩展
package constants

const (
	// 请求成功
	StatusOK = 200
	// 请求参数错误
	StatusBadRequest = 400
	// 未授权
	StatusUnauthorized = 401
	// 禁止访问
	StatusForbidden = 403
	// 资源未找到
	StatusNotFound = 404
	// 服务器内部错误
	StatusInternalServerError = 500
)
