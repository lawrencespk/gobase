package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func BenchmarkMemoryStore(b *testing.B) {
	ctx := context.Background()
	store := blacklist.NewMemoryStore()
	defer store.Close()

	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Hour

	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = store.Add(ctx, tokenID, reason, expiration)
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = store.Get(ctx, tokenID)
		}
	})

	b.Run("Remove", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = store.Remove(ctx, tokenID)
		}
	})
}

func BenchmarkRealRedisStore(b *testing.B) {
	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	require.NoError(b, err)
	defer testutils.CleanupRedisContainers()

	// 创建真实的 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress(addr),
	)
	require.NoError(b, err)
	defer redisClient.Close()

	// 创建日志记录器，设置为 Error 级别避免性能影响
	log, _ := logger.NewLogger(logger.WithLevel(types.ErrorLevel))

	// 创建 Redis Store
	store, err := blacklist.NewRedisStore(redisClient, &blacklist.Options{
		DefaultExpiration: time.Hour,
		CleanupInterval:   time.Minute,
		Logger:            log,
		EnableMetrics:     false, // 基准测试禁用指标收集
	})
	require.NoError(b, err)
	defer store.Close()

	ctx := context.Background()
	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Hour

	b.Run("Add", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = store.Add(ctx, tokenID+string(rune(i)), reason, expiration)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 先添加一个测试数据
		err := store.Add(ctx, tokenID, reason, expiration)
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.Get(ctx, tokenID)
		}
	})

	b.Run("Remove", func(b *testing.B) {
		// 先添加一批测试数据
		for i := 0; i < b.N; i++ {
			_ = store.Add(ctx, tokenID+string(rune(i)), reason, expiration)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = store.Remove(ctx, tokenID+string(rune(i)))
		}
	})
}
