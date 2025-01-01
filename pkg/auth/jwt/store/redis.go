package store

import (
	"context"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// RedisStore Redis存储实现
type RedisStore struct {
	client  redis.Client
	prefix  string
	logger  types.Logger
	metrics *metric.Counter
}

// NewRedisStore 创建Redis存储实例
func NewRedisStore(opts Options, logger types.Logger) (*RedisStore, error) {
	if opts.Redis == nil {
		return nil, errors.NewError(codes.ConfigErrRequired, "redis options required", nil)
	}

	// 初始化Redis客户端
	redisClient, err := redis.NewClient(redis.WithAddress(opts.Redis.Addr),
		redis.WithPassword(opts.Redis.Password),
		redis.WithDB(opts.Redis.DB))
	if err != nil {
		return nil, errors.NewError(codes.StoreErrCreate, "failed to create redis client", err)
	}

	store := &RedisStore{
		client: redisClient,
		prefix: opts.KeyPrefix,
		logger: logger,
	}

	// 初始化监控指标
	if opts.EnableMetrics {
		store.metrics = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_store",
			Name:      "operations_total",
			Help:      "Total number of JWT store operations",
		})
	}

	return store, nil
}

// Set 存储JWT令牌
func (s *RedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.set")
	if span != nil {
		defer span.Finish()
	}

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	key = s.prefix + key
	err := s.client.Set(ctx, key, value, expiration)
	if err != nil {
		s.logger.WithFields(
			types.Field{Key: "key", Value: key},
			types.Field{Key: "error", Value: err},
		).Error(ctx, "failed to set jwt token")
		return errors.NewError(codes.StoreErrSet, "failed to set jwt token", err)
	}

	s.logger.WithFields(
		types.Field{Key: "key", Value: key},
		types.Field{Key: "expiration", Value: expiration},
	).Debug(ctx, "jwt token stored")

	return nil
}

// Get 获取JWT令牌
func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.get")
	if span != nil {
		defer span.Finish()
	}

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	key = s.prefix + key
	value, err := s.client.Get(ctx, key)
	if err == redis.ErrNil {
		return "", errors.NewError(codes.StoreErrNotFound, "token not found", err)
	}
	if err != nil {
		s.logger.WithFields(
			types.Field{Key: "key", Value: key},
			types.Field{Key: "error", Value: err},
		).Error(ctx, "failed to get jwt token")
		return "", errors.NewError(codes.StoreErrGet, "failed to get jwt token", err)
	}

	return value, nil
}

// Delete 删除JWT令牌
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.delete")
	if span != nil {
		defer span.Finish()
	}

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	key = s.prefix + key
	_, err := s.client.Del(ctx, key)
	if err != nil {
		s.logger.WithFields(
			types.Field{Key: "key", Value: key},
			types.Field{Key: "error", Value: err},
		).Error(ctx, "failed to delete jwt token")
		return errors.NewError(codes.StoreErrDelete, "failed to delete jwt token", err)
	}

	s.logger.WithFields(
		types.Field{Key: "key", Value: key},
	).Debug(ctx, "jwt token deleted")

	return nil
}

// Close 关闭存储连接
func (s *RedisStore) Close() error {
	return s.client.Close()
}
