package ratelimit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"gobase/pkg/errors/codes"
	"gobase/pkg/middleware/ratelimit"
	mocklimiter "gobase/pkg/middleware/ratelimit/tests/mock"
)

func TestRateLimit_Retry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLimiter := new(mocklimiter.MockLimiter)

	tests := []struct {
		name           string
		config         *ratelimit.Config
		setupMock      func(*mocklimiter.MockLimiter)
		expectedStatus int
		expectedBody   map[string]string
		expectedCalls  int
	}{
		{
			name: "should succeed after first retry",
			config: &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   100,
				Window:  time.Minute,
				Retry: &ratelimit.RetryStrategy{
					MaxAttempts:           3,
					RetryInterval:         time.Millisecond * 10,
					UseExponentialBackoff: false,
				},
			},
			setupMock: func(m *mocklimiter.MockLimiter) {
				m.On("AllowN", testifymock.Anything, testifymock.Anything, int64(1), int64(100), time.Minute).
					Return(false, assert.AnError).Once()
				m.On("AllowN", testifymock.Anything, testifymock.Anything, int64(1), int64(100), time.Minute).
					Return(true, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedCalls:  2,
		},
		{
			name: "should fail after max retries",
			config: &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   100,
				Window:  time.Minute,
				Retry: &ratelimit.RetryStrategy{
					MaxAttempts:           3,
					RetryInterval:         time.Millisecond * 10,
					UseExponentialBackoff: false,
				},
			},
			setupMock: func(m *mocklimiter.MockLimiter) {
				for i := 0; i < 3; i++ {
					m.On("AllowN", testifymock.Anything, testifymock.Anything, int64(1), int64(100), time.Minute).
						Return(false, assert.AnError).Once()
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"code":    codes.RateLimitError,
				"message": "rate limit check failed",
			},
			expectedCalls: 3,
		},
		{
			name: "should use exponential backoff",
			config: &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   100,
				Window:  time.Minute,
				Retry: &ratelimit.RetryStrategy{
					MaxAttempts:           3,
					RetryInterval:         time.Millisecond * 10,
					UseExponentialBackoff: true,
					MaxRetryInterval:      time.Millisecond * 100,
				},
			},
			setupMock: func(m *mocklimiter.MockLimiter) {
				var lastCallTime time.Time
				m.On("AllowN", testifymock.Anything, testifymock.Anything, int64(1), int64(100), time.Minute).
					Run(func(args testifymock.Arguments) {
						if !lastCallTime.IsZero() {
							interval := time.Since(lastCallTime)
							assert.True(t, interval >= time.Millisecond*10)
						}
						lastCallTime = time.Now()
					}).Return(false, assert.AnError).Times(3)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCalls:  3,
		},
		{
			name: "should respect context cancellation",
			config: &ratelimit.Config{
				Limiter: mockLimiter,
				Limit:   100,
				Window:  time.Minute,
				Retry: &ratelimit.RetryStrategy{
					MaxAttempts:           3,
					RetryInterval:         time.Millisecond * 100,
					UseExponentialBackoff: false,
				},
			},
			setupMock: func(m *mocklimiter.MockLimiter) {
				m.On("AllowN", testifymock.Anything, testifymock.Anything, int64(1), int64(100), time.Minute).
					Return(false, assert.AnError).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"code":    codes.RateLimitError,
				"message": "rate limit check failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个测试用例创建新的路由实例
			router := gin.New()

			// 重置mock
			mockLimiter.ExpectedCalls = nil
			mockLimiter.Calls = nil

			// 设置mock行为
			tt.setupMock(mockLimiter)

			// 设置测试路由
			router.GET("/test", ratelimit.RateLimit(tt.config), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 发送请求
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// 对于上下文取消的测试
			if tt.name == "should respect context cancellation" {
				ctx, cancel := context.WithCancel(req.Context())
				req = req.WithContext(ctx)
				// 在短时间后取消上下文
				go func() {
					time.Sleep(time.Millisecond * 50)
					cancel()
				}()
			}

			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			// 验证调用次数
			if tt.expectedCalls > 0 {
				assert.Equal(t, tt.expectedCalls, len(mockLimiter.Calls))
			}

			// 验证所有mock期望
			mockLimiter.AssertExpectations(t)
		})
	}
}
