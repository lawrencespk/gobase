package collector

import (
	"fmt"
	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// ResourceCollector 系统资源指标收集器
// 负责收集操作系统级别的资源使用指标,如CPU、内存、磁盘等
type ResourceCollector struct {
	// 系统CPU使用率
	cpuUsage *metric.Gauge

	// 系统内存使用
	memUsage *metric.Gauge
	memTotal *metric.Gauge
	memFree  *metric.Gauge

	// 系统磁盘使用
	diskUsage *metric.Gauge
	diskTotal *metric.Gauge
	diskFree  *metric.Gauge
}

// NewResourceCollector 创建资源收集器
func NewResourceCollector(namespace string) *ResourceCollector {
	c := &ResourceCollector{}

	// CPU指标
	c.cpuUsage = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_cpu_usage_percent",
		Help:      "System CPU usage percentage",
	}).WithLabels([]string{"cpu"})

	// 内存指标
	c.memUsage = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_usage_bytes",
		Help:      "System memory usage in bytes",
	})

	c.memTotal = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_total_bytes",
		Help:      "System total memory in bytes",
	})

	c.memFree = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_memory_free_bytes",
		Help:      "System free memory in bytes",
	})

	// 磁盘指标
	c.diskUsage = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_usage_bytes",
		Help:      "System disk usage in bytes",
	}).WithLabels([]string{"path"})

	c.diskTotal = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_total_bytes",
		Help:      "System total disk space in bytes",
	}).WithLabels([]string{"path"})

	c.diskFree = metric.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_disk_free_bytes",
		Help:      "System free disk space in bytes",
	}).WithLabels([]string{"path"})

	return c
}

// Collect 收集系统资源指标
func (c *ResourceCollector) Collect() error {
	// 收集CPU使用率
	cpuPercents, err := cpu.Percent(0, true)
	if err == nil {
		for i, percent := range cpuPercents {
			c.cpuUsage.WithLabelValues(fmt.Sprintf("cpu%d", i)).Set(percent)
		}
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

// Register 注册资源收集器
func (c *ResourceCollector) Register() error {
	collectors := []interface{}{
		c.cpuUsage,
		c.memUsage,
		c.memTotal,
		c.memFree,
		c.diskUsage,
		c.diskTotal,
		c.diskFree,
	}

	for _, collector := range collectors {
		if err := collector.(interface{ Register() error }).Register(); err != nil {
			return err
		}
	}
	return nil
}
