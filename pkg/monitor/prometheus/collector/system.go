package collector

import (
	"gobase/pkg/monitor/prometheus/metric"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

// SystemCollector 系统指标收集器
type SystemCollector struct {
	// CPU相关指标
	cpuUsage   *metric.Gauge // CPU使用率
	goroutines *metric.Gauge // goroutine数量

	// 内存相关指标
	memAlloc     *metric.Gauge // 已分配内存
	memTotal     *metric.Gauge // 总内存
	memSys       *metric.Gauge // 系统内存
	memHeapAlloc *metric.Gauge // 堆内存分配
	memHeapSys   *metric.Gauge // 堆内存系统

	// GC相关指标
	gcPause *metric.Histogram // GC暂停时间
	gcCount *metric.Counter   // GC次数
}

// NewSystemCollector 创建系统指标收集器
func NewSystemCollector(namespace string) *SystemCollector {
	c := &SystemCollector{}

	// CPU指标
	c.cpuUsage = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "cpu_usage_percent",
		Help:      "CPU usage percentage",
	})

	c.goroutines = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "goroutines_total",
		Help:      "Total number of goroutines",
	})

	// 内存指标
	c.memAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_alloc_bytes",
		Help:      "Allocated memory in bytes",
	})

	c.memTotal = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_total_bytes",
		Help:      "Total memory in bytes",
	})

	c.memSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_sys_bytes",
		Help:      "System memory in bytes",
	})

	c.memHeapAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_heap_alloc_bytes",
		Help:      "Heap memory allocated in bytes",
	})

	c.memHeapSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_heap_sys_bytes",
		Help:      "Heap memory obtained from system in bytes",
	})

	// GC指标
	c.gcPause = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "gc_pause_seconds",
		Help:      "GC pause time in seconds",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})

	c.gcCount = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "gc_count_total",
		Help:      "Total number of GC cycles",
	})

	return c
}

// Register 注册所有系统指标
func (c *SystemCollector) Register() error {
	collectors := []interface{}{
		c.cpuUsage,
		c.goroutines,
		c.memAlloc,
		c.memTotal,
		c.memSys,
		c.memHeapAlloc,
		c.memHeapSys,
		c.gcPause,
		c.gcCount,
	}

	for _, collector := range collectors {
		if err := collector.(interface{ Register() error }).Register(); err != nil {
			return err
		}
	}
	return nil
}

// Collect 收集系统指标
func (c *SystemCollector) Collect() {
	// 收集goroutine数量
	c.goroutines.Set(float64(runtime.NumGoroutine()))

	// 收集内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.memAlloc.Set(float64(memStats.Alloc))
	c.memTotal.Set(float64(memStats.TotalAlloc))
	c.memSys.Set(float64(memStats.Sys))
	c.memHeapAlloc.Set(float64(memStats.HeapAlloc))
	c.memHeapSys.Set(float64(memStats.HeapSys))

	// 收集GC统计
	c.gcCount.Add(float64(memStats.NumGC))
	if memStats.NumGC > 0 {
		c.gcPause.Observe(float64(memStats.PauseNs[(memStats.NumGC+255)%256]) / 1e9)
	}
}
