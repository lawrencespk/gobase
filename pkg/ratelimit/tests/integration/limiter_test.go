package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/client"
	"gobase/pkg/cache/redis/config/types"
	"gobase/pkg/cache/redis/tests/testutils"
	"gobase/pkg/ratelimit/redis"
)

func TestMain(m *testing.M) {
	// 启动Redis容器
	cleanup, err := testutils.StartRedisContainer()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 运行测试
	m.Run()
}

// setupRedisClient 创建一个用于测试的Redis客户端
func setupRedisClient(t *testing.T) client.Client {
	require := require.New(t)

	// 创建Redis配置 - 使用本地Docker容器
	cfg := &types.Config{
		Addresses:    []string{"localhost:6379"}, // Docker映射的本地端口
		Username:     "",
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// 创建Redis客户端
	redisClient, err := client.NewClient(cfg)
	require.NoError(err, "Failed to create Redis client")
	require.NotNil(redisClient, "Redis client should not be nil")

	return redisClient
}

func TestSlidingWindowLimiter_Integration(t *testing.T) {
	// 创建Redis客户端
	redisClient := setupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	// 创建限流器
	limiter := redis.NewSlidingWindowLimiter(redisClient)

	// 测试场景
	tests := []struct {
		name      string
		key       string
		limit     int64
		window    time.Duration
		calls     int
		wantAllow bool
	}{
		{
			name:      "should allow first request",
			key:       "test_key_1",
			limit:     2,
			window:    time.Second,
			calls:     1,
			wantAllow: true,
		},
		{
			name:      "should reject when over limit",
			key:       "test_key_2",
			limit:     2,
			window:    time.Second,
			calls:     3,
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// 重置限流器
			err := limiter.Reset(ctx, tt.key)
			require.NoError(t, err)

			// 执行多次请求
			var lastResult bool
			for i := 0; i < tt.calls; i++ {
				lastResult, err = limiter.Allow(ctx, tt.key, tt.limit, tt.window)
				require.NoError(t, err)
			}

			// 验证最后一次请求的结果
			assert.Equal(t, tt.wantAllow, lastResult)
		})
	}
}

func TestSlidingWindowLimiter_Concurrent(t *testing.T) {
	// 创建Redis客户端
	redisClient := setupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	// 创建限流器
	limiter := redis.NewSlidingWindowLimiter(redisClient)

	// 测试参数
	const (
		key        = "test_concurrent"
		limit      = int64(10)
		window     = time.Second
		goroutines = 20
	)

	// 重置限流器
	ctx := context.Background()
	err := limiter.Reset(ctx, key)
	require.NoError(t, err)

	// 并发测试
	done := make(chan bool)
	results := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			allowed, err := limiter.Allow(ctx, key, limit, window)
			require.NoError(t, err)
			results <- allowed
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < goroutines; i++ {
		<-done
	}
	close(results)

	// 统计结果
	allowed := 0
	for result := range results {
		if result {
			allowed++
		}
	}

	// 验证结果
	assert.Equal(t, int(limit), allowed, "Should allow exactly %d requests", limit)
}
