package session

import (
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
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

// ... 会话管理相关方法实现 ...
