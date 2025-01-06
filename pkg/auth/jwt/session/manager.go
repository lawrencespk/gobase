package session

import (
	"context"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
	"gobase/pkg/trace/jaeger"
)

// ManagerOption 会话管理器配置选项
type ManagerOption func(*Manager)

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) ManagerOption {
	return func(m *Manager) {
		m.logger = logger
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(metrics *metrics.JWTMetrics) ManagerOption {
	return func(m *Manager) {
		m.metrics = metrics
	}
}

// Manager 会话管理器
type Manager struct {
	store   Store
	logger  types.Logger
	metrics *metrics.JWTMetrics
}

// NewManager 创建新的会话管理器
func NewManager(store Store, opts ...ManagerOption) *Manager {
	m := &Manager{
		store: store,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Get 获取会话数据
func (m *Manager) Get(ctx context.Context, key string) (*Session, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "Manager.Get")
	defer span.Finish()

	startTime := time.Now()
	value, err := m.store.Get(ctx, key)
	duration := time.Since(startTime)

	if err != nil {
		if m.logger != nil {
			m.logger.Error(ctx, "Failed to get session",
				types.Field{Key: "key", Value: key},
				types.Field{Key: "error", Value: err},
			)
		}
		if m.metrics != nil {
			m.metrics.SessionErrors.WithLabelValues("get").Inc()
		}
		return nil, errors.Wrap(err, "failed to get session")
	}

	if m.metrics != nil {
		m.metrics.SessionOperations.WithLabelValues("get").Observe(duration.Seconds())
	}

	return value, nil
}

// Set 设置会话数据
func (m *Manager) Set(ctx context.Context, key string, value *Session, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "Manager.Set")
	defer span.Finish()

	startTime := time.Now()
	err := m.store.Save(ctx, value)
	duration := time.Since(startTime)

	if err != nil {
		if m.logger != nil {
			m.logger.Error(ctx, "Failed to set session",
				types.Field{Key: "key", Value: key},
				types.Field{Key: "error", Value: err},
			)
		}
		if m.metrics != nil {
			m.metrics.SessionErrors.WithLabelValues("set").Inc()
		}
		return errors.Wrap(err, "failed to set session")
	}

	if m.metrics != nil {
		m.metrics.SessionOperations.WithLabelValues("set").Observe(duration.Seconds())
	}

	return nil
}

// Delete 删除会话数据
func (m *Manager) Delete(ctx context.Context, key string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "Manager.Delete")
	defer span.Finish()

	startTime := time.Now()
	err := m.store.Delete(ctx, key)
	duration := time.Since(startTime)

	if err != nil {
		if m.logger != nil {
			m.logger.Error(ctx, "Failed to delete session",
				types.Field{Key: "key", Value: key},
				types.Field{Key: "error", Value: err},
			)
		}
		if m.metrics != nil {
			m.metrics.SessionErrors.WithLabelValues("delete").Inc()
		}
		return errors.Wrap(err, "failed to delete session")
	}

	if m.metrics != nil {
		m.metrics.SessionOperations.WithLabelValues("delete").Observe(duration.Seconds())
	}

	return nil
}

// ... 会话管理相关方法实现 ...
