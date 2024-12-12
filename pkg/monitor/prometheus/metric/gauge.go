package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Gauge 仪表盘类型指标
type Gauge struct {
	gauge prometheus.Gauge
	vec   *prometheus.GaugeVec
	opts  prometheus.GaugeOpts
}

// NewGauge 创建仪表盘
func NewGauge(opts prometheus.GaugeOpts) *Gauge {
	return &Gauge{
		opts: opts,
	}
}

// WithLabels 设置标签
func (g *Gauge) WithLabels(labels []string) *Gauge {
	if len(labels) > 0 {
		g.vec = prometheus.NewGaugeVec(g.opts, labels)
	} else {
		g.gauge = prometheus.NewGauge(g.opts)
	}
	return g
}

// Set 设置值
func (g *Gauge) Set(val float64) {
	if g.gauge != nil {
		g.gauge.Set(val)
	}
}

// Inc 值加1
func (g *Gauge) Inc() {
	if g.gauge != nil {
		g.gauge.Inc()
	}
}

// Dec 值减1
func (g *Gauge) Dec() {
	if g.gauge != nil {
		g.gauge.Dec()
	}
}

// Add 增加值
func (g *Gauge) Add(val float64) {
	if g.gauge != nil {
		g.gauge.Add(val)
	}
}

// Sub 减少值
func (g *Gauge) Sub(val float64) {
	if g.gauge != nil {
		g.gauge.Sub(val)
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
	var err error
	if g.vec != nil {
		err = prometheus.Register(g.vec)
	} else {
		err = prometheus.Register(g.gauge)
	}
	return err
}
