package collector

import (
	"gobase/pkg/monitor/prometheus/metric"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

// RuntimeCollector Go运行时指标收集器
// 负责收集Go程序运行时的内部指标,如goroutine数量、内存分配、GC等
type RuntimeCollector struct {
	// Runtime相关指标
	goroutines *metric.Gauge // goroutine数量

	// 运行时内存指标
	memAlloc     *metric.Gauge // 运行时已分配内存
	memTotal     *metric.Gauge // 运行时总分配内存
	memSys       *metric.Gauge // 运行时系统内存
	memHeapAlloc *metric.Gauge // 运行时堆内存分配
	memHeapSys   *metric.Gauge // 运行时堆内存系统

	// GC相关指标
	gcPause *metric.Histogram // GC暂停时间
	gcCount *metric.Counter   // GC次数
}

// NewRuntimeCollector 创建运行时指标收集器
func NewRuntimeCollector(namespace string) *RuntimeCollector {
	c := &RuntimeCollector{}

	// Goroutine指标
	c.goroutines = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_goroutines_total",
		Help:      "Total number of goroutines",
	})

	// 运行时内存指标
	c.memAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_memory_alloc_bytes",
		Help:      "Runtime allocated memory in bytes",
	})

	c.memTotal = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_memory_total_bytes",
		Help:      "Runtime total allocated memory in bytes",
	})

	c.memSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_memory_sys_bytes",
		Help:      "Runtime system memory in bytes",
	})

	c.memHeapAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_memory_heap_alloc_bytes",
		Help:      "Runtime heap memory allocated in bytes",
	})

	c.memHeapSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "runtime_memory_heap_sys_bytes",
		Help:      "Runtime heap memory obtained from system in bytes",
	})

	// GC指标
	c.gcPause = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "runtime_gc_pause_seconds",
		Help:      "Runtime GC pause time in seconds",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})

	c.gcCount = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "runtime_gc_count_total",
		Help:      "Runtime total number of GC cycles",
	})

	return c
}

// Register 注册所有运行时指标
func (c *RuntimeCollector) Register() error {
	collectors := []interface{}{
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

// Collect 收集运行时指标
func (c *RuntimeCollector) Collect() {
	// 收集goroutine数量
	c.goroutines.Set(float64(runtime.NumGoroutine()))

	// 收集运行时内存统计
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
