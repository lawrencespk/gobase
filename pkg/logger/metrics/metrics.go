package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 日志计数器(按级别)
	LogCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_entries_total",
			Help: "Total number of log entries by level",
		},
		[]string{"level"},
	)

	// 日志处理延迟
	LogLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_processing_duration_seconds",
			Help:    "Log processing latency distributions",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"}, // 如: write, flush, rotate
	)

	// 日志队列大小
	LogQueueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "log_queue_size",
			Help: "Current size of the log queue",
		},
	)

	// ELK 批处理指标
	ElkBatchSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "elk_batch_size",
			Help: "Current size of the ELK batch",
		},
	)

	// ELK 错误计数
	ElkErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elk_errors_total",
			Help: "Total number of ELK operation errors",
		},
		[]string{"operation"},
	)
)

func init() {
	// 注册所有指标
	prometheus.MustRegister(
		LogCounter,
		LogLatency,
		LogQueueSize,
		ElkBatchSize,
		ElkErrorCounter,
	)
}
