package collector

import (
	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
)

// RedisCollector Redis指标收集器
type RedisCollector struct {
	// 连接池指标
	poolActive  *metric.Gauge   // 活跃连接数
	poolIdle    *metric.Gauge   // 空闲连接数
	poolTotal   *metric.Gauge   // 总连接数
	poolWait    *metric.Counter // 等待连接的次数
	poolTimeout *metric.Counter // 获取连接超时次数
	poolHits    *metric.Counter // 命中连接的次数
	poolMisses  *metric.Counter // 未命中连接的次数

	// 操作指标
	cmdTotal    *metric.Counter   // 命令执行总数
	cmdErrors   *metric.Counter   // 命令执行错误数
	cmdDuration *metric.Histogram // 命令执行耗时
}

// NewRedisCollector 创建Redis指标收集器
func NewRedisCollector(namespace string) *RedisCollector {
	c := &RedisCollector{
		poolActive: metric.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "active_connections",
			Help:      "The number of active connections in the pool",
		}),
		poolIdle: metric.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "idle_connections",
			Help:      "The number of idle connections in the pool",
		}),
		poolTotal: metric.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "total_connections",
			Help:      "The total number of connections in the pool",
		}),
		poolWait: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "wait_total",
			Help:      "The total number of times a connection was waited for",
		}),
		poolTimeout: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "timeout_total",
			Help:      "The total number of times a connection timed out",
		}),
		poolHits: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "hits_total",
			Help:      "The total number of times a connection was reused",
		}),
		poolMisses: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis_pool",
			Name:      "misses_total",
			Help:      "The total number of times a new connection was created",
		}),
		cmdTotal: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "commands_total",
			Help:      "The total number of commands executed",
		}),
		cmdErrors: metric.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "command_errors_total",
			Help:      "The total number of command execution errors",
		}),
		cmdDuration: metric.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "command_duration_seconds",
			Help:      "The duration of command execution",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}),
	}

	return c
}

// UpdatePoolStats 更新连接池统计信息
func (c *RedisCollector) UpdatePoolStats(stats *PoolStats) {
	if stats == nil {
		return
	}

	c.poolActive.Set(float64(stats.ActiveCount))
	c.poolIdle.Set(float64(stats.IdleCount))
	c.poolTotal.Set(float64(stats.TotalCount))
	c.poolWait.Add(float64(stats.WaitCount))
	c.poolTimeout.Add(float64(stats.TimeoutCount))
	c.poolHits.Add(float64(stats.HitCount))
	c.poolMisses.Add(float64(stats.MissCount))
}

// SetPoolStats 更新连接池统计信息
func (c *RedisCollector) SetPoolStats(hits, misses, timeouts, totalConns, idleConns, staleConns float64) {
	if c == nil {
		return
	}
	c.poolHits.Add(hits)
	c.poolMisses.Add(misses)
	c.poolTimeout.Add(timeouts)
	c.poolTotal.Set(totalConns)
	c.poolIdle.Set(idleConns)
	// 可以添加一个新的指标来跟踪过期连接
	// c.poolStale.Set(staleConns)
}

// SetPoolConfig 更新连接池配置指标
func (c *RedisCollector) SetPoolConfig(poolSize, minIdleConns float64) {
	if c == nil {
		return
	}
	// 这些是配置值，使用 Gauge 类型
	c.poolTotal.Set(poolSize)
	c.poolIdle.Set(minIdleConns)
}

// SetCurrentPoolSize 更新当前连接池大小
func (c *RedisCollector) SetCurrentPoolSize(size float64) {
	if c == nil {
		return
	}
	c.poolActive.Set(size)
}

// ObserveCommandExecution 观察命令执行情况
func (c *RedisCollector) ObserveCommandExecution(duration float64, err error) {
	if c == nil {
		return
	}
	c.cmdTotal.Inc()
	if err != nil {
		c.cmdErrors.Inc()
	}
	c.cmdDuration.Observe(duration)
}

// Describe 实现 prometheus.Collector 接口
func (c *RedisCollector) Describe(ch chan<- *prometheus.Desc) {
	collectors := []prometheus.Collector{
		c.poolActive.GetCollector(),
		c.poolIdle.GetCollector(),
		c.poolTotal.GetCollector(),
		c.poolWait.GetCollector(),
		c.poolTimeout.GetCollector(),
		c.poolHits.GetCollector(),
		c.poolMisses.GetCollector(),
		c.cmdTotal.GetCollector(),
		c.cmdErrors.GetCollector(),
		c.cmdDuration.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Describe(ch)
	}
}

// Collect 实现 prometheus.Collector 接口
func (c *RedisCollector) Collect(ch chan<- prometheus.Metric) {
	collectors := []prometheus.Collector{
		c.poolActive.GetCollector(),
		c.poolIdle.GetCollector(),
		c.poolTotal.GetCollector(),
		c.poolWait.GetCollector(),
		c.poolTimeout.GetCollector(),
		c.poolHits.GetCollector(),
		c.poolMisses.GetCollector(),
		c.cmdTotal.GetCollector(),
		c.cmdErrors.GetCollector(),
		c.cmdDuration.GetCollector(),
	}

	for _, collector := range collectors {
		collector.Collect(ch)
	}
}

// PoolStats 连接池统计信息
type PoolStats struct {
	ActiveCount  int64
	IdleCount    int64
	TotalCount   int64
	WaitCount    int64
	TimeoutCount int64
	HitCount     int64
	MissCount    int64
}
