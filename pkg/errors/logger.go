package errors

import "gobase/pkg/errors/codes"

// 日志中间件错误码 (3100-3199)
// NewLogFormatError 创建日志格式化错误
func NewLogFormatError(msg string, cause error) error {
	return NewError(codes.LogFormatError, msg, cause)
}

// NewLogWriteError 创建日志写入错误
func NewLogWriteError(msg string, cause error) error {
	return NewError(codes.LogWriteError, msg, cause)
}

// NewLogFlushError 创建日志刷新错误
func NewLogFlushError(msg string, cause error) error {
	return NewError(codes.LogFlushError, msg, cause)
}

// NewLogRotateError 创建日志轮转错误
func NewLogRotateError(msg string, cause error) error {
	return NewError(codes.LogRotateError, msg, cause)
}

// NewLogBufferError 创建日志缓冲区错误
func NewLogBufferError(msg string, cause error) error {
	return NewError(codes.LogBufferError, msg, cause)
}

// NewLogConfigError 创建日志配置错误
func NewLogConfigError(msg string, cause error) error {
	return NewError(codes.LogConfigError, msg, cause)
}

// NewLogSamplingError 创建日志采样错误
func NewLogSamplingError(msg string, cause error) error {
	return NewError(codes.LogSamplingError, msg, cause)
}

// NewLogMetricsError 创建日志指标错误
func NewLogMetricsError(msg string, cause error) error {
	return NewError(codes.LogMetricsError, msg, cause)
}

// NewLogBodyExceededError 创建日志体超限错误
func NewLogBodyExceededError(msg string, cause error) error {
	return NewError(codes.LogBodyExceededError, msg, cause)
}

// NewLogBufferFullError 创建日志缓冲区满错误
func NewLogBufferFullError(msg string, cause error) error {
	return NewError(codes.LogBufferFullError, msg, cause)
}
