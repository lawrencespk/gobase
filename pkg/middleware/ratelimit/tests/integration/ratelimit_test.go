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
		KeyFunc: func(c *gin.Context) string {
			return "test-key"
		},
		Limit:  2,
		Window: time.Second,
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

func TestRateLimit_DistributedScenario(t *testing.T) {
	// 初始化日志
	_ = logger.InitializeLogger()
	gin.SetMode(gin.TestMode)

	// 设置Redis
	redisClient := testutils.SetupRedisClient(t)
	require.NotNil(t, redisClient)
	defer redisClient.Close()

	// 创建多个限流器实例，模拟多个服务节点
	limiter1 := redisLimiter.NewSlidingWindowLimiter(redisClient)
	limiter2 := redisLimiter.NewSlidingWindowLimiter(redisClient)

	// 使用固定的key进行测试
	const testKey = "test-distributed"
	keyFunc := func(*gin.Context) string {
		return testKey
	}

	// 创建两个服务实例
	r1 := gin.New()
	r1.Use(ratelimit.RateLimit(&ratelimit.Config{
		Limiter: limiter1,    // 使用相同的limiter
		Limit:   2,           // 使用相同的limit
		Window:  time.Second, // 使用相同的window
		KeyFunc: keyFunc,     // 使用相同的key确保分布式限流
	}))
	r1.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success from instance 1")
	})

	r2 := gin.New()
	r2.Use(ratelimit.RateLimit(&ratelimit.Config{
		Limiter: limiter2,    // 使用相同的limiter
		Limit:   2,           // 使用相同的limit
		Window:  time.Second, // 使用相同的window
		KeyFunc: keyFunc,     // 使用相同的key确保分布式限流
	}))
	r2.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success from instance 2")
	})

	// 测试分布式限流
	t.Run("distributed rate limiting", func(t *testing.T) {
		// 第一个实例的请求
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		r1.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 第二个实例的请求
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		r2.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// 第三个请求应该被限流（无论来自哪个实例）
		w3 := httptest.NewRecorder()
		req3, _ := http.NewRequest("GET", "/test", nil)
		r1.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)

		w4 := httptest.NewRecorder()
		req4, _ := http.NewRequest("GET", "/test", nil)
		r2.ServeHTTP(w4, req4)
		assert.Equal(t, http.StatusTooManyRequests, w4.Code)
	})

	// 测试计数器一致性
	t.Run("counter consistency", func(t *testing.T) {
		// 等待上一个测试的窗口期结束
		time.Sleep(time.Second)

		// 使用两个实例交替发送请求
		instances := []*gin.Engine{r1, r2}
		for i := 0; i < 4; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			instances[i%2].ServeHTTP(w, req)

			if i < 2 {
				assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be limited", i+1)
			}
		}
	})
}
