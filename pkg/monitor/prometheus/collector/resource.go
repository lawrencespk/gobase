package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// ResourceCollector 系统资源指标收集器
// 负责收集操作系统级别的资源使用指标,如CPU、内存、磁盘等
type ResourceCollector struct {
	namespace string
	// 系统CPU使用率
	cpuUsage prometheus.Gauge

	// 系统内存使用
	memUsage prometheus.Gauge
	memTotal prometheus.Gauge
	memFree  prometheus.Gauge

	// 系统磁盘使用
	diskUsage prometheus.GaugeVec
	diskTotal prometheus.GaugeVec
	diskFree  prometheus.GaugeVec
}

// NewResourceCollector 创建资源收集器
func NewResourceCollector(namespace string) *ResourceCollector {
	c := &ResourceCollector{
		namespace: namespace,
	}

	// CPU指标
	c.cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "system",
		Name:      "cpu_usage_percent",
		Help:      "System CPU usage percentage",
	})

	// 内存指标
	c.memUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_usage_bytes",
		Help:      "System memory usage in bytes",
	})

	c.memTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_total_bytes",
		Help:      "System total memory in bytes",
	})

	c.memFree = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_free_bytes",
		Help:      "System free memory in bytes",
	})

	// 磁盘指标
	c.diskUsage = *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_usage_bytes",
		Help:      "System disk usage in bytes",
	}, []string{"path"})

	c.diskTotal = *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_total_bytes",
		Help:      "System total disk space in bytes",
	}, []string{"path"})

	c.diskFree = *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_free_bytes",
		Help:      "System free disk space in bytes",
	}, []string{"path"})

	return c
}

// Describe 实现 prometheus.Collector 接口
func (c *ResourceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsage.Desc()
	ch <- c.memUsage.Desc()
	ch <- c.memTotal.Desc()
	ch <- c.memFree.Desc()
	c.diskUsage.Describe(ch)
	c.diskTotal.Describe(ch)
	c.diskFree.Describe(ch)
}

// Collect 实现 prometheus.Collector 接口
func (c *ResourceCollector) Collect(ch chan<- prometheus.Metric) {
	// 先收集最新的指标数据
	if err := c.collect(); err != nil {
		// TODO: 考虑是否需要记录错误日志
		return
	}

	ch <- c.cpuUsage
	ch <- c.memUsage
	ch <- c.memTotal
	ch <- c.memFree
	c.diskUsage.Collect(ch)
	c.diskTotal.Collect(ch)
	c.diskFree.Collect(ch)
}

// collect 内部方法，用于收集系统资源指标
func (c *ResourceCollector) collect() error {
	// 收集CPU使用率
	cpuPercents, err := cpu.Percent(0, true)
	if err == nil && len(cpuPercents) > 0 {
		// 只取第一个CPU的使用率作为总体使用率
		c.cpuUsage.Set(cpuPercents[0])
	}

	// 收集内存使用情况
	if memInfo, err := mem.VirtualMemory(); err == nil {
		c.memUsage.Set(float64(memInfo.Used))
		c.memTotal.Set(float64(memInfo.Total))
		c.memFree.Set(float64(memInfo.Free))
	}

	// 收集磁盘使用情况
	if partitions, err := disk.Partitions(false); err == nil {
		for _, partition := range partitions {
			if usage, err := disk.Usage(partition.Mountpoint); err == nil {
				c.diskUsage.WithLabelValues(partition.Mountpoint).Set(float64(usage.Used))
				c.diskTotal.WithLabelValues(partition.Mountpoint).Set(float64(usage.Total))
				c.diskFree.WithLabelValues(partition.Mountpoint).Set(float64(usage.Free))
			}
		}
	}

	return nil
}

// Register 注册资源收集器到 Prometheus
func (c *ResourceCollector) Register() error {
	// 使用默认的注册表注册收集器
	if err := prometheus.Register(c); err != nil {
		// 如果已经注册过，则忽略错误
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return nil
		}
		return err
	}
	return nil
}
