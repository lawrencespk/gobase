package stress

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtContext "gobase/pkg/middleware/jwt/context"
)

// MockClaims 用于测试的Claims实现
type MockClaims struct {
	*jwt.StandardClaims
}

func TestContextStress(t *testing.T) {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 准备测试数据
	claims := &MockClaims{
		StandardClaims: jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithUserName("Test User"),
			jwt.WithRoles([]string{"admin", "user"}),
			jwt.WithPermissions([]string{"read", "write"}),
			jwt.WithDeviceID("test-device"),
			jwt.WithIPAddress("127.0.0.1"),
			jwt.WithTokenType(jwt.AccessToken),
		),
	}
	token := "test-token"

	tests := []struct {
		name         string
		concurrency  int
		requests     int
		setupRouter  func() *gin.Engine
		validateResp func(t *testing.T, responses []int)
		errorHandler func(t *testing.T, err error)
		timeoutAfter time.Duration
	}{
		{
			name:        "高并发读写测试",
			concurrency: 100,
			requests:    1000,
			setupRouter: func() *gin.Engine {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := jwtContext.WithJWTContext(c.Request.Context(), claims, token)
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/test", func(c *gin.Context) {
					// 读取并验证上下文值
					_, err := jwtContext.GetClaims(c.Request.Context())
					if err != nil {
						c.Status(http.StatusInternalServerError)
						return
					}
					c.Status(http.StatusOK)
				})
				return router
			},
			validateResp: func(t *testing.T, responses []int) {
				for _, code := range responses {
					assert.Equal(t, http.StatusOK, code)
				}
			},
			errorHandler: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			timeoutAfter: 30 * time.Second,
		},
		{
			name:        "上下文值并发修改",
			concurrency: 50,
			requests:    500,
			setupRouter: func() *gin.Engine {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := c.Request.Context()
					for i := 0; i < 5; i++ {
						ctx = jwtContext.WithUserID(ctx, "user-"+string(rune(i)))
						ctx = jwtContext.WithRoles(ctx, []string{"role-" + string(rune(i))})
					}
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/test", func(c *gin.Context) {
					c.Status(http.StatusOK)
				})
				return router
			},
			validateResp: func(t *testing.T, responses []int) {
				for _, code := range responses {
					assert.Equal(t, http.StatusOK, code)
				}
			},
			errorHandler: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			timeoutAfter: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := tt.setupRouter()
			responses := make([]int, tt.requests)
			errors := make(chan error, tt.requests)
			var wg sync.WaitGroup

			// 创建工作池
			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					for j := workerID; j < tt.requests; j += tt.concurrency {
						req := httptest.NewRequest(http.MethodGet, "/test", nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						responses[j] = w.Code
					}
				}(i)
			}

			// 等待所有请求完成或超时
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				tt.validateResp(t, responses)
			case err := <-errors:
				tt.errorHandler(t, err)
			case <-time.After(tt.timeoutAfter):
				t.Fatal("测试超时")
			}
		})
	}
}
