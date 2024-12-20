package integration

import (
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/monitor/prometheus/exporter"
	"gobase/pkg/monitor/prometheus/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSystemMetrics(t *testing.T) {
	suite := setupTestSuite(t)
	defer tearDownTestSuite(t, suite)

	exp, err := exporter.New(suite.cfg, suite.logger)
	assert.NoError(t, err)

	err = exp.Start(suite.ctx)
	assert.NoError(t, err)

	collector := collector.NewSystemCollector()
	defer collector.Stop()

	// 等待收集器有时间收集数据
	time.Sleep(2 * time.Second)

	t.Run("CPU指标", func(t *testing.T) {
		metrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "system_cpu_usage_percent")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics)
		assert.Greater(t, metrics[0].Value, 0.0)
	})

	t.Run("内存指标", func(t *testing.T) {
		time.Sleep(2 * time.Second)

		metrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "system_memory_usage_bytes")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics)
		assert.Greater(t, metrics[0].Value, 0.0)
	})
}
