package benchmark

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/middleware/ratelimit"
	mocklimiter "gobase/pkg/middleware/ratelimit/tests/mock"
)

func BenchmarkRateLimit(b *testing.B) {
	gin.SetMode(gin.ReleaseMode) // 使用发布模式以避免调试日志影响性能

	tests := []struct {
		name   string
		config *ratelimit.Config
		setup  func(*mocklimiter.MockLimiter)
	}{
		{
			name: "basic_allow",
			config: &ratelimit.Config{
				Limit:  1000,
				Window: time.Second,
			},
			setup: func(m *mocklimiter.MockLimiter) {
				m.On("AllowN", mock.Anything, mock.Anything, int64(1), int64(1000), time.Second).
					Return(true, nil)
			},
		},
		{
			name: "with_retry",
			config: &ratelimit.Config{
				Limit:  1000,
				Window: time.Second,
				Retry: &ratelimit.RetryStrategy{
					MaxAttempts:   3,
					RetryInterval: time.Millisecond,
				},
			},
			setup: func(m *mocklimiter.MockLimiter) {
				m.On("AllowN", mock.Anything, mock.Anything, int64(1), int64(1000), time.Second).
					Return(true, nil)
			},
		},
		{
			name: "with_wait_mode",
			config: &ratelimit.Config{
				Limit:       1000,
				Window:      time.Second,
				WaitMode:    true,
				WaitTimeout: time.Second,
			},
			setup: func(m *mocklimiter.MockLimiter) {
				m.On("Wait", mock.Anything, mock.Anything, int64(1000), time.Second).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			mockLimiter := new(mocklimiter.MockLimiter)
			tt.config.Limiter = mockLimiter
			tt.setup(mockLimiter)

			router := gin.New()
			router.Use(ratelimit.RateLimit(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 重置计时器以排除设置时间
			b.ResetTimer()

			// 并行基准测试
			b.RunParallel(func(pb *testing.PB) {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				for pb.Next() {
					router.ServeHTTP(w, req)
				}
			})
		})
	}
}

// BenchmarkRateLimit_Concurrent 测试不同并发级别下的性能
func BenchmarkRateLimit_Concurrent(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	concurrencyLevels := []int{1, 10, 50, 100, 500, 1000}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			mockLimiter := new(mocklimiter.MockLimiter)
			mockLimiter.On("AllowN", mock.Anything, mock.Anything, int64(1), int64(1000), time.Second).
				Return(true, nil)

			config := &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   1000,
				Window:  time.Second,
			}

			router := gin.New()
			router.Use(ratelimit.RateLimit(config))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 重置计时器
			b.ResetTimer()

			// 创建并发请求
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				for pb.Next() {
					router.ServeHTTP(w, req)
				}
			})
		})
	}
}

// BenchmarkRateLimit_WithLoad 测试在不同负载下的性能
func BenchmarkRateLimit_WithLoad(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	loads := []struct {
		name     string
		loadFunc func()
	}{
		{
			name:     "no_load",
			loadFunc: func() {},
		},
		{
			name: "cpu_load",
			loadFunc: func() {
				// 模拟CPU密集型操��
				for i := 0; i < 1000; i++ {
					_ = fmt.Sprintf("test-%d", i)
				}
			},
		},
		{
			name: "memory_load",
			loadFunc: func() {
				// 模拟内存分配
				data := make([]byte, 1024)
				_ = data
			},
		},
	}

	for _, load := range loads {
		b.Run(load.name, func(b *testing.B) {
			mockLimiter := new(mocklimiter.MockLimiter)
			mockLimiter.On("AllowN", mock.Anything, mock.Anything, int64(1), int64(1000), time.Second).
				Return(true, nil)

			config := &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   1000,
				Window:  time.Second,
			}

			router := gin.New()
			router.Use(ratelimit.RateLimit(config))
			router.GET("/test", func(c *gin.Context) {
				load.loadFunc()
				c.Status(http.StatusOK)
			})

			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				for pb.Next() {
					router.ServeHTTP(w, req)
				}
			})
		})
	}
}
