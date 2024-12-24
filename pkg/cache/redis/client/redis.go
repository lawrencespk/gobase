package client

import (
	"context"
	"fmt"
	"time"

	"gobase/pkg/cache/redis/config/types"

	"github.com/go-redis/redis/v8"
)

// redisClient Redis客户端实现
type redisClient struct {
	client *redis.Client
}

// NewClient 创建新的Redis客户端
func NewClient(cfg *types.Config) (Client, error) {
	opt := &redis.Options{
		Addr:         cfg.Addresses[0], // 暂时只使用第一个地址
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	client := redis.NewClient(opt)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisClient{client: client}, nil
}

// Get 实现 Get 方法
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set 实现 Set 方法
func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del 实现 Del 方法
func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Incr 实现 Incr 方法
func (r *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy 实现 IncrBy 方法
func (r *redisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// HGet 实现 HGet 方法
func (r *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HSet 实现 HSet 方法
func (r *redisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// LPush 实现 LPush 方法
func (r *redisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// LPop 实现 LPop 方法
func (r *redisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// SAdd 实现 SAdd 方法
func (r *redisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem 实现 SRem 方法
func (r *redisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// ZAdd 实现 ZAdd 方法
func (r *redisClient) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	// 将 interface{} 转换为 *redis.Z
	zMembers := make([]*redis.Z, 0, len(members)/2)
	for i := 0; i < len(members); i += 2 {
		score, ok := members[i].(float64)
		if !ok {
			return fmt.Errorf("score must be float64")
		}

		member := members[i+1]
		zMembers = append(zMembers, &redis.Z{
			Score:  score,
			Member: member,
		})
	}

	return r.client.ZAdd(ctx, key, zMembers...).Err()
}

// ZRem 实现 ZRem 方法
func (r *redisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Eval 实现 Eval 方法
func (r *redisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

// TxPipeline 实现 TxPipeline 方法
func (r *redisClient) TxPipeline() Pipeline {
	return &redisPipeline{pipeline: r.client.Pipeline()}
}

// Close 实现 Close 方法
func (r *redisClient) Close() error {
	return r.client.Close()
}

// redisPipeline Redis管道实现
type redisPipeline struct {
	pipeline redis.Pipeliner
}

// Exec 实现 Pipeline.Exec 方法
func (p *redisPipeline) Exec(ctx context.Context) error {
	_, err := p.pipeline.Exec(ctx)
	return err
}

// Discard 实现 Pipeline.Discard 方法
func (p *redisPipeline) Discard() error {
	return p.pipeline.Discard()
}
