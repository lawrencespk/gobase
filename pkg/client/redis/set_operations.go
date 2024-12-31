package redis

import "context"

// SAdd 向集合添加元素
func (c *client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.withOperationResult(ctx, "SAdd", func() (interface{}, error) {
		return c.client.SAdd(ctx, key, members...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// SRem 从集合中移除元素
func (c *client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := c.withOperationResult(ctx, "SRem", func() (interface{}, error) {
		return c.client.SRem(ctx, key, members...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}
