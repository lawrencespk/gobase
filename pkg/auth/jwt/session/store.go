package session

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
func (s *SessionStore) Save(ctx context.Context, session *Session) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Save")
	defer span.Finish()

	// 验证会话
	if err := s.validateSession(session); err != nil {
		return err
	}

	// 计算过期时间
	expiration := time.Until(session.ExpiresAt)
	if expiration <= 0 {
		return errors.NewSessionExpiredError("session already expired", nil)
	}

	// 保存会话数据
	if err := s.store.Save(ctx, session); err != nil {
		s.logger.Error(ctx, "failed to save session",
			types.Field{Key: "token_id", Value: session.TokenID},
			types.Field{Key: "error", Value: err})
		return err
	}

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
	session, err := s.store.Get(ctx, sessionID)
	if err != nil {
		s.logger.Error(ctx, "failed to load session",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, err
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		s.logger.Error(ctx, "session expired",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "expires_at", Value: session.ExpiresAt})
		return nil, errors.NewSessionExpiredError("session expired", nil)
	}

	return session, nil
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
func (s *SessionStore) Close(ctx context.Context) error {
	return s.store.Close(ctx)
}

// Get 加载会话数据
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Get")
	defer span.Finish()

	// 获取会话数据
	session, err := s.store.Get(ctx, sessionID)
	if err != nil {
		s.logger.Error(ctx, "failed to get session",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "session_id", Value: sessionID})
		return nil, err
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		s.logger.Error(ctx, "session expired",
			types.Field{Key: "session_id", Value: sessionID},
			types.Field{Key: "expires_at", Value: session.ExpiresAt})
		return nil, errors.NewSessionExpiredError("session expired", nil)
	}

	return session, nil
}

// Set 设置会话数据 (将被废弃，请使用 Save)
func (s *SessionStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Set")
	defer span.Finish()

	session, ok := value.(*Session)
	if !ok {
		return errors.NewValidationError("value must be *Session", nil)
	}

	return s.Save(ctx, session)
}

func (s *SessionStore) Refresh(ctx context.Context, tokenID string, newExpiration time.Time) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "SessionStore.Refresh")
	defer span.Finish()

	// 获取会话
	session, err := s.Get(ctx, tokenID)
	if err != nil {
		return err
	}

	// 更新过期时间
	session.ExpiresAt = newExpiration
	session.UpdatedAt = time.Now()

	// 计算新的过期时间
	expiration := time.Until(newExpiration)
	if expiration <= 0 {
		return errors.NewSessionExpiredError("new expiration time is in the past", nil)
	}

	// 保存更新后的会话
	return s.Save(ctx, session)
}

// validateSession 验证会话数据的有效性
func (s *SessionStore) validateSession(session *Session) error {
	if session == nil {
		return errors.NewValidationError("session cannot be nil", nil)
	}

	if session.UserID == "" {
		return errors.NewValidationError("user id cannot be empty", nil)
	}

	if session.TokenID == "" {
		return errors.NewValidationError("token id cannot be empty", nil)
	}

	if session.ExpiresAt.IsZero() {
		return errors.NewValidationError("expiration time cannot be zero", nil)
	}

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.CreatedAt
	}

	return nil
}
