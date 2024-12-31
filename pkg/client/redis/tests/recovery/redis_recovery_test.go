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
			redis.WithMaxRetries(3),
			redis.WithRetryBackoff(time.Second),
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

		// 验证自动重连和恢复
		time.Sleep(5 * time.Second)
		if _, err := client.Get(ctx, "test_key"); err != nil {
			t.Fatal(err)
		}
	})
}
