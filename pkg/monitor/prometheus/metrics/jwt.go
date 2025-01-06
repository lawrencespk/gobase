package metrics

import (
	"gobase/pkg/monitor/prometheus/metric"
)

// JWTMetrics JWT相关的监控指标
type JWTMetrics struct {
	// TokenAgeGauge Token生命周期指标
	TokenAgeGauge *metric.Gauge

	// TokenReuseIntervalGauge Token重用间隔指标
	TokenReuseIntervalGauge *metric.Gauge

	// TokenDuration Token操作耗时指标
	TokenDuration *metric.Histogram

	// TokenErrors Token错误计数
	TokenErrors *metric.Counter

	// SessionOperations 会话操作耗时指标
	SessionOperations *metric.Histogram

	// SessionErrors 会话操作错误计数
	SessionErrors *metric.Counter
}

var (
	// DefaultJWTMetrics 默认的JWT指标实例
	DefaultJWTMetrics = NewJWTMetrics()

	// 全局指标变量
	TokenAgeGauge           = DefaultJWTMetrics.TokenAgeGauge
	TokenReuseIntervalGauge = DefaultJWTMetrics.TokenReuseIntervalGauge
)

// NewJWTMetrics 创建新的JWT指标实例
func NewJWTMetrics() *JWTMetrics {
	m := &JWTMetrics{
		TokenAgeGauge: metric.NewGauge(metric.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_age_seconds",
			Help:      "The maximum age of JWT tokens in seconds",
		}),

		TokenReuseIntervalGauge: metric.NewGauge(metric.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_reuse_interval_seconds",
			Help:      "The reuse interval of JWT tokens in seconds",
		}),

		TokenDuration: metric.NewHistogram(metric.HistogramOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_duration_seconds",
			Help:      "Duration of JWT token operations in seconds",
		}).WithLabels("operation"),

		TokenErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_errors_total",
			Help:      "Total number of JWT token errors",
		}).WithLabels("operation", "error"),

		// 添加会话操作指标
		SessionOperations: metric.NewHistogram(metric.HistogramOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "session_duration_seconds",
			Help:      "Duration of session operations in seconds",
		}).WithLabels("operation"),

		SessionErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "session_errors_total",
			Help:      "Total number of session operation errors",
		}).WithLabels("operation"),
	}

	// 注册指标
	if err := m.TokenAgeGauge.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := m.TokenReuseIntervalGauge.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := m.TokenDuration.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := m.TokenErrors.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	// 注册新增的会话指标
	if err := m.SessionOperations.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := m.SessionErrors.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	return m
}

// NewTokenMetrics 创建新的 JWT 指标收集器
func NewTokenMetrics() *JWTMetrics {
	m := &JWTMetrics{
		TokenAgeGauge: metric.NewGauge(metric.GaugeOpts{
			Namespace: "jwt",
			Name:      "token_age_seconds",
			Help:      "Token age in seconds",
		}),
		TokenReuseIntervalGauge: metric.NewGauge(metric.GaugeOpts{
			Namespace: "jwt",
			Name:      "token_reuse_interval_seconds",
			Help:      "Token reuse interval in seconds",
		}),
		TokenErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Name:      "token_errors_total",
			Help:      "Total number of token errors",
		}).WithLabels("operation", "reason"),

		// 添加会话操作指标
		SessionOperations: metric.NewHistogram(metric.HistogramOpts{
			Namespace: "jwt",
			Name:      "session_duration_seconds",
			Help:      "Duration of session operations in seconds",
		}).WithLabels("operation"),

		SessionErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Name:      "session_errors_total",
			Help:      "Total number of session operation errors",
		}).WithLabels("operation"),
	}

	// 注册所有指标
	if err := m.registerMetrics(); err != nil {
		panic(err)
	}
	return m
}

// registerMetrics 注册所有指标
func (m *JWTMetrics) registerMetrics() error {
	// 注册 TokenAgeGauge
	if err := m.TokenAgeGauge.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			return err
		}
	}

	// 注册 TokenReuseIntervalGauge
	if err := m.TokenReuseIntervalGauge.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			return err
		}
	}

	// 注册 TokenErrors
	if err := m.TokenErrors.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			return err
		}
	}

	// 注册 SessionOperations
	if err := m.SessionOperations.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			return err
		}
	}

	// 注册 SessionErrors
	if err := m.SessionErrors.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			return err
		}
	}

	return nil
}
