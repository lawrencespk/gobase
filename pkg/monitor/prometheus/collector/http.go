package collector

import (
	"fmt"
	"gobase/pkg/monitor/prometheus/metric"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPCollector HTTP指标收集器
type HTTPCollector struct {
	// 请求总数
	requestTotal *metric.Counter
	// 请求延迟
	requestDuration *metric.Histogram
	// 活跃请求数
	activeRequests *metric.Gauge
	// 请求大小
	requestSize *metric.Histogram
	// 响应大小
	responseSize *metric.Histogram
	// 慢请求计数器
	slowRequests *metric.Counter
}

// NewHTTPCollector 创建HTTP收集器
func NewHTTPCollector(namespace string) *HTTPCollector {
	c := &HTTPCollector{}

	// 初始化请求总数计数器
	c.requestTotal = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
	}).WithLabels("method", "path", "status")

	// 初始化请求延迟直方图
	c.requestDuration = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency in seconds",
		Buckets:   prometheus.DefBuckets,
	}).WithLabels("method", "path")

	// 初始化活跃请求数仪表盘
	c.activeRequests = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "http_active_requests",
		Help:      "Number of active HTTP requests",
	}).WithLabels([]string{"method"})

	// 初始化请求大小直方图
	c.requestSize = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_size_bytes",
		Help:      "HTTP request size in bytes",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
	}).WithLabels("method", "path")

	// 初始化响应大小直方图
	c.responseSize = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_response_size_bytes",
		Help:      "HTTP response size in bytes",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
	}).WithLabels("method", "path")

	// 初始化慢请求计数器
	c.slowRequests = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_slow_requests_total",
		Help:      "Total number of slow HTTP requests",
	}).WithLabels("method", "path")

	return c
}

// Register 注册所有指标
func (c *HTTPCollector) Register() error {
	// 先注册各个指标
	if c.requestTotal != nil {
		if err := c.requestTotal.Register(); err != nil {
			return fmt.Errorf("failed to register request total: %w", err)
		}
	}
	if c.requestDuration != nil {
		if err := c.requestDuration.Register(); err != nil {
			return fmt.Errorf("failed to register request duration: %w", err)
		}
	}
	if c.activeRequests != nil {
		if err := c.activeRequests.Register(); err != nil {
			return fmt.Errorf("failed to register active requests: %w", err)
		}
	}
	if c.requestSize != nil {
		if err := c.requestSize.Register(); err != nil {
			return fmt.Errorf("failed to register request size: %w", err)
		}
	}
	if c.responseSize != nil {
		if err := c.responseSize.Register(); err != nil {
			return fmt.Errorf("failed to register response size: %w", err)
		}
	}
	if c.slowRequests != nil {
		if err := c.slowRequests.Register(); err != nil {
			return fmt.Errorf("failed to register slow requests: %w", err)
		}
	}

	return nil
}

// ObserveRequest 观察HTTP请求
func (c *HTTPCollector) ObserveRequest(method, path string, status int, duration time.Duration, reqSize, respSize int64) {
	statusStr := fmt.Sprintf("%d", status)

	c.requestTotal.WithLabelValues(method, path, statusStr).Inc()
	c.requestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	c.requestSize.WithLabelValues(method, path).Observe(float64(reqSize))
	c.responseSize.WithLabelValues(method, path).Observe(float64(respSize))
}

// IncActiveRequests 增加活跃请求数
func (c *HTTPCollector) IncActiveRequests(method string) {
	c.activeRequests.WithLabelValues(method).Inc()
}

// DecActiveRequests 减少活跃请求数
func (c *HTTPCollector) DecActiveRequests(method string) {
	c.activeRequests.WithLabelValues(method).Dec()
}

// ObserveSlowRequest 记录慢请求
func (c *HTTPCollector) ObserveSlowRequest(method, path string, duration time.Duration) {
	c.slowRequests.WithLabelValues(method, path).Inc()
}

// Describe 实现 prometheus.Collector 接口
func (c *HTTPCollector) Describe(ch chan<- *prometheus.Desc) {
	// 直接使用各个指标的 Describe 方法
	if c.requestTotal != nil {
		c.requestTotal.Describe(ch)
	}
	if c.requestDuration != nil {
		c.requestDuration.Describe(ch)
	}
	if c.activeRequests != nil {
		c.activeRequests.Describe(ch)
	}
	if c.requestSize != nil {
		c.requestSize.Describe(ch)
	}
	if c.responseSize != nil {
		c.responseSize.Describe(ch)
	}
	if c.slowRequests != nil {
		c.slowRequests.Describe(ch)
	}
}

// Collect 实现 prometheus.Collector 接口
func (c *HTTPCollector) Collect(ch chan<- prometheus.Metric) {
	// 直接使用各个指标的 Collect 方法
	if c.requestTotal != nil {
		c.requestTotal.Collect(ch)
	}
	if c.requestDuration != nil {
		c.requestDuration.Collect(ch)
	}
	if c.activeRequests != nil {
		c.activeRequests.Collect(ch)
	}
	if c.requestSize != nil {
		c.requestSize.Collect(ch)
	}
	if c.responseSize != nil {
		c.responseSize.Collect(ch)
	}
	if c.slowRequests != nil {
		c.slowRequests.Collect(ch)
	}
}
