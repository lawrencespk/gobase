package metric

import "github.com/prometheus/client_golang/prometheus"

// CounterOpts 计数器选项
type CounterOpts = prometheus.CounterOpts

// GaugeOpts 仪表盘选项
type GaugeOpts = prometheus.GaugeOpts

// HistogramOpts 直方图选项
type HistogramOpts = prometheus.HistogramOpts

// Desc 指标描述符
type Desc = prometheus.Desc

// Metric 指标接口
type Metric = prometheus.Metric

// Collector 指标收集器接口
type Collector = prometheus.Collector

// Register 注册指标收集器
func Register(c Collector) error {
	return prometheus.Register(c)
}
