package redis

import (
	"context"
	"strings"
	"time"

	"gobase/pkg/errors"
)

// isRetryableError 判断错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络相关错误通常是可重试的
	if isNetworkError(err) {
		return true
	}

	// READONLY 错误（集群模式下可能发生）是可重试的
	if isReadOnlyError(err) {
		return true
	}

	// CLUSTERDOWN 错误是可重试的
	if isClusterDownError(err) {
		return true
	}

	// LOADING 错误是可重试的
	if isLoadingError(err) {
		return true
	}

	return false
}

// withRetry 包装重试逻辑
func withRetry(ctx context.Context, options *Options, op func() error) error {
	var lastErr error
	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		// 首先检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			if err == context.DeadlineExceeded {
				return errors.NewTimeoutError("operation timed out", err)
			}
			return errors.NewRedisCommandError("operation cancelled", err)
		}

		// 执行操作
		err := op()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是上下文超时，直接返回超时错误
		if err == context.DeadlineExceeded ||
			strings.Contains(err.Error(), "context deadline exceeded") {
			return errors.NewTimeoutError("operation timed out", err)
		}

		// 如果已经是包装过的错误，直接返回
		if errors.HasErrorCode(err, "") {
			return err
		}

		// 检查错误是否可重试
		if !isRetryableError(err) {
			return errors.NewRedisCommandError("operation failed", err)
		}

		// 如果是最后一次尝试，返回适当的错误
		if attempt == options.MaxRetries {
			return errors.NewRedisCommandError("max retries exceeded", lastErr)
		}

		// 计算退避时间
		backoff := options.RetryBackoff
		if backoff == 0 {
			backoff = time.Duration(attempt*attempt) * 100 * time.Millisecond
		}

		// 等待退避时间，但要考虑上下文超时
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return errors.NewTimeoutError("operation timed out during retry", ctx.Err())
			}
			return errors.NewRedisCommandError("operation cancelled during retry", ctx.Err())
		case <-time.After(backoff):
			continue
		}
	}

	return errors.NewRedisCommandError("max retries exceeded", lastErr)
}

// isNetworkError 判断是否为网络错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrPoolTimeout) || errors.Is(err, ErrConnPool) {
		return true
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "network")
}

// isLoadingError 判断是否为加载数据错误
func isLoadingError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "LOADING")
}

// isReadOnlyError 判断是否为只读错误
func isReadOnlyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "READONLY")
}

// isClusterDownError 判断是否为集群不可用错误
func isClusterDownError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "CLUSTERDOWN")
}
