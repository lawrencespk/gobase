package errors

import (
	"context"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts     int           // 最大重试次数
	InitialInterval time.Duration // 初始重试间隔
	MaxInterval     time.Duration // 最大重试间隔
	Multiplier      float64       // 重试间隔增长因子
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:     3,                      // 最大重试次数
	InitialInterval: 100 * time.Millisecond, // 初始重试间隔
	MaxInterval:     2 * time.Second,        // 最大重试间隔
	Multiplier:      2.0,                    // 重试间隔增长因子
}

// Retryable 定义可重试的操作
type Retryable func(ctx context.Context) error

// WithRetry 执行可重试操作
func WithRetry(ctx context.Context, op Retryable, opts ...RetryConfig) error {
	// 使用默认配置或自定义配置
	config := DefaultRetryConfig
	if len(opts) > 0 {
		config = opts[0]
	}

	var lastErr error
	interval := config.InitialInterval

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// 检查上下文是否取消
		if ctx.Err() != nil {
			return WrapWithCode(ctx.Err(), "RetryContextCanceled",
				"重试操作被取消")
		}

		// 执行操作
		if err := op(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			// 如果是不可重试的错误，直接返回
			if IsSystemError(err) {
				return WrapWithCode(err, "RetrySystemError",
					"遇到系统错误，终止重试")
			}
		}

		// 如果不是最后一次尝试，则等待后重试
		if attempt < config.MaxAttempts-1 {
			// 使用带有超时的 context 进行等待
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return WrapWithCode(ctx.Err(), "RetryContextCanceled",
					"重试等待被取消")
			case <-timer.C:
				// 增加重试间隔，但不超过最大间隔
				interval = time.Duration(float64(interval) * config.Multiplier)
				if interval > config.MaxInterval {
					interval = config.MaxInterval
				}
			}
		}
	}

	return WrapWithCode(lastErr, "RetryExhausted",
		"达到最大重试次数")
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	// 系统错误通常不可重试
	if IsSystemError(err) {
		return false
	}

	// 检查特定的错误码
	code := GetErrorCode(err)
	switch code {
	case "TimeoutError", // 超时错误
		"NetworkError",       // 网络错误
		"CacheError",         // 缓存错误
		"DBConnError",        // 数据库连接错误
		"ServiceUnavailable": // 服务不可用
		return true
	}

	return false
}
