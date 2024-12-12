package collector_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSystemCollector(t *testing.T) {
	t.Run("创建和注册系统指标收集器", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")

		// Act
		err := prometheus.Register(sc)

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(sc)
	})

	t.Run("系统指标描述符", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")
		ch := make(chan *prometheus.Desc, 10)

		// Act
		sc.Describe(ch)
		close(ch)

		// Assert
		descCount := 0
		for range ch {
			descCount++
		}
		assert.Greater(t, descCount, 0, "应该至少有一个指标描述符")
	})

	t.Run("系统指标收集", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		sc.Collect(ch)
		close(ch)

		// Assert
		metricCount := 0
		for range ch {
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})

	t.Run("系统负载指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		sc.Collect(ch)
		close(ch)

		// Assert
		var hasLoadMetric bool
		for metric := range ch {
			if metric.Desc().String() == "system_load_average" {
				hasLoadMetric = true
				break
			}
		}
		assert.True(t, hasLoadMetric, "应该包含系统负载指标")
	})

	t.Run("系统文件描述符指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		sc.Collect(ch)
		close(ch)

		// Assert
		var hasFDMetric bool
		for metric := range ch {
			if metric.Desc().String() == "system_open_fds" {
				hasFDMetric = true
				break
			}
		}
		assert.True(t, hasFDMetric, "应该包含文件描述符指标")
	})

	t.Run("系统网络连接指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		sc.Collect(ch)
		close(ch)

		// Assert
		var hasNetConnMetric bool
		for metric := range ch {
			if metric.Desc().String() == "system_net_connections" {
				hasNetConnMetric = true
				break
			}
		}
		assert.True(t, hasNetConnMetric, "应该包含网络连接指标")
	})
}
