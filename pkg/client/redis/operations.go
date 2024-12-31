package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Get 获取键值
func (c *client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.withOperationResult(ctx, "Get", func() (interface{}, error) {
		return c.client.Get(ctx, key).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return "", handleRedisError(err, errKeyNotFound)
		}
		return "", err
	}
	return result.(string), nil
}

// Set 设置键值
func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.withOperation(ctx, "Set", func() error {
		return c.client.Set(ctx, key, value, expiration).Err()
	})
}

// Del 删除键
func (c *client) Del(ctx context.Context, keys ...string) (int64, error) {
	result, err := c.withOperationResult(ctx, "Del", func() (interface{}, error) {
		return c.client.Del(ctx, keys...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Incr 实现 Incr 方法
func (c *client) Incr(ctx context.Context, key string) (int64, error) {
	result, err := c.withOperationResult(ctx, "Incr", func() (interface{}, error) {
		return c.client.Incr(ctx, key).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// IncrBy 实现 IncrBy 方法
func (c *client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := c.withOperationResult(ctx, "IncrBy", func() (interface{}, error) {
		return c.client.IncrBy(ctx, key, value).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Close 关闭客户端
func (c *client) Close() error {
	return c.withOperation(context.Background(), "Close", func() error {
		return c.client.Close()
	})
}

// Ping 心跳检测
func (c *client) Ping(ctx context.Context) error {
	return c.withOperation(ctx, "Ping", func() error {
		return c.client.Ping(ctx).Err()
	})
}

// Eval 执行Lua脚本
func (c *client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.withOperationResult(ctx, "Eval", func() (interface{}, error) {
		return c.client.Eval(ctx, script, keys, args...).Result()
	})
}

// Exists 检查键是否存在
func (c *client) Exists(ctx context.Context, key string) (bool, error) {
	var result int64
	err := c.withOperation(ctx, "Exists", func() error {
		var err error
		result, err = c.client.Exists(ctx, key).Result()
		return err
	})
	return result > 0, err
}
