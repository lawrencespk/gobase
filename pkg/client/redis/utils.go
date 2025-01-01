package redis

import (
	"context"
	"fmt"
	"gobase/pkg/logger/types"
	"strings"

	"gobase/pkg/errors"

	"github.com/go-redis/redis/v8"
)

// withOperation 统一的操作包装函数
func (c *client) withOperation(ctx context.Context, operation string, fn func() error) error {
	span, ctx := startSpan(ctx, c.tracer, "redis."+operation)
	defer span.Finish()

	// 将客户端实例添加到上下文中
	ctx = context.WithValue(ctx, clientKey, c)

	c.logger.WithFields(
		types.Field{Key: "operation", Value: operation},
		types.Field{Key: "client_id", Value: fmt.Sprintf("%p", c)},
	).Debug(ctx, "starting redis operation")

	err := withMetrics(ctx, operation, c.options.Registry, fn)
	if err != nil {
		// 1. 如果是上下文超时，返回超时错误
		if err == context.DeadlineExceeded ||
			strings.Contains(err.Error(), "context deadline exceeded") {
			return errors.NewTimeoutError("operation timed out", err)
		}

		// 2. 如果已经是包装过的错误，直接返回
		if errors.HasErrorCode(err, "") {
			return err
		}

		// 3. 处理特定的 Redis 错误
		if strings.Contains(err.Error(), "NOAUTH") || strings.Contains(err.Error(), "AUTH") {
			return errors.NewRedisAuthError("authentication failed", err)
		}

		// 处理连接池错误
		if strings.Contains(err.Error(), "pool exhausted") {
			return errors.NewRedisPoolExhaustedError("connection pool exhausted", err)
		}

		// 处理连接错误
		if strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "network is down") {
			return errors.NewRedisConnError("connection failed", err)
		}

		// 处理只读错误
		if isReadOnlyError(err) {
			return errors.NewRedisReadOnlyError("redis instance is read-only", err)
		}

		// 处理集群错误
		if isClusterDownError(err) {
			return errors.NewRedisClusterError("cluster is down", err)
		}

		// 处理加载错误
		if isLoadingError(err) {
			return errors.NewRedisLoadingError("redis is loading the dataset in memory", err)
		}
	}
	return err
}

// withOperationResult 用于处理有返回值的Redis操作
func (c *client) withOperationResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := c.withOperation(ctx, operation, func() error {
		var err error
		result, err = fn()
		if err != nil {
			// 处理键不存在的情况
			if err == redis.Nil {
				return errors.NewRedisKeyNotFoundError("key not found", err)
			}
			return err
		}
		return nil
	})
	return result, err
}
