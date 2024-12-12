package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Counter 计数器类型指标
type Counter struct {
	counter prometheus.Counter
	vec     *prometheus.CounterVec
	opts    prometheus.CounterOpts
	labels  []string // 添加labels字段保存标签
}

// NewCounter 创建计数器
func NewCounter(opts prometheus.CounterOpts) *Counter {
	c := &Counter{
		opts: opts,
	}
	// 立即创建counter，避免nil指针
	c.counter = prometheus.NewCounter(opts)
	return c
}

// WithLabels 设置标签
func (c *Counter) WithLabels(labels ...string) *Counter {
	c.labels = labels
	if len(labels) > 0 {
		c.vec = prometheus.NewCounterVec(c.opts, labels)
		c.counter = nil // 有标签时不使用普通counter
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
	if c.vec != nil {
		return prometheus.Register(c.vec)
	}
	if c.counter != nil {
		return prometheus.Register(c.counter)
	}
	return fmt.Errorf("no counter or counter vec initialized")
}

// GetCollector 返回底层的 prometheus.Collector
func (c *Counter) GetCollector() prometheus.Collector {
	if c.vec != nil {
		return c.vec
	}
	return c.counter
}

// GetCounter 返回底层的 prometheus.Counter
func (c *Counter) GetCounter() prometheus.Counter {
	return c.counter
}

// Describe 实现 prometheus.Collector 接口
func (c *Counter) Describe(ch chan<- *prometheus.Desc) {
	if c.vec != nil {
		c.vec.Describe(ch)
	} else if c.counter != nil {
		ch <- c.counter.Desc()
	}
}

// Collect 实现 prometheus.Collector 接口
func (c *Counter) Collect(ch chan<- prometheus.Metric) {
	if c.vec != nil {
		c.vec.Collect(ch)
	} else if c.counter != nil {
		ch <- c.counter
	}
}
