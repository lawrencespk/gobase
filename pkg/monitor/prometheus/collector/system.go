package collector

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
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

	// 系统指标
	loadAverage *metric.Gauge // 系统负载
	openFDs     *metric.Gauge // 打开的文件描述符数量
	netConns    *metric.Gauge // 网络连接数量

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
	}).WithLabels([]string{})

	c.goroutines = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "goroutines_total",
		Help:      "Total number of goroutines",
	}).WithLabels([]string{})

	// 内存指标
	c.memAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_alloc_bytes",
		Help:      "Allocated memory in bytes",
	}).WithLabels([]string{})

	c.memTotal = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_total_bytes",
		Help:      "Total memory in bytes",
	}).WithLabels([]string{})

	c.memSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_sys_bytes",
		Help:      "System memory in bytes",
	}).WithLabels([]string{})

	c.memHeapAlloc = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_heap_alloc_bytes",
		Help:      "Heap memory allocated in bytes",
	}).WithLabels([]string{})

	c.memHeapSys = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "memory_heap_sys_bytes",
		Help:      "Heap memory obtained from system in bytes",
	}).WithLabels([]string{})

	// 添加系统负载指标
	c.loadAverage = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_load_average",
		Help:      "System load average",
	}).WithLabels([]string{})

	// 添加文件描述符指标
	c.openFDs = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_open_fds",
		Help:      "Number of open file descriptors",
	}).WithLabels([]string{})

	// 添加网络连接指标
	c.netConns = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_net_connections",
		Help:      "Number of network connections",
	}).WithLabels([]string{})

	// GC指标
	c.gcPause = metric.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "gc_pause_seconds",
		Help:      "GC pause time in seconds",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}).WithLabels()

	c.gcCount = metric.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "gc_count_total",
		Help:      "Total number of GC cycles",
	}).WithLabels()

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
		c.loadAverage,
		c.openFDs,
		c.netConns,
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

// Describe 实现 prometheus.Collector 接口
func (c *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	collectors := []prometheus.Collector{
		c.cpuUsage.GetCollector(),
		c.goroutines.GetCollector(),
		c.memAlloc.GetCollector(),
		c.memTotal.GetCollector(),
		c.memSys.GetCollector(),
		c.memHeapAlloc.GetCollector(),
		c.memHeapSys.GetCollector(),
		c.loadAverage.GetCollector(),
		c.openFDs.GetCollector(),
		c.netConns.GetCollector(),
		c.gcPause.GetCollector(),
		c.gcCount.GetCollector(),
	}

	// 顺序执行 Describe，避免并发问题
	for _, collector := range collectors {
		collector.Describe(ch)
	}
}

// Collect 实现 prometheus.Collector 接口
func (c *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	// 先收集最新的指标数据
	c.collect()

	collectors := []prometheus.Collector{
		c.cpuUsage.GetCollector(),
		c.goroutines.GetCollector(),
		c.memAlloc.GetCollector(),
		c.memTotal.GetCollector(),
		c.memSys.GetCollector(),
		c.memHeapAlloc.GetCollector(),
		c.memHeapSys.GetCollector(),
		c.loadAverage.GetCollector(),
		c.openFDs.GetCollector(),
		c.netConns.GetCollector(),
		c.gcPause.GetCollector(),
		c.gcCount.GetCollector(),
	}

	// 顺序执行 Collect，避免并发问题
	for _, collector := range collectors {
		collector.Collect(ch)
	}
}

// collect 内部方法，用于收集系统指标
func (c *SystemCollector) collect() {
	// 原来的 Collect 方法内容移到这里
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

	// 收集系统负载
	if loadAvg, err := getLoadAverage(); err == nil {
		c.loadAverage.Set(loadAvg)
	}

	// 收集文件描述符数量
	if fds, err := getOpenFDs(); err == nil {
		c.openFDs.Set(float64(fds))
	}

	// 收集网络连接数量
	if conns, err := getNetConnections(); err == nil {
		c.netConns.Set(float64(conns))
	}
}

// getLoadAverage 获取系统负载
func getLoadAverage() (float64, error) {
	if runtime.GOOS == "windows" {
		// Windows 不支持获取系统负载，返回 CPU 使用率作为替代
		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			return 0, fmt.Errorf("获取CPU使用率失败: %v", err)
		}
		if len(cpuPercent) > 0 {
			return cpuPercent[0], nil
		}
		return 0, nil
	}

	// Linux 系统
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, fmt.Errorf("读取系统负载失败: %v", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("无效的系统负载数据")
	}

	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("解析系统负载失败: %v", err)
	}
	return load, nil
}

// getOpenFDs 获取打开的文件描述符数量
func getOpenFDs() (int, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, fmt.Errorf("获取进程信息失败: %v", err)
	}

	if runtime.GOOS == "windows" {
		// Windows 使用 PowerShell 获取句柄数
		cmd := exec.Command("powershell", "-Command", "Get-Process -Id", strconv.Itoa(os.Getpid()), "| Select-Object -ExpandProperty HandleCount")
		output, err := cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("获取句柄数失败: %v", err)
		}
		handleCount, err := strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			return 0, fmt.Errorf("解析句柄数失败: %v", err)
		}
		return handleCount, nil
	}

	// Linux 系统
	fds, err := proc.NumFDs()
	if err != nil {
		return 0, fmt.Errorf("获取文件描述符数量失败: %v", err)
	}
	return int(fds), nil
}

// getNetConnections 获取网络连接数量
func getNetConnections() (int, error) {
	// 设置超时通道
	done := make(chan struct{})
	var (
		conns []net.ConnectionStat
		err   error
	)

	// 在goroutine中执行可能耗时的操作
	go func() {
		proc, procErr := process.NewProcess(int32(os.Getpid()))
		if procErr == nil {
			conns, err = proc.Connections()
		} else {
			err = procErr
		}
		close(done)
	}()

	// 等待结果或超时
	select {
	case <-done:
		if err != nil {
			return 0, fmt.Errorf("获取网络连接信息失败: %v", err)
		}
		return len(conns), nil
	case <-time.After(2 * time.Second): // 设置2秒超时
		return 0, fmt.Errorf("获取网络连接信息超时")
	}
}
