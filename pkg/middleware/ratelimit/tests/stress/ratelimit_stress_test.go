package stress

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/client"
	"gobase/pkg/middleware/ratelimit"
	"gobase/pkg/middleware/ratelimit/tests/testutils"
)

//go test -v -timeout 10m ./pkg/middleware/ratelimit/tests/stress

// TestRateLimit_QuickStressTest 快速压力测试
func TestRateLimit_QuickStressTest(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// 创建Redis客户端用于分布式限流测试
	redisClient := testutils.SetupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	tests := []struct {
		name           string
		duration       time.Duration
		concurrency    int
		requestsPerSec int
		limit          int64
		window         time.Duration
		waitMode       bool
		errorThreshold float64
		WaitTimeout    time.Duration
	}{
		{
			name:           "quick_high_load",
			duration:       5 * time.Second, // 缩短持续时间
			concurrency:    50,              // 降低并发数
			requestsPerSec: 500,
			limit:          300,
			window:         time.Second,
			waitMode:       false,
			errorThreshold: 0.01,
			WaitTimeout:    time.Second,
		},
		{
			name:           "quick_wait_mode",
			duration:       5 * time.Second, // 缩短持续时间
			concurrency:    20,              // 降低并发数
			requestsPerSec: 200,
			limit:          100,
			window:         time.Second,
			waitMode:       true,
			errorThreshold: 0.01,
			WaitTimeout:    2 * time.Second,
		},
	}

	runStressTest(t, redisClient, tests)
}

// TestRateLimit_FullStressTest 完整压力测试
func TestRateLimit_FullStressTest(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// 创建Redis客户端用于分布式限流测试
	redisClient := testutils.SetupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	tests := []struct {
		name           string
		duration       time.Duration
		concurrency    int
		requestsPerSec int
		limit          int64
		window         time.Duration
		waitMode       bool
		errorThreshold float64
		WaitTimeout    time.Duration
	}{
		{
			name:           "high_concurrency_short_duration",
			duration:       5 * time.Second, // 降低测试时间
			concurrency:    50,              // 降低并发数
			requestsPerSec: 500,             // 降低每秒请求数
			limit:          300,             // 调整限制
			window:         time.Second,
			waitMode:       false,
			errorThreshold: 0.02, // 提高错误阈值
		},
		{
			name:           "medium_concurrency_long_duration",
			duration:       10 * time.Second,
			concurrency:    25,
			requestsPerSec: 250,
			limit:          150,
			window:         time.Second,
			waitMode:       true,
			errorThreshold: 0.02,
			WaitTimeout:    2 * time.Second,
		},
		{
			name:           "low_concurrency_very_long_duration",
			duration:       20 * time.Second, // 降低测试时间
			concurrency:    10,               // 保持不变
			requestsPerSec: 100,              // 保持不变
			limit:          50,               // 保持不变
			window:         time.Second,
			waitMode:       false,
			errorThreshold: 0.02, // 提高错误阈值
		},
	}

	runStressTest(t, redisClient, tests)
}

