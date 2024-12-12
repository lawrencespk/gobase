package collector_test

import (
	"strings"
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestResourceCollector(t *testing.T) {
	t.Run("创建和注册资源指标收集器", func(t *testing.T) {
		// Arrange
		rc := collector.NewResourceCollector("test")

		// Act
		err := prometheus.Register(rc)

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(rc)
	})

	t.Run("资源指标描述符", func(t *testing.T) {
		// Arrange
		rc := collector.NewResourceCollector("test")
		ch := make(chan *prometheus.Desc, 10)

		// Act
		rc.Describe(ch)
		close(ch)

		// Assert
		descCount := 0
		for range ch {
			descCount++
		}
		assert.Greater(t, descCount, 0, "应该至少有一个指标描述符")
	})

	t.Run("资源指标收集", func(t *testing.T) {
		// Arrange
		rc := collector.NewResourceCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		rc.Collect(ch)
		close(ch)

		// Assert
		metricCount := 0
		for range ch {
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})

	t.Run("CPU使用率指标", func(t *testing.T) {
		// Arrange
		rc := collector.NewResourceCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		rc.Collect(ch)
		close(ch)

		// Assert
		var hasCPUMetric bool
		for metric := range ch {
			desc := metric.Desc()
			if strings.Contains(desc.String(), "test_system_cpu_usage_percent") {
				hasCPUMetric = true
				break
			}
		}
		assert.True(t, hasCPUMetric, "应该包含CPU使用率指标")
	})

	t.Run("内存使用指标", func(t *testing.T) {
		// Arrange
		rc := collector.NewResourceCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		rc.Collect(ch)
		close(ch)

		// Assert
		var hasMemoryMetric bool
		for metric := range ch {
			desc := metric.Desc()
			if desc.String() == "Desc{fqName: \"test_system_memory_usage_bytes\", help: \"System memory usage in bytes\", constLabels: {}, variableLabels: []}" {
				hasMemoryMetric = true
				break
			}
		}
		assert.True(t, hasMemoryMetric, "应该包含内存使用指标")
	})
}
