package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	prometheusConfig "gobase/pkg/monitor/prometheus/config/types"
	"gobase/pkg/monitor/prometheus/exporter"
	"gobase/pkg/monitor/prometheus/tests/testutils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMetrics(t *testing.T) {
	// 在每个测试开始时重置注册表
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// 使用随机端口避免冲突
	port := testutils.GetFreePort(t)

	// 创建测试配置
	cfg := &prometheusConfig.Config{
		Enabled: true,
		Port:    port,
		Path:    "/metrics",
		Labels: map[string]string{
			"app": "testapp",
		},
		Collectors: []string{"http"},
	}

	// 创建导出器
	exp, err := exporter.New(cfg, testutils.NewTestLogger(t))
	require.NoError(t, err)

	// 启动导出器
	err = exp.Start(context.Background())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, exp.Stop(context.Background()))
	}()

	// 等待服务器完全启动，增加等待时间
	time.Sleep(3 * time.Second)

	// 获取HTTP收集器
	httpCollector := exp.GetHTTPCollector()
	require.NotNil(t, httpCollector)

	t.Run("HTTP请求指标收集", func(t *testing.T) {
		// 模拟HTTP请求
		httpCollector.ObserveRequest("GET", "/test", 200, 100*time.Millisecond, 1000, 2000)

		// 增加重试次数和等待时间
		var metrics []testutils.PrometheusMetric
		var err error
		for i := 0; i < 5; i++ {
			metricName := fmt.Sprintf("%s_http_requests_total", cfg.Labels["app"])
			metrics, err = testutils.QueryPrometheusMetrics(fmt.Sprintf("http://localhost:%d", port), metricName)
			if err == nil && len(metrics) > 0 {
				break
			}
			time.Sleep(2 * time.Second)
		}
		require.NoError(t, err)
		require.NotEmpty(t, metrics, "no metrics found after retries")

		// 验证指标值
		assert.Equal(t, "GET", metrics[0].Labels["method"])
		assert.Equal(t, "/test", metrics[0].Labels["path"])
		assert.Equal(t, "200", metrics[0].Labels["status"])
	})
}

func TestHTTPMetricsIntegration(t *testing.T) {
	// 在每个测试开始时重置注册表
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// 使用随机端口避免冲突
	port := testutils.GetFreePort(t)

	// 创建测试配置
	cfg := &prometheusConfig.Config{
		Enabled: true,
		Port:    port,
		Path:    "/metrics",
		Labels: map[string]string{
			"app": "test",
		},
		Collectors: []string{"http"},
	}

	// 创建并启动导出器
	exp, err := exporter.New(cfg, testutils.NewTestLogger(t))
	require.NoError(t, err)

	err = exp.Start(context.Background())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, exp.Stop(context.Background()))
	}()

	// 等待服务器完全启动，增加等待时间
	time.Sleep(3 * time.Second)

	// 获取HTTP收集器
	httpCollector := exp.GetHTTPCollector()
	require.NotNil(t, httpCollector)

	// 模拟HTTP请求
	httpCollector.ObserveRequest("GET", "/test", 200, 100*time.Millisecond, 1000, 2000)

	// 增加重试次数和等待时间
	var metrics []testutils.PrometheusMetric
	for i := 0; i < 5; i++ {
		metricName := fmt.Sprintf("%s_http_requests_total", cfg.Labels["app"])
		metrics, err = testutils.QueryPrometheusMetrics(fmt.Sprintf("http://localhost:%d", port), metricName)
		if err == nil && len(metrics) > 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err)
	require.NotEmpty(t, metrics)

	// 验证指标值
	assert.Equal(t, 1.0, metrics[0].Value)
}
