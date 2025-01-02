package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/ratelimit"
	"gobase/pkg/cache/redis/ratelimit/tests/testutils"
	"gobase/pkg/client/redis"

	"golang.org/x/sync/errgroup"
)

var (
	container *testutils.RedisContainer
	client    redis.Client
	store     *ratelimit.Store
)

func TestStore(t *testing.T) {
	// 设置测试环境
	setupTest(t)
	defer teardownTest(t)

	t.Run("TestEval", func(t *testing.T) {
		ctx := context.Background()
		script := `return KEYS[1]`
		result, err := store.Eval(
			ctx,
			script,
			[]string{"test-key"},
			1,
		)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", result)
	})

	t.Run("TestDel", func(t *testing.T) {
		ctx := context.Background()

		// 先设置一个键
		err := client.Set(ctx, "test-key", "value", time.Minute)
		require.NoError(t, err)

		// 验证键存在
		exists, err := client.Exists(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, exists)

		// 删除该键
		err = store.Del(ctx, "test-key")
		assert.NoError(t, err)

		// 验证键已被删除
		exists, err = client.Exists(ctx, "test-key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func setupTest(t *testing.T) {
	var err error
	container, err = testutils.StartRedisContainer(t)
	require.NoError(t, err)

	redisAddr := container.GetAddress()

	// 创建 Redis 客户端
	client, err = redis.NewClient(
		redis.WithAddresses([]string{redisAddr}),
		redis.WithDialTimeout(time.Second),
		redis.WithReadTimeout(time.Second),
		redis.WithWriteTimeout(time.Second),
	)
	require.NoError(t, err)

	// 确保 Redis 连接正常
	err = client.Ping(context.Background())
	require.NoError(t, err)

	store = ratelimit.NewStore(client)

	t.Run("Eval", func(t *testing.T) {
		script := `return KEYS[1]`
		result, err := store.Eval(
			context.Background(),
			script,
			[]string{"test-key"},
			1,
		)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", result)
	})

	t.Run("Del", func(t *testing.T) {
		ctx := context.Background()

		// 先设置一个键
		err := client.Set(ctx, "test-key", "value", time.Minute)
		require.NoError(t, err)

		// 验证键存在
		exists, err := client.Exists(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, exists)

		// 删除该键
		err = store.Del(ctx, "test-key")
		assert.NoError(t, err)

		// 验证键已被删除
		exists, err = client.Exists(ctx, "test-key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Concurrent_Operations", func(t *testing.T) {
		ctx := context.Background()
		const concurrency = 10
		const iterations = 100

		// 使用 WaitGroup 来等待所有 goroutine 完成
		var wg sync.WaitGroup
		wg.Add(concurrency)

		// 使用 errgroup 来收集错误
		g, ctx := errgroup.WithContext(ctx)

		for i := 0; i < concurrency; i++ {
			workerID := i // 创建副本避免闭包问题
			g.Go(func() error {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := fmt.Sprintf("concurrent-key-%d-%d", workerID, j)
					script := `return redis.call('SET', KEYS[1], ARGV[1])`
					_, err := store.Eval(ctx, script, []string{key}, "value")
					if err != nil {
						return fmt.Errorf("worker %d iteration %d failed: %v", workerID, j, err)
					}
				}
				return nil
			})
		}

		// 等待所有 goroutine 完成
		wg.Wait()

		// 检查是否有错误发生
		if err := g.Wait(); err != nil {
			t.Errorf("concurrent operations failed: %v", err)
		}
	})
}

func teardownTest(t *testing.T) {
	if client != nil {
		client.Close()
	}
	if container != nil {
		container.Terminate(t)
	}
}
