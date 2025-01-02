package recovery

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"testing"
	"time"
)

func TestRedisRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping recovery test in short mode")
	}

	t.Run("connection_recovery", func(t *testing.T) {
		// 启动 Redis 容器
		addr, err := testutils.StartRedisSingleContainer()
		if err != nil {
			t.Fatal(err)
		}
		defer testutils.CleanupRedisContainers()

		client, err := redis.NewClient(
			redis.WithAddress(addr),
			redis.WithMaxRetries(5),
			redis.WithRetryBackoff(2*time.Second),
			redis.WithDialTimeout(3*time.Second),
			redis.WithReadTimeout(2*time.Second),
			redis.WithWriteTimeout(2*time.Second),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		ctx := context.Background()

		// 测试正常操作
		if err := client.Set(ctx, "test_key", "value", time.Minute); err != nil {
			t.Fatal(err)
		}

		// 重启 Redis 容器
		if err := testutils.RestartRedisContainer(); err != nil {
			t.Fatal(err)
		}

		// 等待 Redis 完全重启
		time.Sleep(10 * time.Second)

		// 先尝试 Ping 来触发重连
		for i := 0; i < 3; i++ {
			if err := client.Ping(ctx); err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}

		// 验证自动重连和恢复
		if _, err := client.Get(ctx, "test_key"); err != nil {
			t.Fatal(err)
		}
	})
}
