package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/auth/jwt/session/tests/testutils"
	"gobase/pkg/client/redis"

	"github.com/stretchr/testify/require"
)

func TestSessionUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	ctx := context.Background()
	redisAddr, err := testutils.StartRedisContainer(ctx)
	require.NoError(t, err)
	defer testutils.StopRedisContainer()

	// 初始化Redis客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress(redisAddr),
		redis.WithEnableMetrics(true),
		redis.WithDialTimeout(time.Second*5),
		redis.WithMaxRetries(3),
	)
	require.NoError(t, err)
	defer redisClient.Close()

	opts := &session.Options{
		Redis: &session.RedisOptions{
			Addr: redisAddr,
		},
		KeyPrefix:     "stress:",
		EnableMetrics: true,
	}

	store := session.NewRedisStore(redisClient, opts)
	require.NoError(t, store.Ping(ctx))

	t.Run("High Concurrent Writes", func(t *testing.T) {
		const (
			numGoroutines = 100
			numOperations = 1000
		)

		var wg sync.WaitGroup
		startTime := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < numOperations; j++ {
					sess := &session.Session{
						UserID:    fmt.Sprintf("user-%d-%d", routineID, j),
						TokenID:   fmt.Sprintf("token-%d-%d", routineID, j),
						ExpiresAt: time.Now().Add(time.Hour),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}

					if err := store.Save(ctx, sess); err != nil {
						t.Errorf("Failed to save session: %v", err)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)
		opsPerSecond := float64(numGoroutines*numOperations) / duration.Seconds()
		t.Logf("Performed %d operations in %v (%.2f ops/sec)",
			numGoroutines*numOperations, duration, opsPerSecond)
	})

	t.Run("Large Session Management", func(t *testing.T) {
		const numSessions = 10000
		sessions := make([]*session.Session, numSessions)

		// 创建大量会话
		for i := 0; i < numSessions; i++ {
			sessions[i] = &session.Session{
				UserID:    fmt.Sprintf("bulk-user-%d", i),
				TokenID:   fmt.Sprintf("bulk-token-%d", i),
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Metadata: map[string]interface{}{
					"index": i,
				},
			}
		}

		startTime := time.Now()

		// 批量保存
		for _, sess := range sessions {
			if err := store.Save(ctx, sess); err != nil {
				t.Fatalf("Failed to save session: %v", err)
			}
		}

		duration := time.Since(startTime)
		t.Logf("Saved %d sessions in %v (%.2f sessions/sec)",
			numSessions, duration, float64(numSessions)/duration.Seconds())

		// 验证所有会话
		for _, sess := range sessions {
			_, err := store.Get(ctx, sess.TokenID)
			if err != nil {
				t.Errorf("Failed to get session %s: %v", sess.TokenID, err)
			}
		}
	})
}
