package collector_test

import (
	"strings"
	"testing"
	"time"

	"gobase/pkg/monitor/prometheus/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSystemCollector(t *testing.T) {
	t.Run("创建和注册系统指标收集器", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()

		// Act & Assert
		assert.NotNil(t, sc, "系统指标收集器不应为空")
		assert.NoError(t, sc.Register(), "注册系统指标收集器应成功")
	})

	t.Run("系统指标描述符", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()
		ch := make(chan *prometheus.Desc, 100)

		// Act
		sc.Describe(ch)
		close(ch)

		// Assert
		var descs []*prometheus.Desc
		for desc := range ch {
			descs = append(descs, desc)
		}
		assert.Greater(t, len(descs), 0, "应该至少有一个指标描述符")
	})

	t.Run("系统指标收集", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()
		ch := make(chan prometheus.Metric, 100)
		done := make(chan struct{})

		// Act
		go func() {
			sc.Collect(ch)
			close(ch)
			close(done)
		}()

		// Assert
		select {
		case <-done:
			var metrics []prometheus.Metric
			for metric := range ch {
				metrics = append(metrics, metric)
			}
			assert.Greater(t, len(metrics), 0, "应该至少有一个指标")
		case <-time.After(5 * time.Second):
			t.Fatal("测试超时")
		}
	})

	t.Run("系统负载指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()
		metricCh := make(chan prometheus.Metric, 100)
		done := make(chan struct{})
		var metrics []prometheus.Metric

		// Act
		go func() {
			sc.Collect(metricCh)
			close(metricCh)
			close(done)
		}()

		// Assert
		select {
		case <-done:
			for metric := range metricCh {
				metrics = append(metrics, metric)
			}
			var hasLoadMetric bool
			for _, metric := range metrics {
				if strings.Contains(metric.Desc().String(), "system_load_average") {
					hasLoadMetric = true
					break
				}
			}
			assert.True(t, hasLoadMetric, "应该包含系统负载指标")
		case <-time.After(5 * time.Second):
			t.Fatal("测试超时")
		}
	})

	t.Run("系统文件描述符指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()
		metricCh := make(chan prometheus.Metric, 100)
		done := make(chan struct{})
		var metrics []prometheus.Metric

		// Act
		go func() {
			sc.Collect(metricCh)
			close(metricCh)
			close(done)
		}()

		// Assert
		select {
		case <-done:
			for metric := range metricCh {
				metrics = append(metrics, metric)
			}
			var hasFDMetric bool
			for _, metric := range metrics {
				if strings.Contains(metric.Desc().String(), "system_open_fds") {
					hasFDMetric = true
					break
				}
			}
			assert.True(t, hasFDMetric, "应该包含文件描述符指标")
		case <-time.After(5 * time.Second):
			t.Fatal("测试超时")
		}
	})

	t.Run("系统网络连接指标", func(t *testing.T) {
		// Arrange
		sc := collector.NewSystemCollector()
		metricCh := make(chan prometheus.Metric, 100)
		done := make(chan struct{})
		var metrics []prometheus.Metric

		// Act
		go func() {
			sc.Collect(metricCh)
			close(metricCh)
			close(done)
		}()

		// Assert
		select {
		case <-done:
			for metric := range metricCh {
				metrics = append(metrics, metric)
			}
			var hasNetConnMetric bool
			for _, metric := range metrics {
				if strings.Contains(metric.Desc().String(), "system_net_connections") {
					hasNetConnMetric = true
					break
				}
			}
			assert.True(t, hasNetConnMetric, "应该包含网络连接指标")
		case <-time.After(5 * time.Second):
			t.Fatal("测试超时")
		}
	})
}
