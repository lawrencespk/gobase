package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/ratelimit"
	"gobase/pkg/cache/redis/ratelimit/tests/testutils"
	redisclient "gobase/pkg/client/redis"
)

var (
	redisContainer *testutils.RedisContainer
	redisClient    redisclient.Client
	redisStore     *ratelimit.Store
)

func setupBenchmark(b *testing.B) {
	var err error
	redisContainer, err = testutils.StartRedisContainer(b)
	require.NoError(b, err)

	redisAddr := redisContainer.GetAddress()

	// 创建 Redis 客户端
	client, err := redisclient.NewClient(
		redisclient.WithAddresses([]string{redisAddr}),
		redisclient.WithDialTimeout(time.Second),
		redisclient.WithReadTimeout(time.Second),
		redisclient.WithWriteTimeout(time.Second),
		redisclient.WithPoolSize(50),
	)
	require.NoError(b, err)
	redisClient = client

	// 确保 Redis 连接正常
	err = redisClient.Ping(context.Background())
	require.NoError(b, err)

	redisStore = ratelimit.NewStore(redisClient)
}

func teardownBenchmark(b *testing.B) {
	if redisClient != nil {
		redisClient.Close()
	}
	if redisContainer != nil {
		redisContainer.Terminate(b)
	}
}

func BenchmarkStore_Eval(b *testing.B) {
	setupBenchmark(b)
	defer teardownBenchmark(b)

	ctx := context.Background()
	script := `return KEYS[1]`
	keys := []string{"bench-key"}
	args := []interface{}{1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := redisStore.Eval(ctx, script, keys, args...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStore_Del(b *testing.B) {
	setupBenchmark(b)
	defer teardownBenchmark(b)

	ctx := context.Background()

	// 预先设置一些键用于删除测试
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		err := redisClient.Set(ctx, key, "value", time.Hour)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		err := redisStore.Del(ctx, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStore_Parallel(b *testing.B) {
	setupBenchmark(b)
	defer teardownBenchmark(b)

	ctx := context.Background()
	script := `return KEYS[1]`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("bench-parallel-key-%d", b.N)
			_, err := redisStore.Eval(ctx, script, []string{key}, 1)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
