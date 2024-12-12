package collector_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestBusinessCollector(t *testing.T) {
	t.Run("创建和注册业务指标收集器", func(t *testing.T) {
		// Arrange
		bc := collector.NewBusinessCollector("test")

		// Act
		err := prometheus.Register(bc)

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(bc)
	})

	t.Run("业务指标描述符", func(t *testing.T) {
		// Arrange
		bc := collector.NewBusinessCollector("test")
		ch := make(chan *prometheus.Desc, 10)

		// Act
		bc.Describe(ch)
		close(ch)

		// Assert
		descCount := 0
		for range ch {
			descCount++
		}
		assert.Greater(t, descCount, 0, "应该至少有一个指标描述符")
	})

	t.Run("业务指标收集", func(t *testing.T) {
		// Arrange
		bc := collector.NewBusinessCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		bc.Collect(ch)
		close(ch)

		// Assert
		metricCount := 0
		for range ch {
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})

	t.Run("重复注册业务指标收集器", func(t *testing.T) {
		// Arrange
		bc1 := collector.NewBusinessCollector("test")
		bc2 := collector.NewBusinessCollector("test")

		// Act
		err1 := prometheus.Register(bc1)
		err2 := prometheus.Register(bc2)

		// Assert
		assert.NoError(t, err1)
		assert.Error(t, err2) // 重复注册应该返回错误

		// Cleanup
		prometheus.Unregister(bc1)
	})

	t.Run("业务指标标签验证", func(t *testing.T) {
		// Arrange
		bc := collector.NewBusinessCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		bc.Collect(ch)
		close(ch)

		// Assert
		var hasOperationLabel bool
		for metric := range ch {
			if metric.Desc().String() == "test_business_operations_total" {
				hasOperationLabel = true
				break
			}
		}
		assert.True(t, hasOperationLabel, "应该包含操作标签")
	})
}
