package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/ratelimit"
	"gobase/pkg/cache/redis/ratelimit/tests/testutils"
	"gobase/pkg/client/redis"
)

func TestStore_StressEval(t *testing.T) {
	// 启动 Redis 容器
	container, err := testutils.StartRedisContainer(t)
	require.NoError(t, err)
	defer container.Terminate(t)

	// 获取 Redis 地址
	redisAddr := container.GetAddress()

	// 创建 Redis 客户端
	client, err := redis.NewClient(
		redis.WithAddresses([]string{redisAddr}),
		redis.WithDialTimeout(time.Second),
		redis.WithReadTimeout(time.Second),
		redis.WithWriteTimeout(time.Second),
	)
	require.NoError(t, err)
	defer client.Close()

	// 确保 Redis 连接正常
	err = client.Ping(context.Background())
	require.NoError(t, err)

	store := ratelimit.NewStore(client)
	ctx := context.Background()

	script := `return KEYS[1]`
	concurrency := 100
	iterations := 1000

	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*iterations)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("stress-key-%d-%d", workerID, j)
				_, err := store.Eval(ctx, script, []string{key}, 1)
				if err != nil {
					errCh <- fmt.Errorf("worker %d iteration %d failed: %v", workerID, j, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	duration := time.Since(start)
	opsPerSec := float64(concurrency*iterations) / duration.Seconds()

	t.Logf("Stress test completed: %d operations in %v (%.2f ops/sec)",
		concurrency*iterations, duration, opsPerSec)

	// 检查是否有错误发生
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}
	require.Empty(t, errors, "stress test encountered errors")
}
