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

	// TokenValidateSuccess Token验证成功计数
	TokenValidateSuccess *metric.Counter

	// TokenValidateFailure Token验证失败计数
	TokenValidateFailure *metric.Counter

	// TokenValidationDuration Token验证耗时
	TokenValidationDuration *metric.Histogram

	// TokenValidationErrors Token验证错误计数
	TokenValidationErrors *metric.Counter
}

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
		}),

		TokenErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_errors_total",
			Help:      "Total number of JWT token errors",
		}),

		SessionOperations: metric.NewHistogram(metric.HistogramOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "session_duration_seconds",
			Help:      "Duration of session operations in seconds",
		}),

		SessionErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "session_errors_total",
			Help:      "Total number of session operation errors",
		}),

		TokenValidateSuccess: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_validate_success_total",
			Help:      "Total number of successful token validations",
		}),

		TokenValidateFailure: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_validate_failure_total",
			Help:      "Total number of failed token validations",
		}),

		TokenValidationDuration: metric.NewHistogram(metric.HistogramOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_validation_duration_seconds",
			Help:      "Token validation duration in seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1},
		}),

		TokenValidationErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt",
			Name:      "token_validation_errors_total",
			Help:      "Total number of token validation errors",
		}),
	}

	return m
}

// DefaultJWTMetrics 默认的JWT指标实例
var DefaultJWTMetrics = NewJWTMetrics()
