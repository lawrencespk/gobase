package types

// Error 自定义错误接口
type Error interface {
	error

	Code() string                             // Code 返回错误码
	Message() string                          // Message 返回错误信息
	Details() []interface{}                   // Details 返回错误详情
	Unwrap() error                            // Unwrap 返回原始错误
	Stack() []string                          // Stack获取错误堆栈
	WithDetails(details ...interface{}) Error // 添加错误详情
}

// ErrorCode 错误码类型
type ErrorCode string
