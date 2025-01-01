package redis

import (
	"context"
	"gobase/pkg/errors"

	"github.com/go-redis/redis/v8"
)

// SAdd 向集合添加元素
func (c *client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// 参数验证
	if key == "" {
		return 0, errors.NewRedisCommandError("key is required", nil)
	}
	if len(members) == 0 {
		return 0, errors.NewRedisCommandError("no members to add", nil)
	}

	result, err := c.withOperationResult(ctx, "SAdd", func() (interface{}, error) {
		res, err := c.client.SAdd(ctx, key, members...).Result()
		if err != nil {
			if err == redis.Nil {
				return 0, errors.NewRedisKeyNotFoundError("key not found", err)
			}
			// 检查是否是只读错误（在集群模式下可能发生）
			if isReadOnlyError(err) {
				return 0, errors.NewRedisReadOnlyError("redis instance is read-only", err)
			}
			// 检查是否是集群错误
			if isClusterDownError(err) {
				return 0, errors.NewRedisClusterError("cluster is down", err)
			}
			return 0, errors.NewRedisCommandError("failed to add members to set", err)
		}
		return res, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// SRem 从集合中移除元素
func (c *client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// 参数验证
	if key == "" {
		return 0, errors.NewRedisCommandError("key is required", nil)
	}
	if len(members) == 0 {
		return 0, errors.NewRedisCommandError("no members to remove", nil)
	}

	result, err := c.withOperationResult(ctx, "SRem", func() (interface{}, error) {
		res, err := c.client.SRem(ctx, key, members...).Result()
		if err != nil {
			if err == redis.Nil {
				return 0, errors.NewRedisKeyNotFoundError("key not found", err)
			}
			// 检查是否是只读错误（在集群模式下可能发生）
			if isReadOnlyError(err) {
				return 0, errors.NewRedisReadOnlyError("redis instance is read-only", err)
			}
			// 检查是否是集群错误
			if isClusterDownError(err) {
				return 0, errors.NewRedisClusterError("cluster is down", err)
			}
			return 0, errors.NewRedisCommandError("failed to remove members from set", err)
		}
		return res, nil
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}
