package benchmark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gobase/pkg/auth/jwt"
	jwtContext "gobase/pkg/middleware/jwt/context"
)

// MockClaims 用于测试的Claims实现
type MockClaims struct {
	*jwt.StandardClaims
}

func BenchmarkContextOperations(b *testing.B) {
	// 设置Gin为发布模式，避免日志影响性能测试
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

	scenarios := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			name: "WithJWTContext",
			fn: func(b *testing.B) {
				ctx := context.Background()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = jwtContext.WithJWTContext(ctx, claims, token)
				}
			},
		},
		{
			name: "GetAllValues",
			fn: func(b *testing.B) {
				ctx := jwtContext.WithJWTContext(context.Background(), claims, token)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _ = jwtContext.GetClaims(ctx)
					_, _ = jwtContext.GetToken(ctx)
					_, _ = jwtContext.GetUserID(ctx)
					_, _ = jwtContext.GetUserName(ctx)
					_, _ = jwtContext.GetRoles(ctx)
					_, _ = jwtContext.GetPermissions(ctx)
				}
			},
		},
		{
			name: "MiddlewareChain",
			fn: func(b *testing.B) {
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := jwtContext.WithJWTContext(c.Request.Context(), claims, token)
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/test", func(c *gin.Context) {
					c.Status(http.StatusOK)
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()

				b.ResetTimer()
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
