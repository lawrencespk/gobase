package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	baseCtx "gobase/pkg/context"
	"gobase/pkg/context/types"
	ctxMiddleware "gobase/pkg/middleware/context"
)

// 执行请求
func performRequest(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

// 测试上下文传播
func TestContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 设置默认的上下文创建函数
	types.DefaultNewContext = baseCtx.NewContext

	// 创建中间件选项
	opts := &ctxMiddleware.Options{}

	// 使用上下文中间件
	r.Use(ctxMiddleware.Middleware(opts))

	r.GET("/test", func(c *gin.Context) {
		ctx := ctxMiddleware.GetContextFromGin(c)

		// 设置元数据
		metadata := map[string]interface{}{
			"test_key": "test_value",
		}
		ctx.SetMetadata(metadata)

		// 获取并验证元数据
		md := ctx.GetMetadata()
		val, ok := md["test_key"]
		assert.True(t, ok)
		assert.Equal(t, "test_value", val)

		c.String(200, "OK")
	})

	w := performRequest(r, "GET", "/test")
	assert.Equal(t, 200, w.Code)
}

// 测试错误处理
func TestErrorHandling(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 初始状态
	assert.Nil(t, ctx.GetError())

	// 设置错误
	err := errors.New("test error")
	ctx.SetError(err)

	// 验证错误
	assert.Equal(t, err, ctx.GetError())

	// 清除错误
	ctx.SetError(nil)
	assert.Nil(t, ctx.GetError())
}

// 测试元数据处理
func TestMetadataHandling(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试字符串值
	ctx.SetMetadata(map[string]interface{}{
		"string-key": "test-value",
	})
	md := ctx.GetMetadata()
	val, ok := md["string-key"]
	assert.True(t, ok)
	assert.Equal(t, "test-value", val)

	// 测试整数值
	ctx.SetMetadata(map[string]interface{}{
		"int-key": 123,
	})
	md = ctx.GetMetadata()
	intVal, ok := md["int-key"].(int)
	assert.True(t, ok)
	assert.Equal(t, 123, intVal)

	// 测试布尔值
	ctx.SetMetadata(map[string]interface{}{
		"bool-key": true,
	})
	md = ctx.GetMetadata()
	boolVal, ok := md["bool-key"].(bool)
	assert.True(t, ok)
	assert.True(t, boolVal)

	// 测试浮点值
	ctx.SetMetadata(map[string]interface{}{
		"float-key": 123.456,
	})
	md = ctx.GetMetadata()
	floatVal, ok := md["float-key"].(float64)
	assert.True(t, ok)
	assert.Equal(t, 123.456, floatVal)

	// 测试时间值
	now := time.Now()
	ctx.SetMetadata(map[string]interface{}{
		"time-key": now,
	})
	md = ctx.GetMetadata()
	timeVal, ok := md["time-key"].(time.Time)
	assert.True(t, ok)
	assert.Equal(t, now, timeVal)
}

// 测试并发安全性
func TestConcurrentAccess(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())
	var wg sync.WaitGroup
	iterations := 1000

	// 并发写入
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx.SetMetadata(map[string]interface{}{
				fmt.Sprintf("key-%d", val): val,
			})
		}(i)
	}

	wg.Wait()

	// 验证数据完整性
	md := ctx.GetMetadata()
	count := 0
	for k, v := range md {
		if val, ok := v.(int); ok {
			assert.Contains(t, k, fmt.Sprintf("key-%d", val))
			count++
		}
	}
	assert.Equal(t, iterations, count)
}

// 测试上下文取消
func TestContextCancellation(t *testing.T) {
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := baseCtx.NewContext(parentCtx)

	// 设置元数据
	ctx.SetMetadata(map[string]interface{}{
		"test-key": "test-value",
	})

	// 取消上下文
	cancel()

	// 验证上下文已取消
	assert.Error(t, ctx.Err())
	assert.Equal(t, context.Canceled, ctx.Err())

	// 验证元数据仍然可以访问
	md := ctx.GetMetadata()
	val, ok := md["test-key"]
	assert.True(t, ok)
	assert.Equal(t, "test-value", val)
}
