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

	return m
}
