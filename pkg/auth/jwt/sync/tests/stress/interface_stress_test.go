package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/client/redis"
)

func TestRedisStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	client, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithEnableMetrics(true),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	workers := 100
	operations := 1000
	var wg sync.WaitGroup

	ctx := context.Background()
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("stress_test_key_%d_%d", workerID, j)
				if err := client.Set(ctx, key, "value", time.Minute); err != nil {
					t.Errorf("Worker %d failed: %v", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	opsPerSecond := float64(workers*operations) / duration.Seconds()
	t.Logf("Stress test completed: %f ops/sec", opsPerSecond)
}
