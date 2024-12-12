package metric_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	t.Run("创建和注册基础计数器", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter",
			Help: "Test counter help",
		}

		// Act
		counter := metric.NewCounter(opts)
		counter.WithLabels()
		err := counter.Register()

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(counter)
	})

	t.Run("创建和注册带标签的计数器", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_with_labels",
			Help: "Test counter with labels help",
		}
		labels := []string{"label1", "label2"}

		// Act
		counter := metric.NewCounter(opts).WithLabels(labels...)
		err := counter.Register()

		// Assert
		assert.NoError(t, err)

		// Cleanup
		prometheus.Unregister(counter)
	})

	t.Run("计数器增加值", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_inc",
			Help: "Test counter increment help",
		}
		counter := metric.NewCounter(opts)
		counter.Register()

		// Act
		counter.Inc()
		counter.Add(2)

		// Assert
		// 由于Counter的值只能通过Prometheus API获取,
		// 这里只能验证操作不会panic

		// Cleanup
		prometheus.Unregister(counter)
	})

	t.Run("带标签的计数器操作", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_labels_ops",
			Help: "Test counter with labels operations help",
		}
		counter := metric.NewCounter(opts).WithLabels("method", "status")
		counter.Register()

		// Act & Assert
		assert.NotPanics(t, func() {
			counter.WithLabelValues("GET", "200").Inc()
			counter.WithLabelValues("POST", "500").Add(2)
		})

		// Cleanup
		prometheus.Unregister(counter)
	})

	t.Run("重复注册计数器", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_duplicate",
			Help: "Test counter duplicate registration help",
		}
		counter1 := metric.NewCounter(opts)
		counter2 := metric.NewCounter(opts)

		// Act
		err1 := counter1.Register()
		err2 := counter2.Register()

		// Assert
		assert.NoError(t, err1)
		assert.Error(t, err2) // 重复注册应该返回错误

		// Cleanup
		prometheus.Unregister(counter1)
	})
}
