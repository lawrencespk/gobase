package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestContextIntegration(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 准备测试数据
	testData := struct {
		claims      jwt.Claims
		token       string
		tokenType   jwt.TokenType
		userID      string
		userName    string
		roles       []string
		permissions []string
		deviceID    string
		ipAddress   string
	}{
		claims: &MockClaims{
			StandardClaims: jwt.NewStandardClaims(
				jwt.WithUserID("test-user"),
				jwt.WithUserName("Test User"),
				jwt.WithRoles([]string{"admin", "user"}),
				jwt.WithPermissions([]string{"read", "write"}),
				jwt.WithDeviceID("test-device"),
				jwt.WithIPAddress("127.0.0.1"),
				jwt.WithTokenType(jwt.AccessToken),
			),
		},
		token:       "test-token",
		tokenType:   jwt.AccessToken,
		userID:      "test-user",
		userName:    "Test User",
		roles:       []string{"admin", "user"},
		permissions: []string{"read", "write"},
		deviceID:    "test-device",
		ipAddress:   "127.0.0.1",
	}

	tests := []struct {
		name       string
		setupCtx   func(ctx context.Context) context.Context
		middleware gin.HandlerFunc
		handler    gin.HandlerFunc
		validate   func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "完整JWT上下文传递",
			setupCtx: func(ctx context.Context) context.Context {
				return jwtContext.WithJWTContext(ctx, testData.claims, testData.token)
			},
			middleware: func(c *gin.Context) {
				// 验证上下文中的所有值
				claims, err := jwtContext.GetClaims(c.Request.Context())
				require.NoError(t, err)
				assert.Equal(t, testData.claims, claims)

				token, err := jwtContext.GetToken(c.Request.Context())
				require.NoError(t, err)
				assert.Equal(t, testData.token, token)

				c.Next()
			},
			handler: func(c *gin.Context) {
				// 在处理器中验证上下文值
				userID, err := jwtContext.GetUserID(c.Request.Context())
				require.NoError(t, err)
				assert.Equal(t, testData.userID, userID)

				c.JSON(http.StatusOK, gin.H{"status": "success"})
			},
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, resp.Code)
			},
		},
		{
			name: "上下文值修改传递",
			setupCtx: func(ctx context.Context) context.Context {
				return jwtContext.WithUserID(ctx, testData.userID)
			},
			middleware: func(c *gin.Context) {
				// 在中间件中修改上下文值
				ctx := jwtContext.WithUserName(c.Request.Context(), testData.userName)
				c.Request = c.Request.WithContext(ctx)
				c.Next()
			},
			handler: func(c *gin.Context) {
				// 验证原始值和修改后的值
				userID, err := jwtContext.GetUserID(c.Request.Context())
				require.NoError(t, err)
				assert.Equal(t, testData.userID, userID)

				userName, err := jwtContext.GetUserName(c.Request.Context())
				require.NoError(t, err)
				assert.Equal(t, testData.userName, userName)

				c.JSON(http.StatusOK, gin.H{"status": "success"})
			},
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, resp.Code)
			},
		},
		{
			name: "错误处理",
			setupCtx: func(ctx context.Context) context.Context {
				return ctx // 不设置任何值
			},
			middleware: func(c *gin.Context) {
				// 验证缺失值的错误处理
				_, err := jwtContext.GetClaims(c.Request.Context())
				assert.Error(t, err)
				c.Next()
			},
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			},
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(tt.setupCtx(c.Request.Context()))
				c.Next()
			})
			router.Use(tt.middleware)
			router.GET("/test", tt.handler)

			// 创建测试请求
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(resp, req)

			// 验证结果
			tt.validate(t, resp)
		})
	}
}

func TestContextChain(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	router := gin.New()

	// 定义中间件链
	middlewares := []gin.HandlerFunc{
		func(c *gin.Context) {
			ctx := jwtContext.WithUserID(c.Request.Context(), "user-1")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		},
		func(c *gin.Context) {
			ctx := jwtContext.WithRoles(c.Request.Context(), []string{"admin"})
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		},
		func(c *gin.Context) {
			ctx := jwtContext.WithPermissions(c.Request.Context(), []string{"read"})
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		},
	}

	// 应用中间件链
	router.Use(middlewares...)

	// 定义处理器
	router.GET("/test", func(c *gin.Context) {
		// 验证所有中间件设置的值
		userID, err := jwtContext.GetUserID(c.Request.Context())
		require.NoError(t, err)
		assert.Equal(t, "user-1", userID)

		roles, err := jwtContext.GetRoles(c.Request.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"admin"}, roles)

		permissions, err := jwtContext.GetPermissions(c.Request.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"read"}, permissions)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// 创建测试请求
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(resp, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, resp.Code)
}
