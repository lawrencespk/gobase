package session

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// SessionStore JWT会话存储
type SessionStore struct {
	store   Store
	logger  types.Logger
	metrics *metric.Counter
}

// NewSessionStore 创建会话存储实例
func NewSessionStore(opts *Options, logger types.Logger) (*SessionStore, error) {
	// 使用我们自己的 Redis 客户端，使用 Option 函数配置
	redisClient, err := redis.NewClient(
		redis.WithAddress(opts.Redis.Addr),
		redis.WithPassword(opts.Redis.Password),
		redis.WithDB(opts.Redis.DB),
		redis.WithLogger(logger),
	)
	if err != nil {
		return nil, errors.NewError(codes.InitializeError, "failed to create redis client", err)
	}

	// 验证连接
	if err := redisClient.Ping(context.Background()); err != nil {
		return nil, errors.NewError(codes.InitializeError, "failed to connect to redis", err)
	}

	// 创建存储实例
	redisStore := NewRedisStore(redisClient, opts)
	if redisStore == nil {
		redisClient.Close()
		return nil, errors.NewError(codes.InitializeError, "failed to create redis store", nil)
	}

	ss := &SessionStore{
		store:  redisStore,
		logger: logger,
	}

	// 初始化监控指标
	if opts.EnableMetrics {
		ss.metrics = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_session",
			Name:      "operations_total",
			Help:      "Total number of JWT session operations",
		})
	}

	return ss, nil
}

// Save 保存会话数据
func (s *SessionStore) Save(ctx context.Context, sessionID string, session *Session) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Save")
	defer span.Finish()

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	// 计算过期时间
	expiration := session.ExpiresAt.Sub(time.Now())
	if expiration <= 0 {
		s.logger.Error(ctx, "session already expired",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "expires_at", Value: session.ExpiresAt})
		return errors.NewError(codes.ValidationError, "session already expired", nil)
	}

	// 使用 Set 方法来处理序列化和存储
	if err := s.Set(ctx, sessionID, session, expiration); err != nil {
		s.logger.Error(ctx, "failed to save session",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "error", Value: err})
		return err
	}

	s.logger.Debug(ctx, "session saved",
		types.Field{Key: "session_id", Value: sessionID},
		types.Field{Key: "expiration", Value: expiration})

	return nil
}

// Load 加载会话数据
func (s *SessionStore) Load(ctx context.Context, sessionID string) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Load")
	defer span.Finish()

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	// 获取会话数据
	data, err := s.store.Get(ctx, sessionID)
	if err != nil {
		s.logger.Error(ctx, "failed to load session",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, err
	}

	// 反序列化会话数据
	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		s.logger.Error(ctx, "failed to unmarshal session data",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, errors.NewError(codes.SerializationError, "failed to unmarshal session data", err)
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		s.logger.Error(ctx, "session expired",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "expires_at", Value: session.ExpiresAt})
		return nil, errors.NewSessionExpiredError("session expired", nil)
	}

	s.logger.Debug(ctx, "session loaded",
		types.Field{Key: "session_id", Value: sessionID},
		types.Field{Key: "expiration", Value: session.ExpiresAt.Sub(time.Now())})
	return &session, nil
}

// Delete 删除会话数据
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Delete")
	defer span.Finish()

	if s.metrics != nil {
		defer s.metrics.Inc()
	}

	// 删除会话数据
	if err := s.store.Delete(ctx, sessionID); err != nil {
		s.logger.Error(ctx, "failed to delete session",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "error", Value: err})
		return err
	}

	s.logger.Debug(ctx, "session deleted",
		types.Field{Key: "session_id", Value: sessionID})
	return nil
}

// Close 关闭存储连接
func (s *SessionStore) Close() error {
	return s.store.Close()
}

// Get 加载会话数据
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Get")
	defer span.Finish()

	// 获取原始数据
	data, err := s.store.Get(ctx, sessionID)
	if err != nil {
		s.logger.Error(ctx, "failed to get session",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, err
	}

	// 反序列化会话数据
	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		s.logger.Error(ctx, "failed to unmarshal session data",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, errors.NewError(codes.SerializationError, "failed to unmarshal session data", err)
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		s.logger.Error(ctx, "session expired",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "expires_at", Value: session.ExpiresAt})
		return nil, errors.NewSessionExpiredError("session expired", nil)
	}

	return &session, nil
}

// Set 设置会话数据
func (s *SessionStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Set")
	defer span.Finish()

	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal session data",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "key", Value: key})
		return errors.NewError(codes.SerializationError, "failed to marshal session data", err)
	}

	// 保存会话数据
	if err := s.store.Set(ctx, key, string(data), ttl); err != nil {
		s.logger.Error(ctx, "failed to save session",
			types.Field{Key: "key", Value: key},
			types.Field{Key: "error", Value: err})
		return err
	}

	return nil
}
