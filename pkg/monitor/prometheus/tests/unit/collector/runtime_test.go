package collector_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeCollector(t *testing.T) {
	t.Run("创建和注册运行时指标收集器", func(t *testing.T) {
		// Arrange
		rc := collector.NewRuntimeCollector("test")

		// Act
		err := prometheus.Register(rc)

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(rc)
	})

	t.Run("运行时指标描述符", func(t *testing.T) {
		// Arrange
		rc := collector.NewRuntimeCollector("test")
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

	t.Run("运行时指标收集", func(t *testing.T) {
		// Arrange
		rc := collector.NewRuntimeCollector("test")
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

	t.Run("Goroutine数量指标", func(t *testing.T) {
		// Arrange
		rc := collector.NewRuntimeCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		rc.Collect(ch)
		close(ch)

		// Assert
		var hasGoroutineMetric bool
		for metric := range ch {
			desc := metric.Desc()
			if desc.String() == "Desc{fqName: \"test_runtime_goroutines_total\", help: \"Total number of goroutines\", constLabels: {}, variableLabels: []}" {
				hasGoroutineMetric = true
				break
			}
		}
		assert.True(t, hasGoroutineMetric, "应该包含Goroutine数量指标")
	})

	t.Run("内存使用指标", func(t *testing.T) {
		// Arrange
		rc := collector.NewRuntimeCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		rc.Collect(ch)
		close(ch)

		// Assert
		var hasMemoryMetric bool
		for metric := range ch {
			desc := metric.Desc()
			if desc.String() == "Desc{fqName: \"test_runtime_memory_alloc_bytes\", help: \"Runtime allocated memory in bytes\", constLabels: {}, variableLabels: []}" {
				hasMemoryMetric = true
				break
			}
		}
		assert.True(t, hasMemoryMetric, "应该包含内存使用指标")
	})
}
