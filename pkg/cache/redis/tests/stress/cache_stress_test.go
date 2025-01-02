package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
)

// go test -v -timeout 5m gobase/pkg/cache/redis/tests/stress -run TestCacheStress

func TestCacheStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// 启动Redis容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	// 创建Redis客户端
	client, err := redisClient.NewClient(
		redisClient.WithAddresses([]string{addr}),
		redisClient.WithDialTimeout(time.Second*5),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// 创建缓存实例
	cache, err := redis.NewCache(redis.Options{
		Client: client,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	const (
		goroutines = 100
		operations = 1000
		duration   = 5 * time.Minute
	)

	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, goroutines*operations)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operations; j++ {
				select {
				case <-time.After(duration):
					return
				default:
					key := fmt.Sprintf("stress_key_%d_%d", id, j)
					value := fmt.Sprintf("value_%d_%d", id, j)

					// Set操作
					if err := cache.Set(ctx, key, value, time.Minute); err != nil {
						errors <- fmt.Errorf("Set error: %v", err)
						continue
					}

					// Get操作
					if _, err := cache.Get(ctx, key); err != nil {
						errors <- fmt.Errorf("Get error: %v", err)
						continue
					}

					// Delete操作
					if err := cache.Delete(ctx, key); err != nil {
						errors <- fmt.Errorf("Delete error: %v", err)
						continue
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 统计错误
	var errorCount int
	for err := range errors {
		t.Log(err)
		errorCount++
	}

	t.Logf("Stress test completed in %v", time.Since(start))
	t.Logf("Total operations: %d", goroutines*operations)
	t.Logf("Error count: %d", errorCount)
	t.Logf("Success rate: %.2f%%", 100-(float64(errorCount)/(float64(goroutines*operations))*100))
}
