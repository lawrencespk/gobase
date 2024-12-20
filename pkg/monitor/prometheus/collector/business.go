package collector

import (
	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
)

// BusinessCollector 业务指标收集器
type BusinessCollector struct {
	// 业务操作计数器
	operationTotal *metric.Counter
	// 业务操作延迟
	operationDuration *metric.Histogram
	// 业务操作错误计数
	operationErrors *metric.Counter
	// 业务队列大小
	queueSize *metric.Gauge
	// 业务处理速率
	processRate *metric.Gauge
}

// NewBusinessCollector 创建业务指标收集器
func NewBusinessCollector(namespace string) *BusinessCollector {
	c := &BusinessCollector{}

	// 初始化操作计数器
	c.operationTotal = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "business_operations_total",
		Help:      "Total number of business operations",
	}).WithLabels("operation", "status")

	// 初始化操作延迟直方图
	c.operationDuration = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "business_operation_duration_seconds",
		Help:      "Business operation latency in seconds",
		Buckets:   prometheus.DefBuckets,
	}).WithLabels("operation")

	// 初始化错误计数器
	c.operationErrors = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "business_operation_errors_total",
		Help:      "Total number of business operation errors",
	}).WithLabels("operation", "error_type")

	// 初始化队列大小仪表盘
	c.queueSize = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "business_queue_size",
		Help:      "Current size of business processing queue",
	}).WithLabels([]string{"queue_name"})

	// 初始化处理速率仪表盘
	c.processRate = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "business_process_rate",
		Help:      "Business processing rate per second",
	}).WithLabels([]string{"operation"})

	return c
}

// Register 注册所有业务指标
func (c *BusinessCollector) Register() error {
	// 直接注册 BusinessCollector 本身，而不是单独注册每个指标
	return prometheus.Register(c)
}

// ObserveOperation 观察业务操作
func (c *BusinessCollector) ObserveOperation(operation string, duration float64, err error) {
	status := "success"
	if err != nil {
		status = "error"
		c.operationErrors.WithLabelValues(operation, err.Error()).Inc()
	}

	c.operationTotal.WithLabelValues(operation, status).Inc()
	c.operationDuration.WithLabelValues(operation).Observe(duration)
}

// SetQueueSize 设置队列大小
func (c *BusinessCollector) SetQueueSize(queueName string, size float64) {
	c.queueSize.WithLabelValues(queueName).Set(size)
}

// SetProcessRate 设置处理速率
func (c *BusinessCollector) SetProcessRate(operation string, rate float64) {
	c.processRate.WithLabelValues(operation).Set(rate)
}

// Describe 实现 prometheus.Collector 接口
func (c *BusinessCollector) Describe(ch chan<- *prometheus.Desc) {
	collectors := []prometheus.Collector{
		c.operationTotal.GetCollector(),
		c.operationDuration.GetCollector(),
		c.operationErrors.GetCollector(),
		c.queueSize.GetCollector(),
		c.processRate.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Describe(ch)
	}
}

// Collect 实现 prometheus.Collector 接口
func (c *BusinessCollector) Collect(ch chan<- prometheus.Metric) {
	collectors := []prometheus.Collector{
		c.operationTotal.GetCollector(),
		c.operationDuration.GetCollector(),
		c.operationErrors.GetCollector(),
		c.queueSize.GetCollector(),
		c.processRate.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Collect(ch)
	}
}
