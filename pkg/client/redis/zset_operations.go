package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// ZAdd 添加元素到有序集合
func (c *client) ZAdd(ctx context.Context, key string, members ...*Z) (int64, error) {
	// 将我们的 Z 类型转换为 redis.Z 类型
	zMembers := make([]*redis.Z, len(members))
	for i, member := range members {
		zMembers[i] = &redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
	}

	var result int64
	err := c.withOperation(ctx, "ZAdd", func() error {
		var err error
		result, err = c.client.ZAdd(ctx, key, zMembers...).Result()
		return err
	})
	return result, err
}

// ZRem 从有序集合中移除元素
func (c *client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	var result int64
	err := c.withOperation(ctx, "ZRem", func() error {
		var err error
		result, err = c.client.ZRem(ctx, key, members...).Result()
		return err
	})
	return result, err
}