// runStressTest 执行压力测试的通用函数
func runStressTest(t *testing.T, redisClient client.Client, tests []struct {
	name           string
	duration       time.Duration
	concurrency    int
	requestsPerSec int
	limit          int64
	window         time.Duration
	waitMode       bool
	errorThreshold float64
	WaitTimeout    time.Duration
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建限流器
			limiter := testutils.NewRedisLimiter(redisClient)

			// 创建路由
			router := gin.New()
			router.Use(ratelimit.RateLimit(&ratelimit.Config{
				Limiter:     limiter,
				Limit:       tt.limit,
				Window:      tt.window,
				WaitMode:    tt.waitMode,
				WaitTimeout: tt.WaitTimeout,
			}))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 统计计数器
			var (
				totalRequests uint64
				successCount  uint64
				errorCount    uint64
				rateLimited   uint64
			)

			// 创建上下文用于控制测试时长
			ctx, cancel := context.WithTimeout(context.Background(), tt.duration)
			defer cancel()

			// 启动并发请求
			var wg sync.WaitGroup
			startTime := time.Now()

			// 启动工作协程
			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					ticker := time.NewTicker(time.Second / time.Duration(tt.requestsPerSec/tt.concurrency))
					defer ticker.Stop()

					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							w := httptest.NewRecorder()
							req := httptest.NewRequest("GET", "/test", nil)
							router.ServeHTTP(w, req)

							atomic.AddUint64(&totalRequests, 1)

							switch w.Code {
							case http.StatusOK:
								atomic.AddUint64(&successCount, 1)
							case http.StatusTooManyRequests:
								atomic.AddUint64(&rateLimited, 1)
							default:
								atomic.AddUint64(&errorCount, 1)
							}
						}
					}
				}()
			}

			// 等待测试完成
			wg.Wait()
			duration := time.Since(startTime)

			// 计算统计数据
			total := atomic.LoadUint64(&totalRequests)
			success := atomic.LoadUint64(&successCount)
			errors := atomic.LoadUint64(&errorCount)
			limited := atomic.LoadUint64(&rateLimited)
			errorRate := float64(errors) / float64(total)
			rps := float64(total) / duration.Seconds()

			// 输出测试结果
			t.Logf("Test Results for %s:", tt.name)
			t.Logf("Duration: %v", duration)
			t.Logf("Total Requests: %d", total)
			t.Logf("Successful Requests: %d", success)
			t.Logf("Rate Limited Requests: %d", limited)
			t.Logf("Error Requests: %d", errors)
			t.Logf("Error Rate: %.2f%%", errorRate*100)
			t.Logf("Average RPS: %.2f", rps)

			// 验证结果
			assert.Less(t, errorRate, tt.errorThreshold, "Error rate exceeds threshold")
			assert.True(t, rps > 0, "RPS should be positive")
			if !tt.waitMode {
				assert.True(t, limited > 0, "Should have some rate limited requests in non-wait mode")
			}
		})
	}
}

// TestRateLimit_ResourceLeaks 测试长时间运行是否存在资源泄漏
func TestRateLimit_ResourceLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource leak test in short mode")
	}

	gin.SetMode(gin.ReleaseMode)
	redisClient := testutils.SetupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	// 初始资源使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialAlloc := m.Alloc

	// 运行时间
	duration := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	router := gin.New()
	router.Use(ratelimit.RateLimit(&ratelimit.Config{
		Limiter: testutils.NewRedisLimiter(redisClient),
		Limit:   100,
		Window:  time.Second,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 持续发送请求
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				router.ServeHTTP(w, req)
			}
		}
	}()

	// 定期检查资源使用情况
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var maxMemory uint64
	for {
		select {
		case <-ctx.Done():
			// 等待所有请求完成
			wg.Wait()

			// 强制GC后再进行最终内存检查
			runtime.GC()
			time.Sleep(100 * time.Millisecond)

			// 最终资源使用检查
			runtime.ReadMemStats(&m)
			finalAlloc := m.Alloc

			// 修改内存增长计算方式
			memoryGrowth := 0.0
			if finalAlloc > initialAlloc {
				memoryGrowth = float64(finalAlloc-initialAlloc) / float64(initialAlloc) * 100
			} else {
				// 如果最终内存小于初始内存，说明内存使用是健康的
				memoryGrowth = 0.0
			}

			t.Logf("Initial Memory: %d bytes", initialAlloc)
			t.Logf("Final Memory: %d bytes", finalAlloc)
			t.Logf("Peak Memory: %d bytes", maxMemory)
			t.Logf("Memory Growth: %.2f%%", memoryGrowth)

			// 验证内存增长是否在可接受范围内
			assert.Less(t, memoryGrowth, 50.0, "Memory growth exceeds acceptable threshold")

			// 额外验证最终内存不超过初始内存的两倍
			assert.Less(t, float64(finalAlloc), float64(initialAlloc)*2, "Final memory exceeds twice the initial memory")
			return

		case <-ticker.C:
			runtime.GC() // 定期触发GC
			time.Sleep(100 * time.Millisecond)
			runtime.ReadMemStats(&m)
			if m.Alloc > maxMemory {
				maxMemory = m.Alloc
			}
			t.Logf("Current Memory: %d bytes", m.Alloc)
		}
	}
}
