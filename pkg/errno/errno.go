package errno

import (
	"fmt"
)

// Errno 定义了自定义错误类型
// 通常包含一个错误码和一个错误消息
// Code: 0 表示成功，非0表示失败
type Errno struct {
	Code    int
	Message string
}

// Error 实现 error 接口
func (e *Errno) Error() string {
	return fmt.Sprintf("Errno - code: %d, message: %s", e.Code, e.Message)
}

// New 创建一个新的 Errno 实例
func New(code int, message string) *Errno {
	return &Errno{
		Code:    code,
		Message: message,
	}
}

// InvalidParameterError 创建一个表示无效参数的错误
// Code: 10001 (示例)
func InvalidParameterError(message string) *Errno {
	if message == "" {
		message = "Invalid parameters"
	}
	return New(10001, message)
}

// DatabaseError 创建一个表示数据库错误的错误
// Code: 10002 (示例)
func DatabaseError(message string) *Errno {
	if message == "" {
		message = "Database error"
	}
	return New(10002, message)
}

// NotFoundError 创建一个表示资源未找到的错误
// Code: 10003 (示例)
func NotFoundError(resourceName string) *Errno {
	message := fmt.Sprintf("%s not found", resourceName)
	return New(10003, message)
}

// Specific error instances (can be expanded)
var (
	// Common errors
	ErrInvalidParameter = InvalidParameterError("") // Default invalid parameter message
	ErrDatabase         = DatabaseError("")         // Default database error message

	// BackgroundInfo specific errors (examples, can be defined per module if needed)
	BackgroundInfoNotFoundError = NotFoundError("BackgroundInfo")
)
