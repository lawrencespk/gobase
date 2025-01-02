package redis

import (
	"gobase/pkg/monitor/prometheus/metric"
)

// RedisMetrics Redis指标收集器
type RedisMetrics struct {
	// 命令执行指标
	commandDuration *metric.Histogram
	commandErrors   *metric.Counter

	// 连接池指标
	poolActiveConnections *metric.Gauge
	poolIdleConnections   *metric.Gauge
	poolTotalConnections  *metric.Gauge
	poolWaitCount         *metric.Gauge
	poolTimeoutCount      *metric.Counter
	poolHitCount          *metric.Counter
	poolMissCount         *metric.Counter
}

// NewRedisMetrics 创建Redis指标收集器
func NewRedisMetrics(namespace string) *RedisMetrics {
	m := &RedisMetrics{
		// 命令执行指标
		commandDuration: metric.NewHistogram(metric.HistogramOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "command_duration_seconds",
			Help:      "Redis command execution duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}),

		commandErrors: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "command_errors_total",
			Help:      "Total number of Redis command errors",
		}),

		// 连接池指标
		poolActiveConnections: metric.NewGauge(metric.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_active_connections",
			Help:      "Number of active connections in the pool",
		}),

		poolIdleConnections: metric.NewGauge(metric.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_idle_connections",
			Help:      "Number of idle connections in the pool",
		}),

		poolTotalConnections: metric.NewGauge(metric.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_total_connections",
			Help:      "Total number of connections in the pool",
		}),

		poolWaitCount: metric.NewGauge(metric.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_wait_count",
			Help:      "Number of connections waited for",
		}),

		poolTimeoutCount: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_timeout_total",
			Help:      "Total number of connection timeouts",
		}),

		poolHitCount: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_hit_total",
			Help:      "Total number of connection pool hits",
		}),

		poolMissCount: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "pool_miss_total",
			Help:      "Total number of connection pool misses",
		}),
	}

	// 注册所有指标
	m.registerMetrics()

	return m
}

// registerMetrics 注册所有指标
func (m *RedisMetrics) registerMetrics() {
	// 注册命令执行指标
	m.commandDuration.Register()
	m.commandErrors.Register()

	// 注册连接池指标
	m.poolActiveConnections.Register()
	m.poolIdleConnections.Register()
	m.poolTotalConnections.Register()
	m.poolWaitCount.Register()
	m.poolTimeoutCount.Register()
	m.poolHitCount.Register()
	m.poolMissCount.Register()
}

// ObserveCommandExecution 观察命令执行
func (m *RedisMetrics) ObserveCommandExecution(duration float64, err error) {
	m.commandDuration.Observe(duration)
	if err != nil {
		m.commandErrors.Inc()
	}
}

// UpdatePoolStats 更新连接池统计信息
func (m *RedisMetrics) UpdatePoolStats(stats *PoolStats) {
	m.poolActiveConnections.Set(float64(stats.ActiveCount))
	m.poolIdleConnections.Set(float64(stats.IdleCount))
	m.poolTotalConnections.Set(float64(stats.TotalCount))
	m.poolWaitCount.Set(float64(stats.WaitCount))
	m.poolTimeoutCount.Add(float64(stats.TimeoutCount))
	m.poolHitCount.Add(float64(stats.HitCount))
	m.poolMissCount.Add(float64(stats.MissCount))
}

// pipelineMetrics Pipeline指标收集器
type pipelineMetrics struct {
	// 命令执行总数
	commandsTotal *metric.Counter
	// 错误总数
	errorTotal *metric.Counter
	// 执行延迟
	executionLatency *metric.Histogram
}

// newPipelineMetrics 创建Pipeline指标收集器
func newPipelineMetrics(namespace string) *pipelineMetrics {
	m := &pipelineMetrics{
		// 命令执行总数
		commandsTotal: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pipeline",
			Name:      "commands_total",
			Help:      "Total number of pipeline commands",
		}),

		// 错误总数
		errorTotal: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pipeline",
			Name:      "errors_total",
			Help:      "Total number of pipeline errors",
		}),

		// 执行延迟
		executionLatency: metric.NewHistogram(metric.HistogramOpts{
			Namespace: namespace,
			Subsystem: "redis_pipeline",
			Name:      "execution_latency_seconds",
			Help:      "Pipeline execution latency in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}),
	}

	// 注册所有指标
	m.register()

	return m
}

func (m *pipelineMetrics) register() {
	// 注册所有指标
	m.commandsTotal.Register()
	m.errorTotal.Register()
	m.executionLatency.Register()
}

// Describe 实现 metric.Collector 接口
func (m *RedisMetrics) Describe(ch chan<- *metric.Desc) {
	if m == nil {
		return
	}
	collectors := []metric.Collector{
		m.poolActiveConnections,
		m.poolIdleConnections,
		m.poolTotalConnections,
		m.poolWaitCount,
		m.poolTimeoutCount,
		m.poolHitCount,
		m.poolMissCount,
		m.commandErrors,
		m.commandDuration,
	}

	for _, collector := range collectors {
		collector.Describe(ch)
	}
}

// Collect 实现 metric.Collector 接口
func (m *RedisMetrics) Collect(ch chan<- metric.Metric) {
	if m == nil {
		return
	}
	collectors := []metric.Collector{
		m.poolActiveConnections,
		m.poolIdleConnections,
		m.poolTotalConnections,
		m.poolWaitCount,
		m.poolTimeoutCount,
		m.poolHitCount,
		m.poolMissCount,
		m.commandErrors,
		m.commandDuration,
	}

	for _, collector := range collectors {
		collector.Collect(ch)
	}
}

// GetCollector 返回 metric.Collector 接口
func (m *RedisMetrics) GetCollector() metric.Collector {
	return m
}
