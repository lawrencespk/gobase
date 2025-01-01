package redis

import (
	"context"
	"gobase/pkg/errors"
	"strings"

	"github.com/go-redis/redis/v8"
)

// LPop 从列表左端弹出元素
func (c *client) LPop(ctx context.Context, key string) (string, error) {
	result, err := c.withOperationResult(ctx, "LPop", func() (interface{}, error) {
		return c.client.LPop(ctx, key).Result()
	})

	if err != nil {
		if err == redis.Nil {
			return "", errors.NewRedisKeyNotFoundError("list is empty", err)
		}
		if isReadOnlyError(err) {
			return "", errors.NewRedisReadOnlyError("failed to pop from list: readonly", err)
		}
		return "", errors.NewRedisCommandError("failed to pop from list", err)
	}

	return result.(string), nil
}

// LPush 从列表左端推入元素
func (c *client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := c.withOperationResult(ctx, "LPush", func() (interface{}, error) {
		return c.client.LPush(ctx, key, values...).Result()
	})

	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to push to list: readonly", err)
		}
		// 检查内存限制错误
		if isMaxMemoryError(err) {
			return 0, errors.NewRedisMaxMemoryError("failed to push to list: memory limit exceeded", err)
		}
		return 0, errors.NewRedisCommandError("failed to push to list", err)
	}

	return result.(int64), nil
}

// isMaxMemoryError 检查是否是内存限制错误
func isMaxMemoryError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "OOM") ||
		strings.Contains(errMsg, "out of memory") ||
		strings.Contains(errMsg, "maxmemory")
}
