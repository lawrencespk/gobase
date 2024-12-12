package metric_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestHistogram(t *testing.T) {
	t.Run("创建和注册基础直方图", func(t *testing.T) {
		// Arrange
		opts := prometheus.HistogramOpts{
			Name:    "test_histogram",
			Help:    "Test histogram help",
			Buckets: prometheus.DefBuckets,
		}

		// Act
		histogram := metric.NewHistogram(opts)
		err := histogram.Register()

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(histogram.GetCollector())
	})

	t.Run("创建和注册带标签的直方图", func(t *testing.T) {
		// Arrange
		opts := prometheus.HistogramOpts{
			Name:    "test_histogram_with_labels",
			Help:    "Test histogram with labels help",
			Buckets: prometheus.DefBuckets,
		}

		// Act
		histogram := metric.NewHistogram(opts).WithLabels("label1", "label2")
		err := histogram.Register()

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(histogram.GetCollector())
	})

	t.Run("直方图观察值", func(t *testing.T) {
		// Arrange
		opts := prometheus.HistogramOpts{
			Name:    "test_histogram_observe",
			Help:    "Test histogram observe help",
			Buckets: prometheus.DefBuckets,
		}
		histogram := metric.NewHistogram(opts)
		histogram.Register()

		// Act & Assert
		assert.NotPanics(t, func() {
			histogram.Observe(0.5)
			histogram.Observe(1.0)
			histogram.Observe(2.0)
		})

		// Cleanup
		prometheus.Unregister(histogram.GetCollector())
	})

	t.Run("带标签的直方图观察值", func(t *testing.T) {
		// Arrange
		opts := prometheus.HistogramOpts{
			Name:    "test_histogram_labels_observe",
			Help:    "Test histogram with labels observe help",
			Buckets: prometheus.DefBuckets,
		}
		histogram := metric.NewHistogram(opts).WithLabels("method", "status")
		histogram.Register()

		// Act & Assert
		assert.NotPanics(t, func() {
			histogram.WithLabelValues("GET", "200").Observe(0.5)
			histogram.WithLabelValues("POST", "500").Observe(1.5)
		})

		// Cleanup
		prometheus.Unregister(histogram.GetCollector())
	})

	t.Run("重复注册直方图", func(t *testing.T) {
		// Arrange
		opts := prometheus.HistogramOpts{
			Name:    "test_histogram_duplicate",
			Help:    "Test histogram duplicate registration help",
			Buckets: prometheus.DefBuckets,
		}
		histogram1 := metric.NewHistogram(opts)
		histogram2 := metric.NewHistogram(opts)

		// Act
		err1 := histogram1.Register()
		err2 := histogram2.Register()

		// Assert
		assert.NoError(t, err1)
		assert.Error(t, err2) // 重复注册应该返回错误

		// Cleanup
		prometheus.Unregister(histogram1.GetCollector())
	})
}
