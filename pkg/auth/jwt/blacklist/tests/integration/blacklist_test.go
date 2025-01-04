package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

// setupRedisStore 创建Redis存储实例
func setupRedisStore(t *testing.T) (blacklist.Store, func()) {
	// 创建Redis客户端
	client, err := redis.NewClientFromConfig(&redis.Config{
		Addresses: []string{"localhost:6379"},
		Username:  "",
		Password:  "",
		Database:  0,

		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,

		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,

		EnableMetrics: true,
		EnableTracing: true,
	})
	require.NoError(t, err)

	// 创建日志记录器
	log, err := logger.NewLogger(logger.WithLevel(types.InfoLevel))
	require.NoError(t, err)

	// 创建存储实例
	store, err := blacklist.NewRedisStore(client, &blacklist.Options{
		DefaultExpiration: time.Hour,
		CleanupInterval:   time.Minute,
		Logger:            log,
		EnableMetrics:     true,
		KeyPrefix:         "test:blacklist:",
	})
	require.NoError(t, err)

	// 返回清理函数
	cleanup := func() {
		store.Close()
	}

	return store, cleanup
}

// setupMemoryStore 创建内存存储实例
func setupMemoryStore() (blacklist.Store, func()) {
	store := blacklist.NewMemoryStore()
	cleanup := func() {
		store.Close()
	}
	return store, cleanup
}

func TestRedisStore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	store, cleanup := setupRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Second * 2

	t.Run("Basic Operations", func(t *testing.T) {
		// 添加token
		err := store.Add(ctx, tokenID, reason, expiration)
		require.NoError(t, err)

		// 获取token
		gotReason, err := store.Get(ctx, tokenID)
		require.NoError(t, err)
		assert.Equal(t, reason, gotReason)

		// 等待过期
		time.Sleep(expiration * 2)

		// 验证已过期
		_, err = store.Get(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		const numWorkers = 10
		var wg sync.WaitGroup
		errCh := make(chan error, numWorkers)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				tid := fmt.Sprintf("%s-%d", tokenID, id)

				// 添加token
				if err := store.Add(ctx, tid, reason, expiration); err != nil {
					errCh <- fmt.Errorf("worker %d add error: %v", id, err)
					return
				}

				// 获取token
				if _, err := store.Get(ctx, tid); err != nil {
					errCh <- fmt.Errorf("worker %d get error: %v", id, err)
					return
				}

				// 移除token
				if err := store.Remove(ctx, tid); err != nil {
					errCh <- fmt.Errorf("worker %d remove error: %v", id, err)
					return
				}
			}(i)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Error(err)
		}
	})
}

func TestMemoryStore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	store, cleanup := setupMemoryStore()
	defer cleanup()

	ctx := context.Background()
	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Second * 2

	t.Run("Expiration and Cleanup", func(t *testing.T) {
		// 添加多个token
		for i := 0; i < 10; i++ {
			tid := fmt.Sprintf("%s-%d", tokenID, i)
			err := store.Add(ctx, tid, reason, expiration)
			require.NoError(t, err)
		}

		// 等待部分过期
		time.Sleep(expiration * 2)

		// 验证清理
		for i := 0; i < 10; i++ {
			tid := fmt.Sprintf("%s-%d", tokenID, i)
			_, err := store.Get(ctx, tid)
			assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())
		}
	})
}
