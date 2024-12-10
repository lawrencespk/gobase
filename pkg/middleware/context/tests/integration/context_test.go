package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gobase/pkg/context"
	"gobase/pkg/context/types"
	contextMiddleware "gobase/pkg/middleware/context"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置默认的上下文创建函数
	types.DefaultNewContext = context.NewContext
}

func TestContextMiddleware_Basic(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	r := gin.New()

	// 使用中间件
	r.Use(contextMiddleware.Middleware(nil))

	// 测试路由
	r.GET("/test", func(c *gin.Context) {
		// 获取上下文
		ctx := contextMiddleware.GetContextFromGin(c)
		assert.NotNil(t, ctx)

		// 测试元数据操作
		metadata := map[string]interface{}{
			"test_key1": "test_value1",
			"test_key2": "test_value2",
		}
		ctx.SetMetadata(metadata)

		// 获取并验证元数据
		gotMetadata := ctx.GetMetadata()
		assert.Equal(t, "test_value1", gotMetadata["test_key1"])
		assert.Equal(t, "test_value2", gotMetadata["test_key2"])

		// 测试请求ID
		assert.NotEmpty(t, ctx.GetRequestID())

		c.String(http.StatusOK, "success")
	})

	// 创建测试请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestContextMiddleware_RequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contextMiddleware.Middleware(nil))

	r.GET("/test", func(c *gin.Context) {
		ctx := contextMiddleware.GetContextFromGin(c)
		assert.NotNil(t, ctx)

		// 验证请求ID存在且不为空
		requestID := ctx.GetRequestID()
		assert.NotEmpty(t, requestID)

		// 验证响应头中包含请求ID
		assert.Equal(t, requestID, c.Writer.Header().Get("X-Request-ID"))

		c.String(http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContextMiddleware_Metadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contextMiddleware.Middleware(nil))

	r.GET("/test", func(c *gin.Context) {
		ctx := contextMiddleware.GetContextFromGin(c)
		assert.NotNil(t, ctx)

		// 测试设置元数据
		metadata := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		ctx.SetMetadata(metadata)

		// 测试获取元数据
		gotMetadata := ctx.GetMetadata()
		assert.Equal(t, "value1", gotMetadata["key1"])
		assert.Equal(t, "value2", gotMetadata["key2"])

		// 测试删除元数据
		ctx.DeleteMetadata("key1")
		gotMetadata = ctx.GetMetadata()
		assert.Empty(t, gotMetadata["key1"])
		assert.Equal(t, "value2", gotMetadata["key2"])

		c.String(http.StatusOK, "success")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
