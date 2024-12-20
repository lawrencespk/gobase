package integration

import (
	"context"
	"errors"
	"gobase/pkg/monitor/prometheus/exporter"
	"gobase/pkg/monitor/prometheus/tests/testutils"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusinessMetrics(t *testing.T) {
	suite := setupTestSuite(t)
	defer tearDownTestSuite(t, suite)

	// 设置指标命名空间
	if suite.cfg.Labels == nil {
		suite.cfg.Labels = make(map[string]string)
	}
	suite.cfg.Labels["app"] = "testapp"

	// 确保启用了所需的收集器
	suite.cfg.Collectors = []string{"business"}

	// 设置合理的超时时间
	ctx, cancel := context.WithTimeout(suite.ctx, 30*time.Second)
	defer cancel()

	exp, err := exporter.New(suite.cfg, suite.logger)
	require.NoError(t, err, "创建导出器失败")

	// 启动导出器
	errChan := make(chan error, 1)
	go func() {
		errChan <- exp.Start(ctx)
	}()

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 获取业务收集器
	businessCollector := exp.GetBusinessCollector()
	require.NotNil(t, businessCollector, "获取业务指标收集器失败")

	t.Run("业务操作指标", func(t *testing.T) {
		// 模拟多次业务操作
		for i := 0; i < 5; i++ {
			businessCollector.ObserveOperation("create_user", float64(i)*0.1, nil)
			time.Sleep(100 * time.Millisecond)
		}
		businessCollector.ObserveOperation("create_user", 0.7, errors.New("test error"))

		// 等待指标收集
		time.Sleep(2 * time.Second)

		// 验证操作计数
		metrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "testapp_business_operations_total")
		require.NoError(t, err, "查询操作计数指标失败")
		require.NotEmpty(t, metrics, "操作计数指标为空")

		// 验证错误计数
		errorMetrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "testapp_business_operation_errors_total")
		require.NoError(t, err, "查询错误计数指标失败")
		require.NotEmpty(t, errorMetrics, "错误计数指标为空")

		// 验证延迟指标 - 使用直方图查询语法
		durationMetrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI,
			"rate(testapp_business_operation_duration_seconds_sum[1m])")
		require.NoError(t, err, "查询延迟指标失败")
		require.NotEmpty(t, durationMetrics, "延迟指标为空")
	})

	t.Run("队列指标", func(t *testing.T) {
		// 多次更新队列大小
		for i := 0; i < 5; i++ {
			businessCollector.SetQueueSize("user_queue", float64(i*20))
			time.Sleep(100 * time.Millisecond)
		}
		businessCollector.SetQueueSize("user_queue", 100)

		// 多次更新处理速率
		for i := 0; i < 5; i++ {
			businessCollector.SetProcessRate("create_user", float64(i*10))
			time.Sleep(100 * time.Millisecond)
		}
		businessCollector.SetProcessRate("create_user", 50)

		// 等待指标收集
		time.Sleep(3 * time.Second)

		// 验证队列大小
		queueMetrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "testapp_business_queue_size")
		require.NoError(t, err, "查询队列大小指标失败")
		require.NotEmpty(t, queueMetrics, "队列大小指标为空")

		if len(queueMetrics) > 0 {
			assert.Equal(t, 100.0, queueMetrics[0].Value, "队列大小值不匹配")
		}

		// 验证处理速率
		rateMetrics, err := testutils.QueryPrometheusMetrics(suite.prometheus.URI, "testapp_business_process_rate")
		require.NoError(t, err, "查询处理速率指标失败")
		require.NotEmpty(t, rateMetrics, "处理速率指标为空")

		if len(rateMetrics) > 0 {
			assert.Equal(t, 50.0, rateMetrics[0].Value, "处理速率值不匹配")
		}
	})

	// 优雅关闭服务器
	cancel()
	select {
	case err := <-errChan:
		if err != nil && !strings.Contains(err.Error(), "http: Server closed") {
			require.NoError(t, err, "服务器关闭出错")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("服务器关闭超时")
	}
}

func tearDownTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	// 取消上下文
	if suite.cancel != nil {
		suite.cancel()
	}

	// 等待一段时间让服务器优雅关闭
	time.Sleep(2 * time.Second)

	// 终止Prometheus容器
	if suite.prometheus != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := suite.prometheus.Terminate(ctx); err != nil {
			t.Logf("Warning: failed to terminate Prometheus container: %v", err)
		}
	}
}
