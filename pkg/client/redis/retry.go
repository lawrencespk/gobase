package redis

import (
	"context"
	"strings"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"

	"github.com/go-redis/redis/v8"
)

// retryStrategy 重试策略
type retryStrategy struct {
	ctx        context.Context
	logger     types.Logger
	maxRetries int
	retryDelay time.Duration
}

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

	return false
}

// withRetry 包装重试逻辑
func withRetry(ctx context.Context, options *Options, op func() error) error {
	var lastErr error
	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		if err := op(); err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return err
			}

			// 如果不是最后一次尝试，则等待后重试
			if attempt < options.MaxRetries {
				// 使用指数退避策略
				backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
				timer := time.NewTimer(backoff)
				select {
				case <-ctx.Done():
					timer.Stop()
					return errors.NewError(codes.CacheError, "context cancelled during retry", ctx.Err())
				case <-timer.C:
					continue
				}
			}
		} else {
			return nil
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

// shouldRetry 判断是否应该重试
func (r *retryStrategy) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	// 检查是否是可重试的错误
	return isRetryableError(err)
}

// onRetry 重试回调
func (r *retryStrategy) onRetry(err error) {
	if err == nil {
		return
	}
	// 记录重试日志
	r.logger.WithError(err).Warn(r.ctx, "redis operation retry")
}

// afterRetry 重试后回调
func (r *retryStrategy) afterRetry(err error) {
	if err == nil {
		return
	}
	// 记录重试失败日志
	r.logger.WithError(err).Error(r.ctx, "redis operation retry failed")
}
