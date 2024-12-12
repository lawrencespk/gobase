package metric_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	// 每个测试用例执行前重置注册表
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	t.Run("创建和注册基础仪表盘", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge",
			Help: "Test gauge help",
		}

		// Act
		gauge := metric.NewGauge(opts)

		// Assert
		assert.NotNil(t, gauge)
		assert.NotNil(t, gauge.GetGauge())
		assert.NotNil(t, gauge.GetCollector())

		// Test registration
		err := gauge.Register()
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(gauge.GetCollector())
	})

	t.Run("创建和注册带标签的仪表盘", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge_with_labels",
			Help: "Test gauge with labels help",
		}
		labels := []string{"label1", "label2"}

		// Act
		gauge := metric.NewGauge(opts).WithLabels(labels)

		// Assert
		assert.NotNil(t, gauge)
		assert.Nil(t, gauge.GetGauge()) // 有标签时，基础仪表盘应为nil
		assert.NotNil(t, gauge.GetCollector())

		// Test registration
		err := gauge.Register()
		assert.NoError(t, err)

		// Verify it's a GaugeVec
		_, ok := gauge.GetCollector().(*prometheus.GaugeVec)
		assert.True(t, ok, "应该是GaugeVec类型")

		// Cleanup
		prometheus.Unregister(gauge.GetCollector())
	})

	t.Run("仪表盘值操作", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge_ops",
			Help: "Test gauge operations help",
		}
		gauge := metric.NewGauge(opts)
		err := gauge.Register()
		assert.NoError(t, err)

		// Act & Assert
		assert.NotPanics(t, func() {
			gauge.Set(10)
			gauge.Inc()
			gauge.Dec()
			gauge.Add(5)
			gauge.Sub(3)
		})

		// Cleanup
		prometheus.Unregister(gauge.GetCollector())
	})

	t.Run("带标签的仪表盘操作", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge_labels_ops",
			Help: "Test gauge with labels operations help",
		}
		gauge := metric.NewGauge(opts).WithLabels([]string{"service", "instance"})
		err := gauge.Register()
		assert.NoError(t, err)

		// Act & Assert
		assert.NotPanics(t, func() {
			gauge.WithLabelValues("auth", "pod-1").Set(10)
			gauge.WithLabelValues("auth", "pod-1").Inc()
			gauge.WithLabelValues("auth", "pod-2").Add(5)
		})

		// Test wrong label count
		assert.Panics(t, func() {
			gauge.WithLabelValues("auth").Inc() // 标签数量不匹配应该panic
		})

		// Cleanup
		prometheus.Unregister(gauge.GetCollector())
	})

	t.Run("重复注册", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge_duplicate",
			Help: "Test gauge duplicate registration",
		}
		gauge1 := metric.NewGauge(opts)
		gauge2 := metric.NewGauge(opts)

		// Act & Assert
		assert.NoError(t, gauge1.Register())
		assert.Error(t, gauge2.Register()) // 重复注册应返回错误

		// Cleanup
		prometheus.Unregister(gauge1.GetCollector())
	})
}
