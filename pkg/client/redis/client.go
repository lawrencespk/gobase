package redis

import (
	"context"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

// client Redis客户端实现
type client struct {
	client  *redis.Client
	logger  types.Logger
	tracer  opentracing.Tracer
	options *Options
}

// NewClient 创建一个新的Redis客户端
func NewClient(opts ...Option) (Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 基本验证
	if len(options.Addresses) == 0 {
		return nil, errors.NewError(codes.CacheError, "redis addresses are required", nil)
	}

	// 创建 Redis 配置
	redisOpts := &redis.Options{
		Addr:         options.Addresses[0], // 暂时只使用第一个地址
		Username:     options.Username,
		Password:     options.Password,
		DB:           options.DB,
		MaxRetries:   options.MaxRetries,
		PoolSize:     options.PoolSize,
		MinIdleConns: options.MinIdleConns,
		IdleTimeout:  options.IdleTimeout,
		DialTimeout:  options.DialTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	}

	// 配置TLS
	if options.EnableTLS {
		tlsConfig, err := loadTLSConfig(options)
		if err != nil {
			return nil, err
		}
		redisOpts.TLSConfig = tlsConfig
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(redisOpts)

	// 验证连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		options.Logger.WithError(err).Error(ctx, "failed to connect to redis")
		return nil, errors.NewError(codes.CacheError, "failed to connect to redis", err)
	}

	return &client{
		client:  rdb,
		logger:  options.Logger,
		tracer:  options.Tracer,
		options: options,
	}, nil
}

// ... 实现所有接口方法 ...

// Get 获取键值
func (c *client) Get(ctx context.Context, key string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Get")
	defer span.Finish()

	var result string
	err := withMetrics(ctx, "get", func() error {
		return withRetry(ctx, c.options, func() error {
			var err error
			result, err = c.client.Get(ctx, key).Result()
			if err == redis.Nil {
				return errors.NewError(codes.CacheError, "key not found", err)
			}
			return err
		})
	})

	if err != nil {
		return "", errors.NewError(codes.CacheError, "failed to get key", err)
	}
	return result, nil
}

// Set 设置键值
func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	span, ctx := startSpan(ctx, c.tracer, "redis.Set")
	defer span.Finish()

	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return errors.NewError(codes.CacheError, "failed to set key", err)
	}
	return nil
}

// Del 删除键
func (c *client) Del(ctx context.Context, keys ...string) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Del")
	defer span.Finish()

	result, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to delete keys", err)
	}
	return result, nil
}

// Incr 实现 Incr 方法
func (c *client) Incr(ctx context.Context, key string) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Incr")
	defer span.Finish()

	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to increment key", err)
	}
	return result, nil
}

// IncrBy 实现 IncrBy 方法
func (c *client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.IncrBy")
	defer span.Finish()

	result, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to increment key by value", err)
	}
	return result, nil
}

// HDel 删除哈希表中的字段
func (c *client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HDel")
	defer span.Finish()

	span.SetTag("redis.key", key)
	span.SetTag("redis.fields", fields)

	n, err := c.client.HDel(ctx, key, fields...).Result()
	if err != nil {
		c.logger.WithError(err).Error(ctx, "redis hdel failed")
		return 0, errors.NewError(codes.CacheError, "redis operation failed", err)
	}
	return n, nil
}

// Close 关闭客户端
func (c *client) Close() error {
	if err := c.client.Close(); err != nil {
		c.logger.WithError(err).Error(context.Background(), "redis close failed")
		return errors.NewError(codes.CacheError, "redis close failed", err)
	}
	return nil
}

// Ping 心跳检测
func (c *client) Ping(ctx context.Context) error {
	span, ctx := startSpan(ctx, c.tracer, "redis.Ping")
	defer span.Finish()

	if err := c.client.Ping(ctx).Err(); err != nil {
		c.logger.WithError(err).Error(ctx, "redis ping failed")
		return errors.NewError(codes.CacheError, "redis connection failed", err)
	}
	return nil
}

// HGet 获取哈希键值
func (c *client) HGet(ctx context.Context, key, field string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HGet")
	defer span.Finish()

	result, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.NewError(codes.CacheError, "field not found", err)
		}
		return "", errors.NewError(codes.CacheError, "failed to get hash field", err)
	}
	return result, nil
}

// HSet 设置哈希键值
func (c *client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HSet")
	defer span.Finish()

	result, err := c.client.HSet(ctx, key, values...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to set hash fields", err)
	}
	return result, nil
}

// Eval 执行Lua脚本
func (c *client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Eval")
	defer span.Finish()

	result, err := c.client.Eval(ctx, script, keys, args...).Result()
	if err != nil {
		return nil, errors.NewError(codes.CacheError, "failed to execute lua script", err)
	}
	return result, nil
}

// LPop 从列表左端弹出元素
func (c *client) LPop(ctx context.Context, key string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.LPop")
	defer span.Finish()

	result, err := c.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.NewError(codes.CacheError, "list is empty", err)
		}
		return "", errors.NewError(codes.CacheError, "failed to pop from list", err)
	}
	return result, nil
}

// LPush 从列表左端推入元素
func (c *client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.LPush")
	defer span.Finish()

	result, err := c.client.LPush(ctx, key, values...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to push to list", err)
	}
	return result, nil
}

// SAdd 向集合添加元素
func (c *client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.SAdd")
	defer span.Finish()

	result, err := c.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to add to set", err)
	}
	return result, nil
}

// SRem 从集合中移除元素
func (c *client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.SRem")
	defer span.Finish()

	result, err := c.client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to remove from set", err)
	}
	return result, nil
}

// TxPipeline 创建一个事务管道
func (c *client) TxPipeline() Pipeline {
	return &redisPipeline{
		pipeline: c.client.Pipeline(),
		tracer:   c.tracer,
	}
}

// redisPipeline Redis管道实现
type redisPipeline struct {
	pipeline redis.Pipeliner
	tracer   opentracing.Tracer
}

// Exec 实现 Pipeline.Exec 方法
func (p *redisPipeline) Exec(ctx context.Context) ([]Cmder, error) {
	span, ctx := startSpan(ctx, p.tracer, "redis.Pipeline.Exec")
	defer span.Finish()

	cmds, err := p.pipeline.Exec(ctx)
	if err != nil {
		return nil, errors.NewError(codes.CacheError, "failed to execute pipeline", err)
	}

	// 转换命令结果
	results := make([]Cmder, len(cmds))
	for i, cmd := range cmds {
		results[i] = cmd
	}
	return results, nil
}

// Close 实现 Pipeline.Close 方法
func (p *redisPipeline) Close() error {
	return p.pipeline.Discard()
}

// ZAdd 添加元素到有序集合
func (c *client) ZAdd(ctx context.Context, key string, members ...*Z) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.ZAdd")
	defer span.Finish()

	// 将我们的 Z 类型转换为 redis.Z 类型
	zMembers := make([]*redis.Z, len(members))
	for i, member := range members {
		zMembers[i] = &redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
	}

	result, err := c.client.ZAdd(ctx, key, zMembers...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to add to sorted set", err)
	}
	return result, nil
}

// ZRem 从有序集合中移除元素
func (c *client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.ZRem")
	defer span.Finish()

	result, err := c.client.ZRem(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to remove from sorted set", err)
	}
	return result, nil
}

// ... 实现其他接口方法 ...
