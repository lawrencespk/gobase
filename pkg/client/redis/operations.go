package redis

import (
	"context"
	"gobase/pkg/errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// Get 获取键值
func (c *client) Get(ctx context.Context, key string) (string, error) {
	// 添加参数验证
	if key == "" {
		return "", errors.NewRedisCommandError("key is required", nil)
	}

	result, err := c.withOperationResult(ctx, "Get", func() (interface{}, error) {
		return c.client.Get(ctx, key).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return "", errors.NewRedisKeyNotFoundError("key not found", err)
		}
		if isReadOnlyError(err) {
			return "", errors.NewRedisReadOnlyError("failed to get key: readonly", err)
		}
		return "", errors.NewRedisCommandError("failed to get key", err)
	}
	return result.(string), nil
}

// Set 设置键值
func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 添加参数验证
	if key == "" {
		return errors.NewRedisCommandError("key is required", nil)
	}

	_, err := c.withOperationResult(ctx, "Set", func() (interface{}, error) {
		return nil, c.client.Set(ctx, key, value, expiration).Err()
	})
	if err != nil {
		if errors.HasErrorCode(err, "") {
			return err
		}
		if isReadOnlyError(err) {
			return errors.NewRedisReadOnlyError("failed to set key: readonly", err)
		}
		if isMaxMemoryError(err) {
			return errors.NewRedisMaxMemoryError("failed to set key: memory limit exceeded", err)
		}
		return errors.NewRedisCommandError("failed to set key", err)
	}
	return nil
}

// Del 删除键
func (c *client) Del(ctx context.Context, keys ...string) (int64, error) {
	// 添加参数验证
	if len(keys) == 0 {
		return 0, errors.NewRedisCommandError("keys are required", nil)
	}

	result, err := c.withOperationResult(ctx, "Del", func() (interface{}, error) {
		return c.client.Del(ctx, keys...).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to delete keys: readonly", err)
		}
		return 0, errors.NewRedisCommandError("failed to delete keys", err)
	}
	return result.(int64), nil
}

// Incr 实现 Incr 方法
func (c *client) Incr(ctx context.Context, key string) (int64, error) {
	result, err := c.withOperationResult(ctx, "Incr", func() (interface{}, error) {
		return c.client.Incr(ctx, key).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to increment key: readonly", err)
		}
		if isMaxMemoryError(err) {
			return 0, errors.NewRedisMaxMemoryError("failed to increment key: memory limit exceeded", err)
		}
		return 0, errors.NewRedisCommandError("failed to increment key", err)
	}
	return result.(int64), nil
}

// IncrBy 实现 IncrBy 方法
func (c *client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := c.withOperationResult(ctx, "IncrBy", func() (interface{}, error) {
		return c.client.IncrBy(ctx, key, value).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return 0, errors.NewRedisReadOnlyError("failed to increment key by value: readonly", err)
		}
		if isMaxMemoryError(err) {
			return 0, errors.NewRedisMaxMemoryError("failed to increment key by value: memory limit exceeded", err)
		}
		return 0, errors.NewRedisCommandError("failed to increment key by value", err)
	}
	return result.(int64), nil
}

// Close 关闭客户端
func (c *client) Close() error {
	err := c.withOperation(context.Background(), "Close", func() error {
		return c.client.Close()
	})
	if err != nil {
		return errors.NewRedisConnError("failed to close redis connection", err)
	}
	return nil
}

// Ping 心跳检测
func (c *client) Ping(ctx context.Context) error {
	err := c.withOperation(ctx, "Ping", func() error {
		return c.client.Ping(ctx).Err()
	})
	if err != nil {
		return errors.NewRedisConnError("failed to ping redis server", err)
	}
	return nil
}

// Eval 执行Lua脚本
func (c *client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	// 验证脚本不能为空
	if script == "" {
		return nil, errors.NewRedisCommandError("script is required", nil)
	}

	result, err := c.withOperationResult(ctx, "Eval", func() (interface{}, error) {
		return c.client.Eval(ctx, script, keys, args...).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return nil, errors.NewRedisReadOnlyError("failed to execute script: readonly", err)
		}
		return nil, errors.NewRedisScriptError("failed to execute script", err)
	}
	return result, nil
}

// Exists 检查键是否存在
func (c *client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.withOperationResult(ctx, "Exists", func() (interface{}, error) {
		return c.client.Exists(ctx, key).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return false, errors.NewRedisReadOnlyError("failed to check key existence: readonly", err)
		}
		return false, errors.NewRedisCommandError("failed to check key existence", err)
	}
	return result.(int64) > 0, nil
}

// Publish 发布消息到指定的频道
func (c *client) Publish(ctx context.Context, channel string, message interface{}) error {
	// 参数验证
	if channel == "" {
		return errors.NewRedisCommandError("channel is required", nil)
	}
	if message == nil {
		return errors.NewRedisCommandError("message is required", nil)
	}

	_, err := c.withOperationResult(ctx, "Publish", func() (interface{}, error) {
		return c.client.Publish(ctx, channel, message).Result()
	})
	if err != nil {
		if isReadOnlyError(err) {
			return errors.NewRedisReadOnlyError("redis instance is read-only", err)
		}
		if isClusterDownError(err) {
			return errors.NewRedisClusterError("cluster is down", err)
		}
		return errors.NewRedisCommandError("failed to publish message", err)
	}
	return nil
}

// Subscribe 订阅频道
func (c *client) Subscribe(ctx context.Context, channels ...string) PubSub {
	// 参数验证
	if len(channels) == 0 {
		return &pubSub{err: errors.NewRedisCommandError("channels are required", nil)}
	}

	// 创建订阅
	ps := c.client.Subscribe(ctx, channels...)
	return &pubSub{
		ps:     ps,
		client: c,
	}
}

// pubSub 实现 PubSub 接口
type pubSub struct {
	ps     *redis.PubSub
	client *client
	err    error
}

// ReceiveMessage 接收消息
func (p *pubSub) ReceiveMessage(ctx context.Context) (*Message, error) {
	if p.err != nil {
		return nil, p.err
	}

	msg, err := p.ps.ReceiveMessage(ctx)
	if err != nil {
		if isReadOnlyError(err) {
			return nil, errors.NewRedisReadOnlyError("failed to receive message: readonly", err)
		}
		return nil, errors.NewRedisCommandError("failed to receive message", err)
	}

	return &Message{
		Channel: msg.Channel,
		Pattern: msg.Pattern,
		Payload: msg.Payload,
	}, nil
}

// Close 关闭订阅
func (p *pubSub) Close() error {
	if p.err != nil {
		return p.err
	}

	err := p.ps.Close()
	if err != nil {
		return errors.NewRedisConnError("failed to close pubsub connection", err)
	}
	return nil
}
