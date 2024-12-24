package metrics

import (
	"gobase/pkg/monitor/prometheus/metric"
)

// RateLimitCollector 限流器指标收集器
type RateLimitCollector struct {
	// 请求计数器
	requestsTotal *metric.Counter
	// 被拒绝请求计数器
	rejectedTotal *metric.Counter
	// 限流器延迟
	limiterLatency *metric.Histogram
	// 当前活跃限流器数量
	activeLimiters *metric.Gauge
	// 等待队列长度
	waitingQueue *metric.Gauge
}

// NewRateLimitCollector 创建限流器指标收集器
func NewRateLimitCollector() *RateLimitCollector {
	c := &RateLimitCollector{}

	// 初始化请求计数器
	c.requestsTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "requests_total",
		Help:      "Total number of requests handled by rate limiter",
	}).WithLabels("key", "result") // result: allowed/rejected

	// 初始化拒绝计数器
	c.rejectedTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "rejected_total",
		Help:      "Total number of requests rejected by rate limiter",
	}).WithLabels("key")

	// 初始化延迟直方图
	c.limiterLatency = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "latency_seconds",
		Help:      "Latency of rate limiter operations",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}).WithLabels("key", "operation") // operation: allow/wait

	// 初始化活跃限流器计数
	c.activeLimiters = metric.NewGauge(metric.GaugeOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "active_limiters",
		Help:      "Number of active rate limiters",
	}).WithLabels([]string{"type"}) // type: sliding_window/token_bucket

	// 初始化等待队列长度
	c.waitingQueue = metric.NewGauge(metric.GaugeOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "waiting_queue_size",
		Help:      "Current size of waiting queue",
	}).WithLabels([]string{"key"})

	return c
}

// Register 注册所有限流器指标
func (c *RateLimitCollector) Register() error {
	return metric.Register(c)
}

// ObserveRequest 观察请求结果
func (c *RateLimitCollector) ObserveRequest(key string, allowed bool) {
	result := "allowed"
	if !allowed {
		result = "rejected"
		c.rejectedTotal.WithLabelValues(key).Inc()
	}
	c.requestsTotal.WithLabelValues(key, result).Inc()
}

// ObserveLatency 观察操作延迟
func (c *RateLimitCollector) ObserveLatency(key string, operation string, duration float64) {
	c.limiterLatency.WithLabelValues(key, operation).Observe(duration)
}

// SetActiveLimiters 设置活跃限流器数量
func (c *RateLimitCollector) SetActiveLimiters(limiterType string, count float64) {
	c.activeLimiters.WithLabelValues(limiterType).Set(count)
}

// SetWaitingQueueSize 设置等待队列长度
func (c *RateLimitCollector) SetWaitingQueueSize(key string, size float64) {
	c.waitingQueue.WithLabelValues(key).Set(size)
}

// Describe 实现 Collector 接口
func (c *RateLimitCollector) Describe(ch chan<- *metric.Desc) {
	collectors := []metric.Collector{
		c.requestsTotal.GetCollector(),
		c.rejectedTotal.GetCollector(),
		c.limiterLatency.GetCollector(),
		c.activeLimiters.GetCollector(),
		c.waitingQueue.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Describe(ch)
	}
}

// Collect 实现 Collector 接口
func (c *RateLimitCollector) Collect(ch chan<- metric.Metric) {
	collectors := []metric.Collector{
		c.requestsTotal.GetCollector(),
		c.rejectedTotal.GetCollector(),
		c.limiterLatency.GetCollector(),
		c.activeLimiters.GetCollector(),
		c.waitingQueue.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Collect(ch)
	}
}
