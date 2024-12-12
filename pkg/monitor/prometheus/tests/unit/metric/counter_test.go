package metric_test

import (
	"testing"

	"gobase/pkg/monitor/prometheus/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	// 每个测试用例执行前重置注册表
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	t.Run("创建计数器", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_create",
			Help: "Test counter creation",
		}

		// Act
		counter := metric.NewCounter(opts)

		// Assert
		assert.NotNil(t, counter)
		assert.NotNil(t, counter.GetCounter())
		// 验证是否为基础计数器（没有标签）
		assert.NotNil(t, counter.GetCollector())
	})

	t.Run("基础计数器操作", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_basic",
			Help: "Test basic counter operations",
		}
		counter := metric.NewCounter(opts)
		err := counter.Register()

		// Assert registration
		assert.NoError(t, err)

		// Act & Assert operations
		assert.NotPanics(t, func() {
			counter.Inc()
			counter.Add(2.5)
		})

		// Cleanup
		prometheus.Unregister(counter.GetCollector())
	})

	t.Run("带标签的计数器", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_labels",
			Help: "Test counter with labels",
		}
		labels := []string{"method", "status"}
		counter := metric.NewCounter(opts).WithLabels(labels...)

		// Act
		err := counter.Register()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, counter.GetCollector())
		assert.Nil(t, counter.GetCounter()) // 有标签时，基础计数器应为nil

		// 验证是否为向量计数器
		_, ok := counter.GetCollector().(*prometheus.CounterVec)
		assert.True(t, ok, "应该是CounterVec类型")

		// Test label operations
		assert.NotPanics(t, func() {
			counter.WithLabelValues("GET", "200").Inc()
			counter.WithLabelValues("POST", "500").Add(1.5)
		})

		// Cleanup
		prometheus.Unregister(counter.GetCollector())
	})

	t.Run("重复注册", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_duplicate",
			Help: "Test counter duplicate registration",
		}
		counter1 := metric.NewCounter(opts)
		counter2 := metric.NewCounter(opts)

		// Act & Assert
		assert.NoError(t, counter1.Register())
		assert.Error(t, counter2.Register()) // 重复注册应返回错误

		// Cleanup
		prometheus.Unregister(counter1.GetCollector())
	})

	t.Run("标签值操作", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_label_values",
			Help: "Test counter label values operations",
		}
		counter := metric.NewCounter(opts).WithLabels("method")
		err := counter.Register()
		assert.NoError(t, err)

		// Act & Assert
		assert.NotPanics(t, func() {
			// 正确的标签数量
			counter.WithLabelValues("GET").Inc()

			// 错误的标签数量应该panic
			assert.Panics(t, func() {
				counter.WithLabelValues("GET", "extra").Inc()
			})
		})

		// Cleanup
		prometheus.Unregister(counter.GetCollector())
	})

	t.Run("Collector接口实现", func(t *testing.T) {
		// Arrange
		opts := prometheus.CounterOpts{
			Name: "test_counter_collector",
			Help: "Test counter collector interface",
		}
		counter := metric.NewCounter(opts)

		// Test Describe
		descCh := make(chan *prometheus.Desc, 1)
		counter.Describe(descCh)
		desc := <-descCh
		assert.NotNil(t, desc)

		// Test Collect
		metricCh := make(chan prometheus.Metric, 1)
		counter.Collect(metricCh)
		metric := <-metricCh
		assert.NotNil(t, metric)

		close(descCh)
		close(metricCh)
	})
}
