package types

// Error 自定义错误接口
type Error interface {
	error
	// Code 返回错误码
	Code() string
	// Message 返回错误信息
	Message() string
	// Details 返回错误详情
	Details() []interface{}
	// Unwrap 返回原始错误
	Unwrap() error
}

// ErrorCode 错误码类型
type ErrorCode string
