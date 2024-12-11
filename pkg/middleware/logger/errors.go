package logger

import (
	"fmt"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// 创建错误构造函数
func newLogBufferError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogBufferError, msg), cause)
}

func newLogFlushError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogFlushError, msg), cause)
}

// newLogRotateError 预留给日志轮转功能使用
//
//lint:ignore U1000 预留给日志轮转功能使用
func newLogRotateError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogRotateError, msg), cause)
}

// newLogFormatError 预留给日志格式化功能使用
//
//lint:ignore U1000 预留给日志格式化功能使用
func newLogFormatError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogFormatError, msg), cause)
}

func newLogWriteError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogWriteError, msg), cause)
}

// newLogConfigError 预留给配置验证功能使用
//
//lint:ignore U1000 预留给配置验证功能使用
func newLogConfigError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogConfigError, msg), cause)
}

// newLogSamplingError 预留给采样功能使用
//
//lint:ignore U1000 预留给采样功能使用
func newLogSamplingError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogSamplingError, msg), cause)
}

// newLogMetricsError 预留给指标收集功能使用
//
//lint:ignore U1000 预留给指标收集功能使用
func newLogMetricsError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogMetricsError, msg), cause)
}

// newLogBodyExceededError 预留给请求体大小限制功能使用
//
//lint:ignore U1000 预留给请求体大小限制功能使用
func newLogBodyExceededError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogBodyExceededError, msg), cause)
}

func newLogBufferFullError(msg string, cause error) error {
	return errors.NewMiddlewareError(fmt.Sprintf("[%s] %s", codes.LogBufferFullError, msg), cause)
}
