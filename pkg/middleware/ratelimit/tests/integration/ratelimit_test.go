package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/logger"
	"gobase/pkg/middleware/ratelimit"
	"gobase/pkg/middleware/ratelimit/tests/testutils"
	redisLimiter "gobase/pkg/ratelimit/redis"
)

func TestRateLimit_Integration(t *testing.T) {
	// 初始化日志
	_ = logger.InitializeLogger()

	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 从环境变量获取Redis配置
	redisHost := os.Getenv("TEST_REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("TEST_REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	// 创建Redis客户端
	redisClient := testutils.SetupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	// 创建限流器
	limiter := redisLimiter.NewSlidingWindowLimiter(redisClient)

	// 创建测试路由
	r := gin.New()
	r.Use(ratelimit.RateLimit(&ratelimit.Config{
		Limiter: limiter,
		Limit:   2,
		Window:  time.Second,
	}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// 执行测试请求
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "success", w.Body.String())
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}
