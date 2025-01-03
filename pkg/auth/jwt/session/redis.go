package session

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/trace/jaeger"
)

// RedisStore Redis实现的会话存储
type RedisStore struct {
	client redis.Client
	prefix string
	logger types.Logger
}

// NewRedisStore 创建Redis存储实例
func NewRedisStore(client redis.Client, opts *Options) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: opts.KeyPrefix + "session:",
		logger: opts.Log,
	}
}

// Save 保存会话
func (s *RedisStore) Save(ctx context.Context, session *Session) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "session.store.redis.save")
	defer span.Finish()

	// 序列化会话数据
	data, err := json.Marshal(session)
	if err != nil {
		return errors.NewSerializationError("failed to marshal session", err)
	}

	// 计算过期时间
	expiration := session.ExpiresAt.Sub(time.Now())
	if expiration <= 0 {
		return errors.NewSessionExpiredError("session already expired", nil)
	}

	// 使用事务管道执行多个操作
	pipe := s.client.TxPipeline()

	// 存储会话数据
	sessionKey := s.prefix + session.TokenID
	pipe.Set(ctx, sessionKey, string(data), expiration)

	// 维护用户会话索引
	userKey := s.prefix + "user:" + session.UserID
	pipe.SAdd(ctx, userKey, session.TokenID)
	pipe.ExpireAt(ctx, userKey, session.ExpiresAt)

	// 执行管道操作
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error(ctx, "failed to save session",
			types.Field{Key: "session_id", Value: session.TokenID},
			types.Field{Key: "error", Value: err},
		)
		return errors.NewCacheError("failed to save session", err)
	}

	return nil
}

// Close 实现 Store 接口
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Delete 实现 Store 接口的 Delete 方法
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Delete")
	defer span.Finish()

	_, err := s.client.Del(ctx, s.prefix+sessionID)
	if err != nil {
		return errors.NewCacheError("failed to delete session", err)
	}
	return nil
}

// Get 实现 Store 接口的 Get 方法
func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Get")
	defer span.Finish()

	// 直接返回字符串数据，不做反序列化
	val, err := s.client.Get(ctx, s.prefix+key)
	if err != nil {
		return "", errors.NewCacheError("failed to get value", err)
	}

	return val, nil
}

// Set 实现 Store 接口的 Set 方法
func (s *RedisStore) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Set")
	defer span.Finish()

	err := s.client.Set(ctx, s.prefix+key, value, expiration)
	if err != nil {
		return errors.NewCacheError("failed to set value", err)
	}
	return nil
}

// ... 其他方法实现 ...
