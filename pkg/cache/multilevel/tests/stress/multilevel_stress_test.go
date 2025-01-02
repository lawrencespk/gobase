package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/multilevel"
	"gobase/pkg/cache/multilevel/tests/testutils"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func TestMultiLevelCacheStress(t *testing.T) {
	// 设置测试环境
	ctx := context.Background()
	redisClient, redisAddr, err := testutils.StartRedisContainer(ctx)
	require.NoError(t, err)
	defer testutils.StopRedisContainer(redisClient)

	// 创建日志实例
	logger, err := logrus.NewLogger(
		logrus.NewFileManager(logrus.FileOptions{
			DefaultPath:   "logs/stress_test.log",
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
			OutputPaths:  []string{"stdout", "logs/stress_test.log"},
		},
	)
	require.NoError(t, err)

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
	require.NoError(t, err)

	// 压力测试参数
	const (
		numGoroutines = 50
		numOperations = 1000
		keyPrefix     = "stress_test_key_"
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 记录开始时间
	startTime := time.Now()

	// 启动多个goroutine并发操作
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("%s%d_%d", keyPrefix, routineID, j)
				value := fmt.Sprintf("value_%d_%d", routineID, j)

				// 写入测试
				err := manager.Set(ctx, key, value, time.Hour)
				require.NoError(t, err)

				// 读取测试
				got, err := manager.Get(ctx, key)
				require.NoError(t, err)
				require.Equal(t, value, got)

				// 删除测试
				err = manager.Delete(ctx, key)
				require.NoError(t, err)

				// 验证删除
				_, err = manager.Get(ctx, key)
				require.Error(t, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 计算总执行时间和操作数
	duration := time.Since(startTime)
	totalOperations := numGoroutines * numOperations * 4 // 每次循环执行4个操作

	t.Logf("Stress Test Results:")
	t.Logf("Total Operations: %d", totalOperations)
	t.Logf("Total Duration: %v", duration)
	t.Logf("Operations/Second: %.2f", float64(totalOperations)/duration.Seconds())
}
