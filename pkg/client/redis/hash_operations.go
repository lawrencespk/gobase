package redis

import (
	"context"
	"gobase/pkg/errors"

	"github.com/go-redis/redis/v8"
)

// HGet 获取哈希键值
func (c *client) HGet(ctx context.Context, key, field string) (string, error) {
	// 参数验证
	if key == "" {
		return "", errors.NewRedisCommandError("key is required", nil)
	}
	if field == "" {
		return "", errors.NewRedisCommandError("field is required", nil)
	}

	result, err := c.withOperationResult(ctx, "HGet", func() (interface{}, error) {
		return c.client.HGet(ctx, key, field).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return "", errors.NewRedisKeyNotFoundError(errFieldNotFound, err)
		}
		return "", errors.NewRedisCommandError("failed to get hash field", err)
	}
	return result.(string), nil
}

// HSet 设置哈希键值
func (c *client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := c.withOperationResult(ctx, "HSet", func() (interface{}, error) {
		return c.client.HSet(ctx, key, values...).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to set hash field: readonly", err)
		}
		return 0, errors.NewRedisCommandError("failed to set hash field", err)
	}
	return result.(int64), nil
}

// HDel 删除哈希表中的字段
func (c *client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	result, err := c.withOperationResult(ctx, "HDel", func() (interface{}, error) {
		return c.client.HDel(ctx, key, fields...).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to delete hash field: readonly", err)
		}
		return 0, errors.NewRedisCommandError("failed to delete hash field", err)
	}
	return result.(int64), nil
}
