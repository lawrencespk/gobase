package redis

import (
	"context"
	"strings"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/go-redis/redis/v8"
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
	if strings.Contains(err.Error(), "LOADING") {
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
				return errors.NewError(codes.TimeoutError, "operation timed out", err)
			}
			return errors.NewError(codes.CacheError, "operation cancelled", err)
		}

		// 执行操作
		err := op()
		if err == nil {
			return nil
		}

		lastErr = err

		// 1. 检查是否已经是包装过的错误
		if errors.HasErrorCode(err, "") {
			return err
		}

		// 2. 检查是否是上下文超时
		if err == context.DeadlineExceeded || strings.Contains(err.Error(), "context deadline exceeded") {
			return errors.NewError(codes.TimeoutError, "operation timed out", err)
		}

		// 3. 检查是否可重试
		if !isRetryableError(err) {
			return errors.NewError(codes.CacheError, "operation failed", err)
		}

		// 4. 如果是最后一次尝试，直接返回错误
		if attempt == options.MaxRetries {
			return errors.NewError(codes.CacheError, "max retries exceeded", err)
		}

		// 5. 计算退避时间
		backoff := options.RetryBackoff
		if backoff == 0 {
			backoff = time.Duration(attempt*attempt) * 100 * time.Millisecond
		}

		// 6. 等待退避时间，但要考虑上下文超时
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return errors.NewError(codes.TimeoutError, "operation timed out during retry", ctx.Err())
			}
			return errors.NewError(codes.CacheError, "operation cancelled during retry", ctx.Err())
		case <-time.After(backoff):
			continue
		}
	}

	return errors.NewError(codes.CacheError, "max retries exceeded", lastErr)
}

// isNetworkError 判断是否为网络错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为网络超时错误
	if errors.Is(err, ErrPoolTimeout) || errors.Is(err, ErrConnPool) || err == redis.ErrClosed {
		return true
	}

	// 检查是否为连接错误
	if strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "network") {
		return true
	}

	return false
}

// isReadOnlyError 判断是否为只读错误
func isReadOnlyError(err error) bool {
	if err == nil {
		return false
	}

	// READONLY 错误在集群模式下表示节点处于只读状态
	return strings.Contains(err.Error(), "READONLY")
}

// isClusterDownError 判断是否为集群不可用错误
func isClusterDownError(err error) bool {
	if err == nil {
		return false
	}

	// CLUSTERDOWN 错误表示集群不可用
	return strings.Contains(err.Error(), "CLUSTERDOWN")
}
