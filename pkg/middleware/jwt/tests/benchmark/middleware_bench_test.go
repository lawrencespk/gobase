package benchmark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtmw "gobase/pkg/middleware/jwt"
)

func BenchmarkMiddleware(b *testing.B) {
	// 创建token管理器
	tokenManager, err := jwt.NewTokenManager("test-secret")
	require.NoError(b, err)

	// 创建中间件
	middleware, err := jwtmw.New(tokenManager)
	require.NoError(b, err)

	// 生成有效token
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
	)
	token, err := tokenManager.GenerateToken(context.Background(), claims)
	require.NoError(b, err)

	// 设置测试场景
	scenarios := []struct {
		name string
		fn   func(*testing.B)
	}{
		{
			name: "有效Token",
			fn: func(b *testing.B) {
				gin.SetMode(gin.ReleaseMode)
				router := gin.New()
				router.Use(middleware.Handle())
				router.GET("/test", func(c *gin.Context) {
					c.Status(http.StatusOK)
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					router.ServeHTTP(w, req)
				}
			},
		},
		{
			name: "无效Token",
			fn: func(b *testing.B) {
				gin.SetMode(gin.ReleaseMode)
				router := gin.New()
				router.Use(middleware.Handle())
				router.GET("/test", func(c *gin.Context) {
					c.Status(http.StatusOK)
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer invalid-token")
				w := httptest.NewRecorder()

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					router.ServeHTTP(w, req)
				}
			},
		},
	}

	for _, s := range scenarios {
		b.Run(s.name, s.fn)
	}
}
