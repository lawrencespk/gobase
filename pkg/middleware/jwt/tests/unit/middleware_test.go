package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtmw "gobase/pkg/middleware/jwt"
)

func TestNew(t *testing.T) {
	// 创建token管理器
	tokenManager, err := jwt.NewTokenManager("test-secret")
	require.NoError(t, err)

	tests := []struct {
		name    string
		opts    []jwtmw.Option
		wantErr bool
	}{
		{
			name:    "默认选项",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "自定义选项",
			opts: []jwtmw.Option{
				jwtmw.WithTracing(true),
				jwtmw.WithMetrics(true),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := jwtmw.New(tokenManager, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, m)
		})
	}
}

func TestMiddleware_Handle(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建token管理器
	tokenManager, err := jwt.NewTokenManager("test-secret")
	require.NoError(t, err)

	// 创建测试用token，添加必要的字段
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),                  // 添加必要的 UserID
		jwt.WithTokenType(jwt.AccessToken),           // 添加必要的 TokenType
		jwt.WithExpiresAt(time.Now().Add(time.Hour)), // 添加过期时间
	)
	token, err := tokenManager.GenerateToken(context.Background(), claims)
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
	}{
		{
			name: "有效token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+token)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无token",
			setupRequest:   func(r *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "无效token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建中间件
			m, err := jwtmw.New(tokenManager)
			require.NoError(t, err)

			// 创建测试路由
			router := gin.New()
			router.Use(m.Handle())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// 创建测试请求
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupRequest(req)
			resp := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}
