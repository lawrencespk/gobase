package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// 禁用连接检查
	redis.DisableConnectionCheck = true
	defer func() {
		redis.DisableConnectionCheck = false
	}()

	// 创建一个新的 registry
	registry := prometheus.NewRegistry()

	metricsPrefix := "test_metrics_" + time.Now().Format("150405")
	client, err := redis.NewClient(
		redis.WithMetrics(true),
		redis.WithMetricsNamespace(metricsPrefix),
		redis.WithRegistry(registry),
		// 使用一个不存在的地址，因为我们已经禁用了连接检查
		redis.WithAddress("localhost:0"),
	)

	assert.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// 执行一些操作
	err = client.Set(context.Background(), "test_key", "test_value", time.Second)
	// 由于没有实际的 Redis 连接，我们期望这里会返回错误
	assert.Error(t, err)

	// 获取并验证指标
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// 验证指标名称前缀
	for _, m := range metrics {
		assert.Contains(t, *m.Name, metricsPrefix)
	}
}
