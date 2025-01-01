package metrics

import (
	"gobase/pkg/monitor/prometheus/metric"
)

var (
	// TokenGenerateCounter 令牌生成计数器
	TokenGenerateCounter = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt",
		Name:      "token_generate_total",
		Help:      "JWT token generation total count",
	}).WithLabels("status")

	// TokenValidateCounter 令牌验证计数器
	TokenValidateCounter = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt",
		Name:      "token_validate_total",
		Help:      "JWT token validation total count",
	}).WithLabels("status")

	// TokenDuration 令牌处理时间直方图
	TokenDuration = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "jwt",
		Name:      "token_duration_seconds",
		Help:      "JWT token processing duration in seconds",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
	})

	// TokenErrors 令牌错误计数器
	TokenErrors = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt",
		Name:      "token_errors_total",
		Help:      "JWT token error total count",
	}).WithLabels("type")
)

func init() {
	// 注册所有指标
	if err := TokenGenerateCounter.Register(); err != nil {
		// 忽略已注册错误，其他错误则panic
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := TokenValidateCounter.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := TokenDuration.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}

	if err := TokenErrors.Register(); err != nil {
		if err.Error() != "duplicate metrics collector registration attempted" {
			panic(err)
		}
	}
}
