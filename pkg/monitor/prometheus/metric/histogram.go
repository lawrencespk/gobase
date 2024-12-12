package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Histogram 直方图类型指标
type Histogram struct {
	histogram prometheus.Histogram
	vec       *prometheus.HistogramVec
	opts      prometheus.HistogramOpts
}

// NewHistogram 创建直方图
func NewHistogram(opts prometheus.HistogramOpts) *Histogram {
	return &Histogram{
		opts: opts,
	}
}

// WithLabels 设置标签
func (h *Histogram) WithLabels(labels []string) *Histogram {
	if len(labels) > 0 {
		h.vec = prometheus.NewHistogramVec(h.opts, labels)
	} else {
		h.histogram = prometheus.NewHistogram(h.opts)
	}
	return h
}

// Observe 观察值
func (h *Histogram) Observe(val float64) {
	if h.histogram != nil {
		h.histogram.Observe(val)
	}
}

// WithLabelValues 使用标签值
func (h *Histogram) WithLabelValues(lvs ...string) prometheus.Observer {
	if h.vec != nil {
		return h.vec.WithLabelValues(lvs...)
	}
	return h.histogram
}

// Register 注册指标
func (h *Histogram) Register() error {
	var err error
	if h.vec != nil {
		err = prometheus.Register(h.vec)
	} else {
		err = prometheus.Register(h.histogram)
	}
	return err
}
