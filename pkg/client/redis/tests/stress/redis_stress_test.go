package stress

import (
	"context"
	"fmt"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"sync"
	"testing"
	"time"
)

func TestStressRedisOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	client, err := redis.NewClient(
		redis.WithAddress(addr),
		redis.WithPoolSize(100),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	concurrency := 100
	operations := 1000

	t.Run("concurrent_operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errCh := make(chan error, concurrency*operations)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					key := fmt.Sprintf("stress_key_%d_%d", routineID, j)
					if err := client.Set(ctx, key, "value", time.Minute); err != nil {
						errCh <- err
						return
					}
					if _, err := client.Get(ctx, key); err != nil {
						errCh <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Errorf("stress test error: %v", err)
		}
	})
}
