package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/cache/multilevel"
	"gobase/pkg/cache/multilevel/tests/testutils"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func BenchmarkMultiLevelCache(b *testing.B) {
	// 设置测试环境
	ctx := context.Background()
	redisClient, redisAddr, err := testutils.StartRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer testutils.StopRedisContainer(redisClient)

	// 创建日志实例
	logger, err := logrus.NewLogger(
		logrus.NewFileManager(logrus.FileOptions{
			DefaultPath:   "logs/bench_test.log",
			BufferSize:    32 * 1024,
			FlushInterval: time.Second,
			MaxOpenFiles:  100,
		}),
		logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     100,
			Workers:       1,
			FlushInterval: time.Second,
		},
		&logrus.Options{
			Level:        types.InfoLevel,
			Development:  true,
			ReportCaller: true,
			TimeFormat:   time.RFC3339,
			OutputPaths:  []string{"stdout", "logs/bench_test.log"},
		},
	)
	if err != nil {
		b.Fatal(err)
	}

	// 创建多级缓存管理器
	manager, err := multilevel.NewManager(&multilevel.Config{
		L1Config: &multilevel.L1Config{
			MaxEntries:      10000,
			CleanupInterval: time.Minute,
		},
		L2Config: &multilevel.L2Config{
			RedisAddr:     redisAddr,
			RedisPassword: "",
			RedisDB:       0,
		},
		L1TTL:             time.Hour,
		EnableAutoWarmup:  true,
		WarmupInterval:    time.Minute,
		WarmupConcurrency: 10,
	}, redisClient, logger)
	if err != nil {
		b.Fatal(err)
	}

	// 基准测试子项
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			if err := manager.Set(ctx, key, value, time.Hour); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		key := "bench_get_key"
		value := "bench_get_value"
		if err := manager.Set(ctx, key, value, time.Hour); err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := manager.Get(ctx, key); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_del_key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			if err := manager.Set(ctx, key, value, time.Hour); err != nil {
				b.Fatal(err)
			}
			if err := manager.Delete(ctx, key); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_mixed_key_%d", i)
			value := fmt.Sprintf("value_%d", i)

			// Set
			if err := manager.Set(ctx, key, value, time.Hour); err != nil {
				b.Fatal(err)
			}

			// Get
			if _, err := manager.Get(ctx, key); err != nil {
				b.Fatal(err)
			}

			// Delete
			if err := manager.Delete(ctx, key); err != nil {
				b.Fatal(err)
			}
		}
	})
}
