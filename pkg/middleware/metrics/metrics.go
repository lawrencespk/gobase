package metrics

import (
	"gobase/pkg/monitor/prometheus/collector"
	"math/rand/v2"
	"time"

	"github.com/gin-gonic/gin"
)

// Config 指标中间件配置
type Config struct {
	// 是否启用采样
	EnableSampling bool
	// 采样率 0.0-1.0
	SamplingRate float64
	// 是否收集请求体大小
	EnableRequestSize bool
	// 是否收集响应体大小
	EnableResponseSize bool
	// 慢请求阈值(毫秒)
	SlowRequestThreshold int64
}

// Middleware Prometheus指标收集中间件
func Middleware(httpCollector *collector.HTTPCollector, cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 采样控制
		if cfg.EnableSampling && !shouldSample(cfg.SamplingRate) {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// 增加活跃请求计数
		httpCollector.IncActiveRequests(c.Request.Method)

		// 处理请求
		c.Next()

		// 减少活跃请求计数
		httpCollector.DecActiveRequests(c.Request.Method)

		// 收集请求指标
		duration := time.Since(start)

		// 构建指标数据
		requestSize := int64(0)
		if cfg.EnableRequestSize {
			requestSize = c.Request.ContentLength
		}

		responseSize := int64(0)
		if cfg.EnableResponseSize {
			responseSize = int64(c.Writer.Size())
		}

		// 记录请求指标
		httpCollector.ObserveRequest(
			c.Request.Method,
			path,
			c.Writer.Status(),
			duration,
			requestSize,
			responseSize,
		)

		// 记录慢请求
		if cfg.SlowRequestThreshold > 0 && duration.Milliseconds() > cfg.SlowRequestThreshold {
			httpCollector.ObserveSlowRequest(c.Request.Method, path, duration)
		}
	}
}

// shouldSample 采样判断
func shouldSample(rate float64) bool {
	return rand.Float64() < rate
}
