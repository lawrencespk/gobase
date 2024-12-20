package collector_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/collector"

	"strings"

	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestHTTPCollector(t *testing.T) {
	t.Run("指标注册和收集", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")

		// 先注册收集器
		err := hc.Register()
		assert.NoError(t, err)

		// 添加测试数据
		hc.ObserveRequest("GET", "/test", 200, 100*time.Millisecond, 1000, 2000)

		// 验证指标收集
		ch := make(chan prometheus.Metric, 10)
		hc.Collect(ch)
		close(ch)

		// 验证指标
		metrics := make([]prometheus.Metric, 0)
		for m := range ch {
			metrics = append(metrics, m)
		}

		// 验证是否包含所有必要的指标
		assert.Greater(t, len(metrics), 0)

		// 清理
		prometheus.Unregister(hc)
	})

	t.Run("HTTP指标描述符", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		ch := make(chan *prometheus.Desc, 10)

		// Act
		hc.Describe(ch)
		close(ch)

		// Assert
		descCount := 0
		for desc := range ch {
			assert.NotNil(t, desc)
			descCount++
		}
		assert.Equal(t, 6, descCount, "应该有6个指标描述符")
	})

	t.Run("HTTP指标收集", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// 添加一些测试数据
		hc.ObserveRequest("GET", "/test", 200, 100, 1000, 2000)
		hc.IncActiveRequests("GET")
		hc.ObserveSlowRequest("GET", "/test", 5000)

		// Act
		hc.Collect(ch)
		close(ch)

		// Assert
		metricCount := 0
		for metric := range ch {
			assert.NotNil(t, metric)
			metricCount++
		}
		assert.Greater(t, metricCount, 0, "应该至少有一个指标")
	})

	t.Run("HTTP指标标签验证", func(t *testing.T) {
		// Arrange
		hc := collector.NewHTTPCollector("test")
		ch := make(chan prometheus.Metric, 10)

		// 添加测试数据
		hc.ObserveRequest("GET", "/test", 200, 100, 1000, 2000)

		// Act
		hc.Collect(ch)
		close(ch)

		// Assert
		var foundRequestMetric bool
		for metric := range ch {
			desc := metric.Desc().String()
			if strings.Contains(desc, "test_http_requests_total") {
				foundRequestMetric = true
				break
			}
		}
		assert.True(t, foundRequestMetric, "应该包含请求总数指标")
	})
}
