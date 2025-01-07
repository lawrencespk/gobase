package stress

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtmw "gobase/pkg/middleware/jwt"
)

func TestMiddlewareStress(t *testing.T) {
	// 设置测试参数
	concurrency := 100          // 并发goroutine数
	requestsPerGoroutine := 100 // 每个goroutine的请求数
	maxErrorRate := 0.001       // 最大允许错误率0.1%

	// 创建token管理器
	tokenManager, err := jwt.NewTokenManager("test-secret")
	require.NoError(t, err)

	// 创建中间件
	middleware, err := jwtmw.New(tokenManager)
	require.NoError(t, err)

	// 创建测试路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 生成有效token
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
	)
	token, err := tokenManager.GenerateToken(context.Background(), claims)
	require.NoError(t, err)

	// 执行压力测试
	var (
		wg         sync.WaitGroup
		errorCount int32
		start      = time.Now()
	)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				resp := httptest.NewRecorder()

				router.ServeHTTP(resp, req)

				if resp.Code != http.StatusOK {
					atomic.AddInt32(&errorCount, 1)
				}
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()
	duration := time.Since(start)

	// 计算统计信息
	totalRequests := concurrency * requestsPerGoroutine
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	errorRate := float64(errorCount) / float64(totalRequests)

	// 输出测试结果
	t.Logf("总请求数: %d", totalRequests)
	t.Logf("总耗时: %v", duration)
	t.Logf("QPS: %.2f", requestsPerSecond)
	t.Logf("错误数: %d", errorCount)
	t.Logf("错误率: %.4f", errorRate)

	// 验证错误率
	assert.Less(t, errorRate, maxErrorRate, "错误率超过阈值")
}
