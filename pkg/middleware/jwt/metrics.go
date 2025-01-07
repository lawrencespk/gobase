package jwt

import (
	"time"

	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/monitor/prometheus/metrics"
)

// 定义指标名称
const (
	metricNamespace = "gobase"
	metricSubsystem = "jwt"
)

// init 初始化JWT指标
func init() {
	metrics.DefaultJWTMetrics = metrics.NewJWTMetrics()

	// 注册Token验证持续时间指标
	metrics.DefaultJWTMetrics.TokenValidationDuration = metric.NewHistogram(
		metric.HistogramOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "token_validation_duration_seconds",
			Help:      "Token validation duration in seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1},
		},
	)

	// 注册Token验证错误计数器
	metrics.DefaultJWTMetrics.TokenValidationErrors = metric.NewCounter(
		metric.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "token_validation_errors_total",
			Help:      "Total number of token validation errors",
		},
	)
}

// StartTimer 开始计时
func StartTimer() time.Time {
	return time.Now()
}

// ObserveTokenValidationDuration 观察token验证耗时
func ObserveTokenValidationDuration(start time.Time) {
	metrics.DefaultJWTMetrics.TokenValidationDuration.Observe(time.Since(start).Seconds())
}

// IncTokenValidationError 增加token验证错误计数
func IncTokenValidationError(reason string) {
	metrics.DefaultJWTMetrics.TokenValidationErrors.Inc()
}
