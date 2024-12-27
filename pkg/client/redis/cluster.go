package redis

import (
	"context"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

// clusterClient Redis集群客户端实现
type clusterClient struct {
	client  *redis.ClusterClient
	logger  types.Logger
	tracer  opentracing.Tracer
	options *Options
}

// NewClusterClient 创建一个新的Redis集群客户端
func NewClusterClient(opts ...Option) (Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 基本验证
	if len(options.Addresses) == 0 {
		return nil, errors.NewError(codes.CacheError, "redis addresses are required", nil)
	}

	// 创建集群配置
	clusterOpts := &redis.ClusterOptions{
		Addrs:         options.Addresses,
		Username:      options.Username,
		Password:      options.Password,
		MaxRetries:    options.MaxRetries,
		PoolSize:      options.PoolSize,
		MinIdleConns:  options.MinIdleConns,
		IdleTimeout:   options.IdleTimeout,
		DialTimeout:   options.DialTimeout,
		ReadTimeout:   options.ReadTimeout,
		WriteTimeout:  options.WriteTimeout,
		RouteRandomly: options.RouteRandomly,
	}

	// 配置TLS
	if options.EnableTLS {
		tlsConfig, err := loadTLSConfig(options)
		if err != nil {
			return nil, err
		}
		clusterOpts.TLSConfig = tlsConfig
	}

	// 创建集群客户端
	rdb := redis.NewClusterClient(clusterOpts)

	// 验证连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		options.Logger.WithError(err).Error(ctx, "failed to connect to redis cluster")
		return nil, errors.NewError(codes.CacheError, "failed to connect to redis cluster", err)
	}

	return &clusterClient{
		client:  rdb,
		logger:  options.Logger,
		tracer:  options.Tracer,
		options: options,
	}, nil
}

// Close 关闭集群客户端
func (c *clusterClient) Close() error {
	if err := c.client.Close(); err != nil {
		c.logger.WithError(err).Error(context.Background(), "redis cluster close failed")
		return errors.NewError(codes.CacheError, "redis cluster close failed", err)
	}
	return nil
}

// Del 删除键
func (c *clusterClient) Del(ctx context.Context, keys ...string) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Del")
	defer span.Finish()

	result, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to delete keys", err)
	}
	return result, nil
}

// Eval 执行Lua脚本
func (c *clusterClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Eval")
	defer span.Finish()

	result, err := c.client.Eval(ctx, script, keys, args...).Result()
	if err != nil {
		return nil, errors.NewError(codes.CacheError, "failed to execute lua script", err)
	}
	return result, nil
}

// Get 获取键值
func (c *clusterClient) Get(ctx context.Context, key string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.Get")
	defer span.Finish()

	result, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.NewError(codes.CacheError, "key not found", err)
	}
	if err != nil {
		return "", errors.NewError(codes.CacheError, "failed to get key", err)
	}
	return result, nil
}

// Set 设置键值
func (c *clusterClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	span, ctx := startSpan(ctx, c.tracer, "redis.Set")
	defer span.Finish()

	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return errors.NewError(codes.CacheError, "failed to set key", err)
	}
	return nil
}

// HGet 获取哈希字段值
func (c *clusterClient) HGet(ctx context.Context, key, field string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HGet")
	defer span.Finish()

	result, err := c.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", errors.NewError(codes.CacheError, "field not found", err)
	}
	if err != nil {
		return "", errors.NewError(codes.CacheError, "failed to get hash field", err)
	}
	return result, nil
}

// HSet 设置哈希字段值
func (c *clusterClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HSet")
	defer span.Finish()

	result, err := c.client.HSet(ctx, key, values...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to set hash field", err)
	}
	return result, nil
}

// HDel 删除哈希字段
func (c *clusterClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.HDel")
	defer span.Finish()

	result, err := c.client.HDel(ctx, key, fields...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to delete hash fields", err)
	}
	return result, nil
}

// LPush 从列表左端推入元素
func (c *clusterClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.LPush")
	defer span.Finish()

	result, err := c.client.LPush(ctx, key, values...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to push to list", err)
	}
	return result, nil
}

// LPop 从列表左端弹出元素
func (c *clusterClient) LPop(ctx context.Context, key string) (string, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.LPop")
	defer span.Finish()

	result, err := c.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.NewError(codes.CacheError, "list is empty", err)
	}
	if err != nil {
		return "", errors.NewError(codes.CacheError, "failed to pop from list", err)
	}
	return result, nil
}

// SAdd 向集合添加元素
func (c *clusterClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.SAdd")
	defer span.Finish()

	result, err := c.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to add to set", err)
	}
	return result, nil
}

// SRem 从集合中移除元素
func (c *clusterClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.SRem")
	defer span.Finish()

	result, err := c.client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to remove from set", err)
	}
	return result, nil
}

// TxPipeline 创建一个事务管道
func (c *clusterClient) TxPipeline() Pipeline {
	return &redisPipeline{
		pipeline: c.client.Pipeline(),
		tracer:   c.tracer,
	}
}

// Ping 检查连接
func (c *clusterClient) Ping(ctx context.Context) error {
	span, ctx := startSpan(ctx, c.tracer, "redis.Ping")
	defer span.Finish()

	err := c.client.Ping(ctx).Err()
	if err != nil {
		return errors.NewError(codes.CacheError, "failed to ping redis cluster", err)
	}
	return nil
}

// ZAdd 添加元素到有序集合
func (c *clusterClient) ZAdd(ctx context.Context, key string, members ...*Z) (int64, error) {
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
func (c *clusterClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis.ZRem")
	defer span.Finish()

	result, err := c.client.ZRem(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewError(codes.CacheError, "failed to remove from sorted set", err)
	}
	return result, nil
}

// 实现所有接口方法...
// 注意：集群客户端的实现与单机客户端类似，只是底层使用 ClusterClient
