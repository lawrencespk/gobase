package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/store"
	"gobase/pkg/client/redis"
	"gobase/pkg/logger"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func BenchmarkStores(b *testing.B) {
	ctx := context.Background()

	// 初始化 Redis client
	client, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithDB(0),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	// 创建 logger
	log, err := logger.NewLogger(
		logger.WithLevel(types.InfoLevel),
		logger.WithOutputPaths([]string{"logs/benchmark_test.log"}),
		logger.WithAsyncConfig(logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1000,
			FlushInterval: time.Second,
			BlockOnFull:   false,
			DropOnFull:    true,
			FlushOnExit:   true,
		}),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer log.Sync()

	// 创建stores
	redisOpts := &store.Options{
		Redis: &store.RedisOptions{
			Addr: "localhost:6379",
		},
		EnableMetrics: true,
		EnableTracing: true,
		KeyPrefix:     "test:",
	}
	redisStore := store.NewRedisTokenStore(client, redisOpts, log)

	memOpts := store.Options{
		EnableMetrics:   true,
		EnableTracing:   true,
		KeyPrefix:       "test:",
		CleanupInterval: time.Minute,
	}
	memStore, err := store.NewMemoryStore(memOpts)
	if err != nil {
		b.Fatal(err)
	}

	// 创建 Claims
	claims := jwt.NewStandardClaims(
		jwt.WithUserID(fmt.Sprintf("user-%d", 123)),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
	)

	// 准备测试数据
	tokenInfo := &jwt.TokenInfo{
		Raw:       "bench_token",
		Type:      jwt.AccessToken,
		ExpiresAt: time.Now().Add(time.Hour),
		IsRevoked: false,
		Claims:    claims,
	}

	b.Run("Redis Store Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenInfo.Raw = fmt.Sprintf("token_%d", i)
			_ = redisStore.Set(ctx, tokenInfo.Raw, tokenInfo, time.Hour)
		}
	})

	b.Run("Memory Store Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenInfo.Raw = fmt.Sprintf("token_%d", i)
			_ = memStore.Set(ctx, tokenInfo.Raw, tokenInfo, time.Hour)
		}
	})

	b.Run("Redis Store Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = redisStore.Get(ctx, tokenInfo.Raw)
		}
	})

	b.Run("Memory Store Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = memStore.Get(ctx, tokenInfo.Raw)
		}
	})
}
