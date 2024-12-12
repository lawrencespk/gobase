package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Counter 计数器类型指标
type Counter struct {
	counter prometheus.Counter
	vec     *prometheus.CounterVec
	opts    prometheus.CounterOpts
}

// NewCounter 创建计数器
func NewCounter(opts prometheus.CounterOpts) *Counter {
	return &Counter{
		opts: opts,
	}
}

// WithLabels 设置标签
func (c *Counter) WithLabels(labels []string) *Counter {
	if len(labels) > 0 {
		c.vec = prometheus.NewCounterVec(c.opts, labels)
	} else {
		c.counter = prometheus.NewCounter(c.opts)
	}
	return c
}

// Inc 计数器加1
func (c *Counter) Inc() {
	if c.counter != nil {
		c.counter.Inc()
	}
}

// Add 计数器增加指定值
func (c *Counter) Add(val float64) {
	if c.counter != nil {
		c.counter.Add(val)
	}
}

// WithLabelValues 使用标签值
func (c *Counter) WithLabelValues(lvs ...string) prometheus.Counter {
	if c.vec != nil {
		return c.vec.WithLabelValues(lvs...)
	}
	return c.counter
}

// Register 注册指标
func (c *Counter) Register() error {
	var err error
	if c.vec != nil {
		err = prometheus.Register(c.vec)
	} else {
		err = prometheus.Register(c.counter)
	}
	return err
}
