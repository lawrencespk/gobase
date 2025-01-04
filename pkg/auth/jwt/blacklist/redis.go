package blacklist

import (
	"context"
	"fmt"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/monitor/prometheus/metric"
)

// RedisStore Redis存储实现
type RedisStore struct {
	client redis.Client
	opts   *Options

	// 监控指标
	tokenCount *metric.Gauge   // 当前黑名单中的token数量
	addTotal   *metric.Counter // 添加token的总次数
	hitTotal   *metric.Counter // 命中黑名单的总次数
	missTotal  *metric.Counter // 未命中黑名单的总次数
}

// NewRedisStore 创建Redis存储实例
func NewRedisStore(client redis.Client, opts *Options) (Store, error) {
	if client == nil {
		return nil, errors.NewError(codes.InvalidParams, "redis client cannot be nil", nil)
	}
	if opts == nil {
		opts = DefaultOptions()
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	store := &RedisStore{
		client: client,
		opts:   opts,
	}

	// 初始化监控指标
	if opts.EnableMetrics {
		store.tokenCount = metric.NewGauge(metric.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "redis_tokens_total",
			Help:      "Total number of tokens in Redis blacklist",
		})

		store.addTotal = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "redis_add_total",
			Help:      "Total number of tokens added to Redis blacklist",
		})

		store.hitTotal = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "redis_hit_total",
			Help:      "Total number of Redis blacklist hits",
		})

		store.missTotal = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "redis_miss_total",
			Help:      "Total number of Redis blacklist misses",
		})
	}

	return store, nil
}

// Add 添加到黑名单
func (s *RedisStore) Add(ctx context.Context, tokenID, reason string, expiration time.Duration) error {
	if tokenID == "" {
		return errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}
	if expiration <= 0 {
		return errors.NewError(codes.InvalidParams, "expiration must be positive", nil)
	}

	key := fmt.Sprintf("%s%s", s.opts.KeyPrefix, tokenID)
	if err := s.client.Set(ctx, key, reason, expiration); err != nil {
		return errors.NewError(codes.StoreErrAdd, "failed to add token to blacklist", err)
	}

	if s.opts.EnableMetrics {
		s.tokenCount.Inc()
		s.addTotal.Inc()
	}

	return nil
}

// Get 获取黑名单原因
func (s *RedisStore) Get(ctx context.Context, tokenID string) (string, error) {
	if tokenID == "" {
		return "", errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}

	key := fmt.Sprintf("%s%s", s.opts.KeyPrefix, tokenID)
	reason, err := s.client.Get(ctx, key)
	if err != nil {
		// 使用 errors.Is 和 NotFound 错误码来检查
		notFoundErr := errors.NewError(codes.NotFound, "", nil)
		if errors.Is(err, redis.ErrNil) || errors.Is(err, notFoundErr) {
			if s.opts.EnableMetrics {
				s.missTotal.Inc()
			}
			return "", errors.NewError(codes.StoreErrNotFound, "token not found in blacklist", nil)
		}
		return "", errors.NewError(codes.StoreErrGet, "failed to get token from blacklist", err)
	}

	if s.opts.EnableMetrics {
		s.hitTotal.Inc()
	}

	return reason, nil
}

// Remove 从黑名单中移除
func (s *RedisStore) Remove(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}

	key := fmt.Sprintf("%s%s", s.opts.KeyPrefix, tokenID)
	n, err := s.client.Del(ctx, key)
	if err != nil {
		return errors.NewError(codes.StoreErrDelete, "failed to remove token from blacklist", err)
	}

	if n > 0 && s.opts.EnableMetrics {
		s.tokenCount.Dec()
	}

	return nil
}

// Close 关闭存储
func (s *RedisStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
