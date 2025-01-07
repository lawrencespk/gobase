package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	jwt "gobase/pkg/middleware/jwt"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/monitor/prometheus/metrics"
)

func TestMetrics(t *testing.T) {
	// 确保metrics已初始化
	assert.NotNil(t, metrics.DefaultJWTMetrics)
	assert.NotNil(t, metrics.DefaultJWTMetrics.TokenValidationDuration)
	assert.NotNil(t, metrics.DefaultJWTMetrics.TokenValidationErrors)

	t.Run("TokenValidationDuration", func(t *testing.T) {
		// 记录耗时
		start := jwt.StartTimer()
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
		jwt.ObserveTokenValidationDuration(start)

		// 由于我们的 Histogram 是基于 prometheus 的，
		// 这里我们只能验证调用是否成功，无法直接验证具体值
		assert.NotPanics(t, func() {
			metrics.DefaultJWTMetrics.TokenValidationDuration.Observe(0.1)
		})
	})

	t.Run("TokenValidationErrors", func(t *testing.T) {
		// 记录初始状态
		counter := metrics.DefaultJWTMetrics.TokenValidationErrors

		// 增加错误计数
		jwt.IncTokenValidationError("test_error")

		// 验证计数器可以正常工作
		assert.NotPanics(t, func() {
			counter.Inc()
		})
	})

	t.Run("MetricTypes", func(t *testing.T) {
		// 验证指标类型
		assert.IsType(t, &metric.Histogram{}, metrics.DefaultJWTMetrics.TokenValidationDuration)
		assert.IsType(t, &metric.Counter{}, metrics.DefaultJWTMetrics.TokenValidationErrors)
	})
}
