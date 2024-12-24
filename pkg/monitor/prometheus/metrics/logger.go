package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// LogCounter 日志计数器(按级别)
	LogCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gobase",
			Subsystem: "logger",
			Name:      "entries_total",
			Help:      "Total number of log entries by level",
		},
		[]string{"level"},
	)

	// LogLatency 日志处理延迟指标
	LogLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gobase",
			Subsystem: "logger",
			Name:      "processing_duration_seconds",
			Help:      "Log processing latency distributions",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// LogQueueSize 日志队列大小指标
	LogQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "logger",
			Name:      "queue_size",
			Help:      "Current size of the log queue",
		},
	)

	// ElkBatchSize ELK批处理大小指标
	ElkBatchSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "logger",
			Name:      "elk_batch_size",
			Help:      "Current size of the ELK batch",
		},
	)

	// ElkErrorCounter ELK错误计数器
	ElkErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gobase",
			Subsystem: "logger",
			Name:      "elk_errors_total",
			Help:      "Total number of ELK operation errors",
		},
		[]string{"operation"},
	)
)
