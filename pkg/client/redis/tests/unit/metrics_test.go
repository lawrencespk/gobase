package unit

import (
	"testing"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"

	"github.com/stretchr/testify/assert"
)

func TestRedisMetrics(t *testing.T) {
	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	// 创建一个测试用的 namespace
	namespace := "test_redis"

	// 创建一个新的 RedisMetrics
	metrics := redis.NewRedisMetrics(namespace)
	assert.NotNil(t, metrics, "RedisMetrics should not be nil")

	// 测试创建客户端时的指标配置
	client, err := redis.NewClient(
		redis.WithAddresses([]string{addr}),
		redis.WithEnableMetrics(true),
		redis.WithMetricsNamespace(namespace),
		redis.WithCollector(metrics),
	)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	defer client.Close()

	// 测试指标记录
	metrics.ObserveCommandExecution(0.1, nil)            // 成功的命令
	metrics.ObserveCommandExecution(0.2, assert.AnError) // 失败的命令

	// 测试连接池指标
	stats := &redis.PoolStats{
		ActiveCount:  10,
		IdleCount:    5,
		TotalCount:   15,
		WaitCount:    2,
		TimeoutCount: 1,
		HitCount:     100,
		MissCount:    10,
	}
	metrics.UpdatePoolStats(stats)

	// 等待一小段时间以确保指标被收集
	time.Sleep(100 * time.Millisecond)

	// 这里可以添加更多的断言来验证指标值
	// 注意：在实际测试中，你可能需要使用 prometheus 的测试工具来验证指标值
}
