package benchmark

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
)

func BenchmarkCache(b *testing.B) {
	// 启动Redis容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		b.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	// 创建Redis客户端
	client, err := redisClient.NewClient(
		redisClient.WithAddresses([]string{addr}),
		redisClient.WithDialTimeout(time.Second*5),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	// 创建缓存实例
	cache, err := redis.NewCache(redis.Options{
		Client: client,
	})
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := cache.Set(ctx, "bench_key", "value", time.Minute)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预先设置一个值
		err := cache.Set(ctx, "bench_key", "value", time.Minute)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cache.Get(ctx, "bench_key")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 每次先设置再删除
			err := cache.Set(ctx, "bench_key", "value", time.Minute)
			if err != nil {
				b.Fatal(err)
			}

			err = cache.Delete(ctx, "bench_key")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
