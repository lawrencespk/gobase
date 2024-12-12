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
	h := &Histogram{
		opts: opts,
	}
	h.histogram = prometheus.NewHistogram(opts)
	return h
}

// WithLabels 设置标签
func (h *Histogram) WithLabels(labels ...string) *Histogram {
	if len(labels) > 0 {
		h.vec = prometheus.NewHistogramVec(h.opts, labels)
		h.histogram = nil
	} else {
		h.histogram = prometheus.NewHistogram(h.opts)
		h.vec = nil
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

// GetCollector 返回底层的 prometheus.Collector
func (h *Histogram) GetCollector() prometheus.Collector {
	if h.vec != nil {
		return h.vec
	}
	return h.histogram
}

// GetHistogram 返回底层的 prometheus.Histogram
func (h *Histogram) GetHistogram() prometheus.Histogram {
	return h.histogram
}

// Describe 实现 prometheus.Collector 接口
func (h *Histogram) Describe(ch chan<- *prometheus.Desc) {
	if h.vec != nil {
		h.vec.Describe(ch)
	} else if h.histogram != nil {
		h.histogram.Describe(ch)
	}
}

// Collect 实现 prometheus.Collector 接口
func (h *Histogram) Collect(ch chan<- prometheus.Metric) {
	if h.vec != nil {
		h.vec.Collect(ch)
	} else if h.histogram != nil {
		h.histogram.Collect(ch)
	}
}
