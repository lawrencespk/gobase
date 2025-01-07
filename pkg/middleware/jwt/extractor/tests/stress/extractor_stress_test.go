package stress

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/middleware/jwt/extractor"
)

func setupRouter(e extractor.TokenExtractor) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		token, err := e.Extract(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	return r
}

func TestExtractor_StressConcurrent(t *testing.T) {
	tests := []struct {
		name      string
		extractor extractor.TokenExtractor
		setup     func(*http.Request)
	}{
		{
			name:      "Header提取器-并发压力",
			extractor: extractor.NewHeaderExtractor("Authorization", "Bearer "),
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.123")
			},
		},
		{
			name:      "Cookie提取器-并发压力",
			extractor: extractor.NewCookieExtractor("jwt"),
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "test.token.456",
				})
			},
		},
		{
			name:      "Query提取器-并发压力",
			extractor: extractor.NewQueryExtractor("token"),
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "test.token.789")
				r.URL.RawQuery = q.Encode()
			},
		},
		{
			name: "链式提取器-并发压力",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.chain")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(tt.extractor)

			// 并发数
			concurrency := 100
			// 每个goroutine的请求数
			requestsPerGoroutine := 1000

			var wg sync.WaitGroup
			start := time.Now()

			// 错误计数器
			errorCount := int32(0)

			// 启动并发goroutine
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; j < requestsPerGoroutine; j++ {
						w := httptest.NewRecorder()
						req := httptest.NewRequest("GET", "/test", nil)
						tt.setup(req)
						router.ServeHTTP(w, req)

						if w.Code != http.StatusOK {
							atomic.AddInt32(&errorCount, 1)
						}
					}
				}()
			}

			// 等待所有goroutine完成
			wg.Wait()
			duration := time.Since(start)

			// 计算统计信息
			totalRequests := concurrency * requestsPerGoroutine
			requestsPerSecond := float64(totalRequests) / duration.Seconds()

			// 输出测试结果
			t.Logf("总请求数: %d", totalRequests)
			t.Logf("总耗时: %v", duration)
			t.Logf("QPS: %.2f", requestsPerSecond)
			t.Logf("错误数: %d", errorCount)

			// 验证错误率
			errorRate := float64(errorCount) / float64(totalRequests)
			assert.Less(t, errorRate, 0.001, "错误率应小于0.1%")
		})
	}
}
