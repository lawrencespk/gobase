package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/store"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
)

func TestStressStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	// 创建文件管理器
	fm := logrus.NewFileManager(logrus.FileOptions{
		DefaultPath: "logs/stress_test.log",
	})

	// 创建队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,
		BatchSize:       100,
		Workers:         1,
		FlushInterval:   time.Second,
		RetryCount:      3,
		RetryInterval:   time.Second,
		MaxBatchWait:    time.Second * 5,
		ShutdownTimeout: time.Second * 10,
	}

	// 创建logger选项
	opts := &logrus.Options{
		Level: types.DebugLevel,
	}

	// 创建logger
	log, err := logrus.NewLogger(fm, queueConfig, opts)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Sync()

	// 创建 Redis 客户端
	client, err := redis.NewClient(
		redis.WithAddress(addr),
		redis.WithPoolSize(100),
		redis.WithEnableMetrics(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Redis存储测试
	redisOpts := &store.Options{
		Redis: &store.RedisOptions{
			Addr: addr,
		},
		EnableMetrics: true,
		EnableTracing: true,
		KeyPrefix:     "test:",
	}

	redisStore := store.NewRedisTokenStore(client, redisOpts, log)
	runStressTest(t, "Redis", redisStore)

	// 内存存储测试
	memOpts := store.Options{
		EnableMetrics:   true,
		EnableTracing:   true,
		KeyPrefix:       "test:",
		CleanupInterval: time.Minute,
	}
	memStore, err := store.NewMemoryStore(memOpts)
	if err != nil {
		t.Fatalf("Failed to create memory store: %v", err)
	}
	runStressTest(t, "Memory", memStore)
}

func runStressTest(t *testing.T, name string, s store.Store) {
	const (
		workers    = 10
		iterations = 1000
	)

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// 创建token
				token := fmt.Sprintf("token-%d-%d", workerID, j)

				// 创建 Claims
				claims := jwt.NewStandardClaims(
					jwt.WithUserID(fmt.Sprintf("user-%d", workerID)),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)

				// 创建 TokenInfo
				info := &jwt.TokenInfo{
					Raw:       token,
					Type:      jwt.AccessToken,
					ExpiresAt: time.Now().Add(time.Hour),
					IsRevoked: false,
					Claims:    claims, // 设置 Claims
				}

				// 存储token
				err := s.Set(context.Background(), token, info, time.Hour)
				if err != nil {
					metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("set", err.Error()).Inc()
					continue
				}

				// 获取token
				_, err = s.Get(context.Background(), token)
				if err != nil {
					metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("get", err.Error()).Inc()
					continue
				}

				// 删除token
				err = s.Delete(context.Background(), token)
				if err != nil {
					metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("delete", err.Error()).Inc()
				}

				duration := time.Since(startTime).Seconds()
				metrics.DefaultJWTMetrics.SessionOperations.WithLabelValues("complete").Observe(duration)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("[%s] Stress test completed", name)
}
