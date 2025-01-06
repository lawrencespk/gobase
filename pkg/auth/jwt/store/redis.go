package store

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/trace/jaeger"
)

// RedisTokenStore Redis实现的Token存储
type RedisTokenStore struct {
	client redis.Client
	prefix string
	logger types.Logger
}

// NewRedisTokenStore 创建Redis Token存储实例
func NewRedisTokenStore(client redis.Client, options *Options, logger types.Logger) *RedisTokenStore {
	return &RedisTokenStore{
		client: client,
		prefix: options.KeyPrefix + "token:",
		logger: logger,
	}
}

// Set 存储Token信息
func (s *RedisTokenStore) Set(ctx context.Context, token string, info *jwt.TokenInfo, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.set")
	if span != nil {
		defer span.Finish()
	}

	// 序列化Token信息
	data, err := json.Marshal(info)
	if err != nil {
		return errors.NewSerializationError("failed to marshal token info", err)
	}

	// 存储到Redis
	key := s.prefix + token
	if err := s.client.Set(ctx, key, string(data), expiration); err != nil {
		s.logger.Error(ctx, "failed to store token",
			types.Field{Key: "token", Value: token},
			types.Field{Key: "error", Value: err},
		)
		return errors.NewRedisCommandError("failed to store token", err)
	}

	s.logger.Debug(ctx, "token stored",
		types.Field{Key: "token", Value: token},
		types.Field{Key: "expiration", Value: expiration},
	)

	return nil
}

// Get 获取Token信息
func (s *RedisTokenStore) Get(ctx context.Context, token string) (*jwt.TokenInfo, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.get")
	if span != nil {
		defer span.Finish()
	}

	// 从Redis获取
	key := s.prefix + token
	data, err := s.client.Get(ctx, key)
	if err != nil {
		if err == redis.ErrNil {
			return nil, err
		}
		return nil, errors.NewRedisCommandError("failed to get token", err)
	}

	// 反序列化Token信息
	var info jwt.TokenInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, errors.NewSerializationError("failed to unmarshal token info", err)
	}

	return &info, nil
}

// Delete 删除Token信息
func (s *RedisTokenStore) Delete(ctx context.Context, token string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.redis.delete")
	if span != nil {
		defer span.Finish()
	}

	key := s.prefix + token
	_, err := s.client.Del(ctx, key)
	if err != nil {
		s.logger.Error(ctx, "failed to delete token",
			types.Field{Key: "token", Value: token},
			types.Field{Key: "error", Value: err},
		)
		return errors.NewRedisCommandError("failed to delete token", err)
	}

	s.logger.Debug(ctx, "token deleted",
		types.Field{Key: "token", Value: token},
	)

	return nil
}

// Close 关闭存储连接
func (s *RedisTokenStore) Close() error {
	return s.client.Close()
}
