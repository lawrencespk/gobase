package logger

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics Prometheus指标收集器
type PrometheusMetrics struct {
	// 请求计数器
	requestCounter *prometheus.CounterVec
	// 请求延迟直方图
	requestLatency *prometheus.HistogramVec
	// 请求大小直方图
	requestSize *prometheus.HistogramVec
	// 响应大小直方图
	responseSize *prometheus.HistogramVec
	// 活跃请求计数器
	activeRequests *prometheus.GaugeVec
	// 配置
	config *MetricsConfig
}

// NewPrometheusMetrics 创建Prometheus指标收集器
func NewPrometheusMetrics(config *MetricsConfig) (*PrometheusMetrics, error) {
	if !config.Enable {
		return nil, nil
	}

	m := &PrometheusMetrics{
		config: config,
	}

	// 请求计数器
	m.requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: config.Prefix + "_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// 请求延迟直方图
	if config.EnableLatencyHistogram {
		m.requestLatency = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    config.Prefix + "_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: config.Buckets,
			},
			[]string{"method", "path", "status"},
		)
	}

	// 请求大小直方图
	if config.EnableSizeHistogram {
		m.requestSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    config.Prefix + "_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 100B ~ 1GB
			},
			[]string{"method", "path"},
		)
	}

	// 响应大小直方图
	if config.EnableSizeHistogram {
		m.responseSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    config.Prefix + "_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 100B ~ 1GB
			},
			[]string{"method", "path"},
		)
	}

	// 活跃请求计数器
	m.activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: config.Prefix + "_active_requests",
			Help: "Number of active HTTP requests",
		},
		[]string{"method", "path"},
	)

	// 注册指标
	if err := m.registerMetrics(); err != nil {
		return nil, err
	}

	return m, nil
}

// registerMetrics 注册指标
func (m *PrometheusMetrics) registerMetrics() error {
	collectors := []prometheus.Collector{
		m.requestCounter,
		m.activeRequests,
	}

	if m.requestLatency != nil {
		collectors = append(collectors, m.requestLatency)
	}

	if m.requestSize != nil {
		collectors = append(collectors, m.requestSize)
	}

	if m.responseSize != nil {
		collectors = append(collectors, m.responseSize)
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			return err
		}
	}

	return nil
}

// CollectRequest 实现MetricsCollector接口
func (m *PrometheusMetrics) CollectRequest(ctx *gin.Context, param *MetricsParam) {
	if !m.config.Enable {
		return
	}

	labels := []string{
		param.Method,
		param.Path,
		strconv.Itoa(param.StatusCode),
	}

	// 增加请求计数
	m.requestCounter.WithLabelValues(labels...).Inc()

	// 记录请求延迟
	if m.requestLatency != nil {
		m.requestLatency.WithLabelValues(labels...).Observe(param.Latency.Seconds())
	}

	// 记录请求大小
	if m.requestSize != nil {
		m.requestSize.WithLabelValues(param.Method, param.Path).Observe(float64(param.RequestSize))
	}

	// 记录响应大小
	if m.responseSize != nil {
		m.responseSize.WithLabelValues(param.Method, param.Path).Observe(float64(param.ResponseSize))
	}

	// 更新活跃请求数
	m.activeRequests.WithLabelValues(param.Method, param.Path).Dec()
}

// BeginRequest 开始请求监控
func (m *PrometheusMetrics) BeginRequest(method, path string) {
	if !m.config.Enable {
		return
	}
	m.activeRequests.WithLabelValues(method, path).Inc()
}

// Close 关闭指标收集器
func (m *PrometheusMetrics) Close() error {
	if !m.config.Enable {
		return nil
	}

	collectors := []prometheus.Collector{
		m.requestCounter,
		m.activeRequests,
	}

	if m.requestLatency != nil {
		collectors = append(collectors, m.requestLatency)
	}

	if m.requestSize != nil {
		collectors = append(collectors, m.requestSize)
	}

	if m.responseSize != nil {
		collectors = append(collectors, m.responseSize)
	}

	for _, collector := range collectors {
		prometheus.Unregister(collector)
	}

	return nil
}
