package redis

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
)

// Cache Redis缓存实现
type Cache struct {
	client redis.Client
	logger types.Logger
}

// Options Redis缓存配置选项
type Options struct {
	Client redis.Client
	Logger types.Logger
}

// NewCache 创建Redis缓存实例
func NewCache(opts Options) (*Cache, error) {
	if opts.Client == nil {
		return nil, errors.NewRedisInvalidConfigError("redis client is required", nil)
	}

	return &Cache{
		client: opts.Client,
		logger: opts.Logger,
	}, nil
}

// Set 设置缓存
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 参数验证
	if key == "" {
		return errors.NewRedisCommandError("key is required", nil)
	}

	// 验证值不能为nil
	if value == nil {
		return errors.NewRedisCommandError("value cannot be nil", nil)
	}

	// 验证过期时间不能为负数
	if ttl < 0 {
		return errors.NewRedisCommandError("ttl cannot be negative", nil)
	}

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return errors.NewRedisCommandError("failed to marshal value", err)
	}

	// 设置缓存
	err = c.client.Set(ctx, key, string(data), ttl)
	if err != nil {
		return errors.NewRedisCommandError("failed to set cache", err)
	}

	return nil
}

// Get 获取缓存
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	// 参数验证
	if key == "" {
		return nil, errors.NewRedisCommandError("key is required", nil)
	}

	// 获取缓存
	data, err := c.client.Get(ctx, key)
	if err != nil {
		if errors.HasErrorCode(err, codes.RedisKeyNotFoundError) {
			return nil, errors.NewRedisKeyNotFoundError("cache not found", err)
		}
		return nil, errors.NewRedisCommandError("failed to get cache", err)
	}

	// 反序列化值
	var value interface{}
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil, errors.NewRedisCommandError("failed to unmarshal value", err)
	}

	return value, nil
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key string) error {
	// 参数验证
	if key == "" {
		return errors.NewRedisCommandError("key is required", nil)
	}

	// 删除缓存
	_, err := c.client.Del(ctx, key)
	if err != nil {
		return errors.NewRedisCommandError("failed to delete cache", err)
	}

	return nil
}

// Clear 清空所有缓存
func (c *Cache) Clear(ctx context.Context) error {
	// 使用FLUSHDB命令清空当前数据库
	_, err := c.client.Eval(ctx, "return redis.call('FLUSHDB')", nil)
	if err != nil {
		return errors.NewRedisCommandError("failed to clear cache", err)
	}

	return nil
}

// GetLevel 获取缓存级别
func (c *Cache) GetLevel() cache.Level {
	return cache.L2Cache
}
