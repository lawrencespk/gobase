package errors

import (
	"gobase/pkg/errors/codes"
)

// NewMiddlewareError 创建中间件错误
func NewMiddlewareError(message string, cause error) error {
	return NewError(codes.MiddlewareError, message, cause)
}

// NewAuthMiddlewareError 创建认证中间件错误
func NewAuthMiddlewareError(message string, cause error) error {
	return NewError(codes.AuthError, message, cause)
}

// NewRateLimitError 创建限流中间件错误
func NewRateLimitError(message string, cause error) error {
	return NewError(codes.RateLimitError, message, cause)
}

// NewTimeoutMiddlewareError 创建超时中间件错误
func NewTimeoutMiddlewareError(message string, cause error) error {
	return NewError(codes.TimeoutMWError, message, cause)
}

// NewCORSError 创建跨域中间件错误
func NewCORSError(message string, cause error) error {
	return NewError(codes.CORSError, message, cause)
}

// NewTracingError 创建追踪中间件错误
func NewTracingError(message string, cause error) error {
	return NewError(codes.TracingError, message, cause)
}

// NewMetricsError 创建指标中间件错误
func NewMetricsError(message string, cause error) error {
	return NewError(codes.MetricsError, message, cause)
}

// NewLoggingError 创建日志中间件错误
func NewLoggingError(message string, cause error) error {
	return NewError(codes.LoggingError, message, cause)
}
