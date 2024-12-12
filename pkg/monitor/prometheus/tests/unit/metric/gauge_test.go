package metric_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	t.Run("创建和注册基础仪表盘", func(t *testing.T) {
		// Arrange
		opts := prometheus.GaugeOpts{
			Name: "test_gauge",
			Help: "Test gauge help",
		}

		// Act
		gauge := metric.NewGauge(opts)
		err := gauge.Register()

		// Assert
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

		// Act
		gauge := metric.NewGauge(opts).WithLabels([]string{"label1", "label2"})
		err := gauge.Register()

		// Assert
		assert.NoError(t, err)

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
		gauge.Register()

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
		gauge.Register()

		// Act & Assert
		assert.NotPanics(t, func() {
			gauge.WithLabelValues("auth", "pod-1").Set(10)
			gauge.WithLabelValues("auth", "pod-1").Inc()
			gauge.WithLabelValues("auth", "pod-2").Add(5)
		})

		// Cleanup
		prometheus.Unregister(gauge.GetCollector())
	})
}
