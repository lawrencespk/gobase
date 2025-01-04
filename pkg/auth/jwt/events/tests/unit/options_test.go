package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt/events"
	"gobase/pkg/monitor/prometheus/metric"
)

func TestPublisherOptions(t *testing.T) {
	t.Run("WithPublisherMetrics", func(t *testing.T) {
		// 准备
		metrics := metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Subsystem: "events",
			Name:      "test_counter",
			Help:      "Test counter for JWT events",
		})
		// 添加标签支持
		metrics.WithLabels("event_type")

		pub := &events.Publisher{}
		opt := events.WithMetrics(metrics)

		// 执行
		opt(pub)

		// 验证
		assert.NotNil(t, pub)
	})

	t.Run("WithChannel", func(t *testing.T) {
		// 准备
		pub := &events.Publisher{}
		channel := "test_channel"
		opt := events.WithChannel(channel)

		// 执行
		opt(pub)

		// 验证
		assert.NotNil(t, pub)
	})
}

func TestSubscriberOptions(t *testing.T) {
	t.Run("WithSubscriberMetrics", func(t *testing.T) {
		// 准备
		metrics := metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Subsystem: "events",
			Name:      "test_counter",
			Help:      "Test counter for JWT events",
		})
		// 添加标签支持
		metrics.WithLabels("event_type")

		sub := &events.Subscriber{}
		opt := events.WithSubscriberMetrics(metrics)

		// 执行
		opt(sub)

		// 验证
		assert.NotNil(t, sub)
	})
}
