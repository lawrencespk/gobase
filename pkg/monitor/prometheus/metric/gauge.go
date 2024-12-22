package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Gauge 仪表盘类型指标
type Gauge struct {
	gauge  prometheus.Gauge
	vec    *prometheus.GaugeVec
	opts   prometheus.GaugeOpts
	labels []string
}

// NewGauge 创建仪表盘
func NewGauge(opts prometheus.GaugeOpts) *Gauge {
	g := &Gauge{
		opts: opts,
	}
	// 立即创建gauge，避免nil指针
	g.gauge = prometheus.NewGauge(opts)
	return g
}

// WithLabels 设置标签
func (g *Gauge) WithLabels(labels []string) *Gauge {
	g.labels = labels
	if len(labels) > 0 {
		g.vec = prometheus.NewGaugeVec(g.opts, labels)
		g.gauge = nil // 有标签时不使用普通gauge
	}
	return g
}

// Inc 仪表盘加1
func (g *Gauge) Inc() {
	if g.vec != nil {
		g.vec.WithLabelValues().Inc()
		return
	}
	if g.gauge != nil {
		g.gauge.Inc()
	}
}

// Dec 仪表盘减1
func (g *Gauge) Dec() {
	if g.vec != nil {
		g.vec.WithLabelValues().Dec()
		return
	}
	if g.gauge != nil {
		g.gauge.Dec()
	}
}

// Add 仪表盘增加指定值
func (g *Gauge) Add(val float64) {
	if g.vec != nil {
		g.vec.WithLabelValues().Add(val)
		return
	}
	if g.gauge != nil {
		g.gauge.Add(val)
	}
}

// Sub 仪表盘减少指定值
func (g *Gauge) Sub(val float64) {
	if g.vec != nil {
		g.vec.WithLabelValues().Sub(val)
		return
	}
	if g.gauge != nil {
		g.gauge.Sub(val)
	}
}

// Set 仪表盘设置指定值
func (g *Gauge) Set(val float64) {
	if g.vec != nil {
		g.vec.WithLabelValues().Set(val)
		return
	}
	if g.gauge != nil {
		g.gauge.Set(val)
	}
}

// WithLabelValues 使用标签值
func (g *Gauge) WithLabelValues(lvs ...string) prometheus.Gauge {
	if g.vec != nil {
		return g.vec.WithLabelValues(lvs...)
	}
	return g.gauge
}

// Register 注册指标
func (g *Gauge) Register() error {
	if g.vec != nil {
		return prometheus.Register(g.vec)
	}
	if g.gauge != nil {
		return prometheus.Register(g.gauge)
	}
	return fmt.Errorf("no gauge or gauge vec initialized")
}

// GetCollector 返回底层的 prometheus.Collector
func (g *Gauge) GetCollector() prometheus.Collector {
	if g.vec != nil {
		return g.vec
	}
	return g.gauge
}

// Describe 实现 prometheus.Collector 接口
func (g *Gauge) Describe(ch chan<- *prometheus.Desc) {
	if g.vec != nil {
		g.vec.Describe(ch)
	} else if g.gauge != nil {
		ch <- g.gauge.Desc()
	}
}

// Collect 实现 prometheus.Collector 接口
func (g *Gauge) Collect(ch chan<- prometheus.Metric) {
	if g.vec != nil {
		g.vec.Collect(ch)
	} else if g.gauge != nil {
		ch <- g.gauge
	}
}

// GetGauge 返回底层的 prometheus.Gauge
func (g *Gauge) GetGauge() prometheus.Gauge {
	return g.gauge
}
