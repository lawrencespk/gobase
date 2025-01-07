package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtmw "gobase/pkg/middleware/jwt"
	"gobase/pkg/monitor/prometheus/metrics"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *jwt.TokenManager) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建token管理器
	tokenManager, err := jwt.NewTokenManager("integration-test-secret")
	require.NoError(t, err)

	// 创建中间件
	middleware, err := jwtmw.New(tokenManager,
		jwtmw.WithMetrics(true),
		jwtmw.WithTracing(true),
	)
	require.NoError(t, err)

	// 创建路由
	router := gin.New()
	router.Use(middleware.Handle())

	// 添加测试路由
	router.GET("/protected", func(c *gin.Context) {
		// 获取Claims
		claims, exists := jwt.FromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":   claims.GetUserID(),
			"user_name": claims.GetUserName(),
		})
	})

	return router, tokenManager
}

func TestJWTMiddlewareIntegration(t *testing.T) {
	router, tokenManager := setupTestRouter(t)

	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, error)
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
		expectedStatus int
	}{
		{
			name: "成功访问受保护的路由",
			setupRequest: func() (*http.Request, error) {
				// 创建带有必要字段的claims
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithUserName("Test User"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)

				// 生成token
				token, err := tokenManager.GenerateToken(context.Background(), claims)
				if err != nil {
					return nil, err
				}

				// 创建请求
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				return req, nil
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, resp.Code)
				assert.Contains(t, resp.Body.String(), "test-user")
				assert.Contains(t, resp.Body.String(), "Test User")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "过期token",
			setupRequest: func() (*http.Request, error) {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(-time.Hour)), // 过期时间设置为过去
				)

				token, err := tokenManager.GenerateToken(context.Background(), claims)
				if err != nil {
					return nil, err
				}

				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				return req, nil
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, resp.Code)
				assert.Contains(t, resp.Body.String(), "token has expired")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "无效的token格式",
			setupRequest: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "InvalidToken")
				return req, nil
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, resp.Code)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录初始指标值
			initialErrorCount := testutil.ToFloat64(metrics.DefaultJWTMetrics.TokenValidationErrors)

			// 执行请求
			req, err := tt.setupRequest()
			require.NoError(t, err)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// 检查响应
			tt.checkResponse(t, resp)
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// 验证指标收集
			if tt.expectedStatus != http.StatusOK {
				newErrorCount := testutil.ToFloat64(metrics.DefaultJWTMetrics.TokenValidationErrors)
				assert.Greater(t, newErrorCount, initialErrorCount, "错误计数应该增加")
			}
		})
	}
}
