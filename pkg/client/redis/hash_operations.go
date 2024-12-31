package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// HGet 获取哈希键值
func (c *client) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := c.withOperationResult(ctx, "HGet", func() (interface{}, error) {
		return c.client.HGet(ctx, key, field).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return "", handleRedisError(err, errFieldNotFound)
		}
		return "", err
	}
	return result.(string), nil
}

// HSet 设置哈希键值
func (c *client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := c.withOperationResult(ctx, "HSet", func() (interface{}, error) {
		return c.client.HSet(ctx, key, values...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// HDel 删除哈希表中的字段
func (c *client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	result, err := c.withOperationResult(ctx, "HDel", func() (interface{}, error) {
		return c.client.HDel(ctx, key, fields...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}
