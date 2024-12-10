package unit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	baseContext "gobase/pkg/context"
	"gobase/pkg/context/types"
	contextMiddleware "gobase/pkg/middleware/context"
)

func init() {
	fmt.Println("Setting up DefaultNewContext...")
	types.DefaultNewContext = baseContext.NewContext
	if types.DefaultNewContext == nil {
		panic("Failed to set DefaultNewContext")
	}
	fmt.Println("DefaultNewContext set successfully")
}

func TestContextMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(*gin.Engine)
		request    func() *http.Request
		assertFunc func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "基础功能测试 - 验证上下文注入",
			setup: func(r *gin.Engine) {
				r.Use(contextMiddleware.Middleware(nil))
				r.GET("/test", func(c *gin.Context) {
					ctx := contextMiddleware.GetContextFromGin(c)
					assert.NotNil(t, ctx)

					// 使用 SetValue 设置值
					ctx.SetValue("test_key", "test_value")
					value := ctx.GetValue("test_key")
					assert.Equal(t, "test_value", value)

					c.String(http.StatusOK, "ok")
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "请求ID测试",
			setup: func(r *gin.Engine) {
				r.Use(contextMiddleware.Middleware(nil))
				r.GET("/test", func(c *gin.Context) {
					ctx := contextMiddleware.GetContextFromGin(c)
					requestID := ctx.GetRequestID()

					// 验证请求ID不为空
					assert.NotEmpty(t, requestID)

					// 验证Header中的请求ID与上下文中的一致
					headerID := c.Writer.Header().Get("X-Request-ID")
					assert.Equal(t, requestID, headerID)

					c.Status(http.StatusOK)
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)

				// 验证响应Header中的请求ID
				assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			tt.setup(r)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.request())
			tt.assertFunc(t, w)
		})
	}
}
