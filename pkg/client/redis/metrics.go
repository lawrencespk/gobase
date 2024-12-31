package redis

import (
	"context"
	"time"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

// 定义指标结构
type redisMetrics struct {
	namespace             string
	operationTotal        *metric.Counter
	operationDuration     *metric.Histogram
	poolActiveConnections *metric.Gauge
	poolIdleConnections   *metric.Gauge
	errorTotal            *metric.Counter
}

var metrics *redisMetrics

// initMetrics 初始化指标
func initMetrics(registry prometheus.Registerer, namespace string) error {
	// 如果已经初始化且命名空间相同，则直接返回
	if metrics != nil && metrics.namespace == namespace {
		return nil
	}

	// 如果已经初始化但命名空间不同，则需要重新初始化
	if metrics != nil {
		// 取消注册旧的指标
		registry.Unregister(metrics.operationTotal)
		registry.Unregister(metrics.operationDuration)
		registry.Unregister(metrics.poolActiveConnections)
		registry.Unregister(metrics.poolIdleConnections)
		registry.Unregister(metrics.errorTotal)
	}

	// 创建新的指标
	metrics = &redisMetrics{
		namespace: namespace, // 添加 namespace 字段
		operationTotal: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "operations_total",
			Help:      "Total number of Redis operations",
		}).WithLabels("operation", "status"),

		operationDuration: metric.NewHistogram(metric.HistogramOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "operation_duration_seconds",
			Help:      "Redis operation latency in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}).WithLabels("operation"),

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

		errorTotal: metric.NewCounter(metric.CounterOpts{
			Namespace: namespace,
			Subsystem: "redis",
			Name:      "errors_total",
			Help:      "Total number of Redis errors",
		}).WithLabels("type"),
	}

	// 注册所有指标
	collectors := []prometheus.Collector{
		metrics.operationTotal,
		metrics.operationDuration,
		metrics.poolActiveConnections,
		metrics.poolIdleConnections,
		metrics.errorTotal,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			// 忽略已注册的错误
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				return err
			}
		}
	}

	return nil
}

// withMetrics 包装Redis操作并记录监控指标
func withMetrics(ctx context.Context, operation string, registry prometheus.Registerer, f func() error) error {
	// 从上下文中获取 namespace
	namespace := "gobase"
	if client, ok := ctx.Value(clientKey).(*client); ok && client != nil {
		// 优先使用客户端配置的 namespace
		if client.options != nil && client.options.MetricsNamespace != "" {
			namespace = client.options.MetricsNamespace
		}
	} else if ns, ok := ctx.Value(namespaceKey).(string); ok && ns != "" {
		// 如果没有客户端配置，则尝试从上下文中获取
		namespace = ns
	}

	// 确保指标已初始化
	if err := initMetrics(registry, namespace); err != nil {
		return err
	}

	startTime := time.Now()
	err := f()
	duration := time.Since(startTime).Seconds()

	if metrics != nil {
		// 记录操作状态和延迟
		status := "success"
		if err != nil {
			status = "error"
			// 记录错误
			errType := "unknown"
			if redisErr, ok := err.(redis.Error); ok {
				errType = redisErr.Error()
			}
			metrics.errorTotal.WithLabelValues(errType).Inc()
		}

		// 记录操作指标
		metrics.operationTotal.WithLabelValues(operation, status).Inc()
		metrics.operationDuration.WithLabelValues(operation).Observe(duration)

		// 更新连接池指标
		if client, ok := ctx.Value(clientKey).(*client); ok && client != nil {
			if stats := client.Pool().Stats(); stats != nil {
				metrics.poolActiveConnections.Set(float64(stats.TotalConns))
				metrics.poolIdleConnections.Set(float64(stats.IdleConns))
			}
		}
	}

	return err
}

// 定义上下文键类型
type contextKey string

const (
	clientKeyStr    contextKey = "redis_client"
	namespaceKeyStr contextKey = "redis_namespace"
)

// 定义上下文键
var (
	clientKey    = clientKeyStr
	namespaceKey = namespaceKeyStr
)

// 添加 pipelineMetrics 结构体
type pipelineMetrics struct {
	commandsTotal    *prometheus.CounterVec
	executionLatency *prometheus.HistogramVec
	errorTotal       *prometheus.CounterVec
}

// 添加 newPipelineMetrics 函数
func newPipelineMetrics(namespace string) *pipelineMetrics {
	return &pipelineMetrics{
		commandsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pipeline_commands_total",
				Help:      "Total number of commands added to pipeline",
			},
			[]string{"operation"},
		),
		executionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "pipeline_execution_duration_seconds",
				Help:      "Pipeline execution latency in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"status"},
		),
		errorTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pipeline_errors_total",
				Help:      "Total number of pipeline errors",
			},
			[]string{"operation", "error_type"},
		),
	}
}
