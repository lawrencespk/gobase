package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/auth/jwt/session/tests/testutils"
	"gobase/pkg/client/redis"

	"github.com/stretchr/testify/require"
)

func BenchmarkSessionOperations(b *testing.B) {
	ctx := context.Background()

	// 启动Redis容器
	redisAddr, err := testutils.StartRedisContainer(ctx)
	require.NoError(b, err)
	defer testutils.StopRedisContainer()

	// 初始化Redis客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress(redisAddr),
		redis.WithEnableMetrics(true),
		redis.WithDialTimeout(time.Second*5),
		redis.WithMaxRetries(3),
	)
	require.NoError(b, err)
	defer redisClient.Close()

	opts := &session.Options{
		Redis: &session.RedisOptions{
			Addr: redisAddr,
		},
		KeyPrefix:     "bench:",
		EnableMetrics: true,
	}

	store := session.NewRedisStore(redisClient, opts)
	require.NoError(b, store.Ping(ctx))

	b.Run("Save", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sess := &session.Session{
				UserID:    fmt.Sprintf("bench-user-%d", i),
				TokenID:   fmt.Sprintf("bench-token-%d", i),
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := store.Save(ctx, sess); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预先创建一个会话用于测试
		sess := &session.Session{
			UserID:    "bench-user",
			TokenID:   "bench-token",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		require.NoError(b, store.Save(ctx, sess))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := store.Get(ctx, sess.TokenID); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Delete", func(b *testing.B) {
		sessions := make([]*session.Session, b.N)
		// 预先创建会话
		for i := 0; i < b.N; i++ {
			sessions[i] = &session.Session{
				UserID:    fmt.Sprintf("del-user-%d", i),
				TokenID:   fmt.Sprintf("del-token-%d", i),
				ExpiresAt: time.Now().Add(time.Hour),
			}
			require.NoError(b, store.Save(ctx, sessions[i]))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := store.Delete(ctx, sessions[i].TokenID); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("BatchOperations", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sess := &session.Session{
				UserID:    fmt.Sprintf("batch-user-%d", i),
				TokenID:   fmt.Sprintf("batch-token-%d", i),
				ExpiresAt: time.Now().Add(time.Hour),
			}

			// 执行一系列操作
			if err := store.Save(ctx, sess); err != nil {
				b.Fatal(err)
			}
			if _, err := store.Get(ctx, sess.TokenID); err != nil {
				b.Fatal(err)
			}
			if err := store.Delete(ctx, sess.TokenID); err != nil {
				b.Fatal(err)
			}
		}
	})
}
