package integration

import (
	"gobase/pkg/monitor/prometheus/exporter"
	"gobase/pkg/monitor/prometheus/tests/testutils"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSystemMetrics(t *testing.T) {
	suite := setupTestSuite(t)
	defer tearDownTestSuite(t, suite)

	suite.cfg.Collectors = append(suite.cfg.Collectors, "system")

	exp, err := exporter.New(suite.cfg, suite.logger)
	assert.NoError(t, err)

	err = exp.Start(suite.ctx)
	assert.NoError(t, err)

	if runtime.GOOS == "windows" {
		t.Log("在 Windows 系统上等待防火墙授权...")
		time.Sleep(5 * time.Second)
	} else {
		time.Sleep(100 * time.Millisecond)
	}

	// 增加等待时间，确保有足够时间收集和暴露指标
	time.Sleep(2 * time.Second)

	t.Run("CPU指标", func(t *testing.T) {
		metrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "system_cpu_usage_percent")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics, "应该有CPU使用率指标数据")

		if len(metrics) > 0 {
			assert.Greater(t, metrics[0].Value, 0.0, "CPU使用率应该大于0")
		}
	})

	t.Run("内存指标", func(t *testing.T) {
		metrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "system_memory_usage_bytes")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics, "应该有内存使用指标数据")

		if len(metrics) > 0 {
			assert.Greater(t, metrics[0].Value, 0.0, "内存使用量应该大于0")
		}
	})
}
