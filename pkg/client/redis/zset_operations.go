package redis

import (
	"context"
	"gobase/pkg/errors"

	"github.com/go-redis/redis/v8"
)

// ZAdd 添加元素到有序集合
func (c *client) ZAdd(ctx context.Context, key string, members ...*Z) (int64, error) {
	// 参数验证
	if key == "" {
		return 0, errors.NewRedisCommandError("key is required", nil)
	}
	if len(members) == 0 {
		return 0, errors.NewRedisCommandError("no members to add", nil)
	}

	// 将我们的 Z 类型转换为 redis.Z 类型
	zMembers := make([]*redis.Z, len(members))
	for i, member := range members {
		if member == nil {
			return 0, errors.NewRedisCommandError("member cannot be nil", nil)
		}
		zMembers[i] = &redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
	}

	var result int64
	err := c.withOperation(ctx, "ZAdd", func() error {
		var err error
		result, err = c.client.ZAdd(ctx, key, zMembers...).Result()
		if err != nil {
			if err == redis.Nil {
				return errors.NewRedisKeyNotFoundError("key not found", err)
			}
			if isReadOnlyError(err) {
				return errors.NewRedisReadOnlyError("redis instance is read-only", err)
			}
			if isClusterDownError(err) {
				return errors.NewRedisClusterError("cluster is down", err)
			}
			if isLoadingError(err) {
				return errors.NewRedisLoadingError("redis is loading the dataset", err)
			}
			return errors.NewRedisCommandError("failed to add members to sorted set", err)
		}
		return nil
	})
	return result, err
}

// ZRem 从有序集合中移除元素
func (c *client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// 参数验证
	if key == "" {
		return 0, errors.NewRedisCommandError("key is required", nil)
	}
	if len(members) == 0 {
		return 0, errors.NewRedisCommandError("no members to remove", nil)
	}

	// 验证成员
	for _, member := range members {
		if member == nil {
			return 0, errors.NewRedisCommandError("member cannot be nil", nil)
		}
	}

	var result int64
	err := c.withOperation(ctx, "ZRem", func() error {
		var err error
		result, err = c.client.ZRem(ctx, key, members...).Result()
		if err != nil {
			if err == redis.Nil {
				return errors.NewRedisKeyNotFoundError("key not found", err)
			}
			if isReadOnlyError(err) {
				return errors.NewRedisReadOnlyError("redis instance is read-only", err)
			}
			if isClusterDownError(err) {
				return errors.NewRedisClusterError("cluster is down", err)
			}
			if isLoadingError(err) {
				return errors.NewRedisLoadingError("redis is loading the dataset", err)
			}
			return errors.NewRedisCommandError("failed to remove members from sorted set", err)
		}
		return nil
	})
	return result, err
}
