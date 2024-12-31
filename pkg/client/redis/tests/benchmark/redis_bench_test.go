package benchmark

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"testing"
	"time"
)

func BenchmarkRedisOperations(b *testing.B) {
	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		b.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	// 创建 Redis 客户端
	client, err := redis.NewClient(
		redis.WithAddress(addr),
		redis.WithPoolSize(10),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// 预热
	for i := 0; i < 100; i++ {
		if err := client.Set(ctx, "warmup_key", "value", time.Minute); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()                       // 重置计时器
		b.RunParallel(func(pb *testing.PB) { // 使用并行测试
			for pb.Next() {
				err := client.Set(ctx, "bench_key", "value", time.Minute)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Get", func(b *testing.B) {
		// 确保有数据可以获取
		if err := client.Set(ctx, "bench_key", "value", time.Minute); err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := client.Get(ctx, "bench_key")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Pipeline", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				pipe := client.TxPipeline()
				pipe.Set(ctx, "pipe_key", "value", time.Minute)
				pipe.Get(ctx, "pipe_key")
				_, err := pipe.Exec(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
