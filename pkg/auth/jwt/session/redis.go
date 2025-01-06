package session

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
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
	if client == nil {
		panic("redis client cannot be nil")
	}
	if opts == nil {
		opts = &Options{
			KeyPrefix: "session:",
			Log:       &types.NoopLogger{},
		}
	}
	if opts.Log == nil {
		opts.Log = &types.NoopLogger{}
	}

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
	expiration := time.Until(session.ExpiresAt)
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
func (s *RedisStore) Close(ctx context.Context) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Close")
	defer span.Finish()

	if err := s.client.Close(); err != nil {
		if s.logger != nil {
			s.logger.Error(ctx, "Failed to close redis connection",
				types.Field{Key: "error", Value: err},
			)
		}
		return errors.NewCacheError("failed to close redis connection", err)
	}
	return nil
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
func (s *RedisStore) Get(ctx context.Context, key string) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Get")
	defer span.Finish()

	// 获取字符串数据
	val, err := s.client.Get(ctx, s.prefix+key)
	if err != nil {
		// 保持原始的 RedisKeyNotFoundError
		if errors.HasErrorCode(err, codes.RedisKeyNotFoundError) {
			return nil, err
		}
		return nil, errors.NewCacheError("failed to get value", err)
	}

	// 反序列化会话数据
	var session Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, errors.NewSerializationError("failed to unmarshal session", err)
	}

	return &session, nil
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

// Ping 实现 Store 接口的 Ping 方法
func (s *RedisStore) Ping(ctx context.Context) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Ping")
	defer span.Finish()

	if s.client == nil {
		return errors.NewCacheError("redis client is nil", nil)
	}

	if err := s.client.Ping(ctx); err != nil {
		if s.logger != nil {
			s.logger.Error(ctx, "Failed to ping redis",
				types.Field{Key: "error", Value: err},
			)
		}
		return errors.NewCacheError("failed to ping redis", err)
	}
	return nil
}

// Refresh 实现 Store 接口的 Refresh 方法
func (s *RedisStore) Refresh(ctx context.Context, tokenID string, newExpiration time.Time) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "RedisStore.Refresh")
	defer span.Finish()

	// 获取现有会话
	session, err := s.Get(ctx, tokenID)
	if err != nil {
		return err
	}

	// 更新过期时间
	session.ExpiresAt = newExpiration
	session.UpdatedAt = time.Now()

	// 保存更新后的会话
	return s.Save(ctx, session)
}

// IsRedisKeyNotFoundError 检查是否为Redis键不存在错误
func IsRedisKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// 使用我们的错误包来检查错误码
	return errors.HasErrorCode(err, codes.RedisKeyNotFoundError)
}

// ... 其他方法实现 ...
