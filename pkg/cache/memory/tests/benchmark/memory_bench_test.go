package benchmark

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"gobase/pkg/cache/memory"
	"gobase/pkg/logger"
)

// setupCache 创建测试用的缓存实例
func setupCache(b *testing.B, maxEntries int) *memory.Cache {
	log, err := logger.NewLogger()
	if err != nil {
		b.Fatal(err)
	}

	cache, err := memory.NewCache(&memory.Config{
		MaxEntries:      maxEntries,
		CleanupInterval: time.Hour,
		DefaultTTL:      time.Hour,
	}, log)
	if err != nil {
		b.Fatal(err)
	}

	return cache
}

// BenchmarkCache_Basic 基本操作的基准测试
func BenchmarkCache_Basic(b *testing.B) {
	ctx := context.Background()

	b.Run("Set", func(b *testing.B) {
		cache := setupCache(b, b.N+1000) // 确保有足够空间
		defer cache.Stop()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i)
			err := cache.Set(ctx, key, "value", time.Hour)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get/Hit", func(b *testing.B) {
		cache := setupCache(b, b.N+1000) // 确保有足够空间
		defer cache.Stop()

		// 预填充数据
		key := "test-key"
		err := cache.Set(ctx, key, "test-value", time.Hour)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cache.Get(ctx, key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get/Miss", func(b *testing.B) {
		cache := setupCache(b, b.N+1000) // 确保有足够空间
		defer cache.Stop()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(ctx, "non-existent-key")
		}
	})

	b.Run("Delete", func(b *testing.B) {
		cache := setupCache(b, b.N+1000)
		defer cache.Stop()

		// 预填充数据
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			keys[i] = fmt.Sprintf("del-key-%d", i)
			err := cache.Set(ctx, keys[i], "value", time.Hour)
			if err != nil {
				b.Fatal(err)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cache.Delete(ctx, keys[i]) // 忽略删除错误，因为key可能已经不存在
		}
	})
}

// BenchmarkCache_Concurrent 并发基准测试
func BenchmarkCache_Concurrent(b *testing.B) {
	cache := setupCache(b, 1000000)
	defer cache.Stop()
	ctx := context.Background()

	// 使用固定的key范围，避免无限增长
	const keyRange = 10000

	for _, procs := range []int{1, 4, 8, 16, 32} {
		b.Run(fmt.Sprintf("Mixed/Procs-%d", procs), func(b *testing.B) {
			b.SetParallelism(procs)
			b.RunParallel(func(pb *testing.PB) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				for pb.Next() {
					key := strconv.Itoa(r.Intn(keyRange))
					switch r.Intn(10) {
					case 0, 1, 2:
						_ = cache.Set(ctx, key, "value", time.Hour)
					case 3, 4, 5, 6, 7, 8:
						_, _ = cache.Get(ctx, key)
					default:
						_ = cache.Delete(ctx, key)
					}
				}
			})
		})
	}
}

// BenchmarkCache_DataSize 不同数据量的基准测试
func BenchmarkCache_DataSize(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}
	ctx := context.Background()

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			cache := setupCache(b, size*2) // 确保容量足够
			defer cache.Stop()

			// 预填充数据
			for i := 0; i < size; i++ {
				key := fmt.Sprintf("key-%d", i)
				err := cache.Set(ctx, key, "value", time.Hour)
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				for pb.Next() {
					key := fmt.Sprintf("key-%d", r.Intn(size))
					_, _ = cache.Get(ctx, key)
				}
			})
		})
	}
}

// BenchmarkCache_TTL TTL相关的基准测试
func BenchmarkCache_TTL(b *testing.B) {
	cache := setupCache(b, 1000000)
	defer cache.Stop()
	ctx := context.Background()

	b.Run("SetWithDifferentTTL", func(b *testing.B) {
		ttls := []time.Duration{
			time.Second,
			time.Minute,
			time.Hour,
			24 * time.Hour,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i)
			ttl := ttls[i%len(ttls)]
			err := cache.Set(ctx, key, "value", ttl)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetWithExpiration", func(b *testing.B) {
		// 预填充即将过期的数据
		key := "expire-key"
		err := cache.Set(ctx, key, "value", 100*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond) // 等待接近过期

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(ctx, key)
		}
	})
}
