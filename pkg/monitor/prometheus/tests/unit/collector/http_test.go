package collector_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestHTTPCollector(t *testing.T) {
	t.Run("创建和注册HTTP指标收集器", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")

		// Act
		err := prometheus.Register(hc)

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(hc)
	})

	t.Run("HTTP指标描述符", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		ch := make(chan *prometheus.Desc, 10)

		// Act
		hc.Describe(ch)

		// Assert
		descCount := 0
		for range ch {
			descCount++
		}
		assert.Greater(t, descCount, 0, "应该至少有一个指标描述符")
	})

	t.Run("HTTP指标收集", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// Act
		hc.Collect(ch)

		// Assert
		metricCount := 0
		for range ch {
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})

	t.Run("HTTP请求计数器", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		method := "GET"
		path := "/test"
		status := 200

		// Act
		hc.ObserveRequest(method, path, status, 0, 100, 200)

		// Assert
		ch := make(chan prometheus.Metric, 10)
		hc.Collect(ch)
		close(ch)

		metricCount := 0
		for range ch {
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})
}
