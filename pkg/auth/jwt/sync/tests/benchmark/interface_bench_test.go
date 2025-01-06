package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/client/redis"
)

func BenchmarkRedisOperations(b *testing.B) {
	client, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithPoolSize(10),
		redis.WithMinIdleConns(2),
	)
	if err != nil {
		b.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := client.Set(ctx, "bench_key", "value", 0); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := client.Get(ctx, "bench_key"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Pipeline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipe := client.TxPipeline()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("bench_key_%d_%d", i, j)
				pipe.Set(ctx, key, "value", time.Minute)
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatalf("Pipeline execution failed: %v", err)
			}
		}
	})
}
