package contexttest

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
	r.Use(ctxMiddleware.Middleware())

	r.GET("/test", func(c *gin.Context) {
		ctx := ctxMiddleware.GetContextFromGin(c)

		// 测试设置和获取值
		ctx.SetMetadata("test_key", "test_value")

		// 验证值是否正确设置
		val, exists := ctx.GetMetadata("test_key")
		assert.True(t, exists)
		assert.Equal(t, "test_value", val)

		c.String(200, "OK")
	})

	// 执行测试
	w := performRequest(r, "GET", "/test")
	assert.Equal(t, 200, w.Code)
}

// 测试错误处理
func TestErrorHandling(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 初始状态应该没有错误
	assert.False(t, ctx.HasError(), "初始状态不应该有错误")
	assert.Nil(t, ctx.GetError(), "初始错误应该为nil")

	// 设置错误
	err := errors.New("test error")
	ctx.SetError(err)

	// 验证错误已设置
	assert.True(t, ctx.HasError(), "设置错误后HasError应该返回true")
	assert.Equal(t, err, ctx.GetError(), "GetError应该返回设置的错误")

	// 清除错误
	ctx.SetError(nil)

	// 验证错误已清除
	currentErr := ctx.GetError()
	t.Logf("清除错误后的当前错误: %v", currentErr)
	assert.False(t, ctx.HasError(), "清除错误后HasError应该返回false")
	assert.Nil(t, ctx.GetError(), "清除错误后GetError应该返回nil")
}

// 测试追踪相关功能
func TestTracing(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试 TraceID
	ctx.SetTraceID("trace-123")
	assert.Equal(t, "trace-123", ctx.GetTraceID())

	// 测试 SpanID
	ctx.SetSpanID("span-456")
	assert.Equal(t, "span-456", ctx.GetSpanID())
}

// 测试超时控制
func TestTimeoutControl(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试 WithTimeout
	t.Run("timeout", func(t *testing.T) {
		timeoutCtx, cancel := ctx.WithTimeout(100 * time.Millisecond)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			assert.Error(t, timeoutCtx.Err())
		case <-time.After(200 * time.Millisecond):
			t.Error("context should have timed out")
		}
	})

	// 测试 WithDeadline
	t.Run("deadline", func(t *testing.T) {
		deadline := time.Now().Add(100 * time.Millisecond)
		deadlineCtx, cancel := ctx.WithDeadline(deadline)
		defer cancel()

		gotDeadline, ok := deadlineCtx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, deadline, gotDeadline)
	})

	// 测试 WithCancel
	t.Run("cancel", func(t *testing.T) {
		cancelCtx, cancel := ctx.WithCancel()
		cancel()
		assert.Error(t, cancelCtx.Err())
	})
}

// 测试克隆功能
func TestClone(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 设置一些测试数据
	ctx.SetMetadata("key", "value")
	ctx.SetUserID("user-123")
	ctx.SetRequestID("req-456")

	// 克隆上下文
	clonedCtx := ctx.Clone()

	// 验证克隆的数据
	val, ok := clonedCtx.GetMetadata("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)
	assert.Equal(t, "user-123", clonedCtx.GetUserID())
	assert.Equal(t, "req-456", clonedCtx.GetRequestID())

	// 验证修改克隆不影响原始上下文
	clonedCtx.SetMetadata("key", "new-value")
	originalVal, _ := ctx.GetMetadata("key")
	assert.Equal(t, "value", originalVal)
}

// 测试并发安全性
func TestConcurrency(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			ctx.SetMetadata(key, i)
		}(i)
	}

	// 并发读取
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			_, _ = ctx.GetMetadata(key)
		}(i)
	}

	wg.Wait()

	// 验证数据完整性
	count := 0
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		if val, ok := ctx.GetMetadata(key); ok {
			assert.Equal(t, i, val)
			count++
		}
	}
	assert.Equal(t, 100, count)
}

// 测试验证器
func TestValidator(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试缺少必要信息时的验证
	err := baseCtx.ValidateUserContext(ctx)
	assert.Error(t, err)

	// 设置必要信息后的验证
	ctx.SetUserID("user-123")
	ctx.SetUserName("test-user")
	err = baseCtx.ValidateUserContext(ctx)
	assert.NoError(t, err)

	// 测试追踪上下文验证
	err = baseCtx.ValidateTraceContext(ctx)
	assert.Error(t, err)

	ctx.SetRequestID("req-123")
	ctx.SetTraceID("trace-123")
	err = baseCtx.ValidateTraceContext(ctx)
	assert.NoError(t, err)
}

// 测试值类型转换
func TestValueConversion(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试字符串值
	ctx.SetMetadata("string-key", "test-value")
	strVal, ok := baseCtx.GetStringValue(ctx, "string-key")
	assert.True(t, ok)
	assert.Equal(t, "test-value", strVal)

	// 测试整数值
	ctx.SetMetadata("int-key", 123)
	intVal, ok := baseCtx.GetIntValue(ctx, "int-key")
	assert.True(t, ok)
	assert.Equal(t, 123, intVal)

	// 测试布尔值
	ctx.SetMetadata("bool-key", true)
	boolVal, ok := baseCtx.GetBoolValue(ctx, "bool-key")
	assert.True(t, ok)
	assert.True(t, boolVal)

	// 测试浮点值
	ctx.SetMetadata("float-key", 123.456)
	floatVal, ok := baseCtx.GetFloat64Value(ctx, "float-key")
	assert.True(t, ok)
	assert.Equal(t, 123.456, floatVal)

	// 测试时间值
	now := time.Now()
	ctx.SetMetadata("time-key", now)
	timeVal, ok := baseCtx.GetTimeValue(ctx, "time-key")
	assert.True(t, ok)
	assert.Equal(t, now, timeVal)
}

// 测试请求信息
func TestRequestInfo(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试设置和获取请求ID
	ctx.SetRequestID("req-123")
	assert.Equal(t, "req-123", ctx.GetRequestID(), "请求ID应该正确设置和获取")

	// 测试设置和获取客户端IP
	ctx.SetClientIP("192.168.1.1")
	assert.Equal(t, "192.168.1.1", ctx.GetClientIP(), "客户端IP应该正确设置和获取")
}

// 测试边界条件
func TestBoundaryConditions(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试设置空值
	ctx.SetMetadata("empty-key", "")
	val, ok := ctx.GetMetadata("empty-key")
	assert.True(t, ok, "应该能够获取设置的空值")
	assert.Equal(t, "", val, "空值应该正确获取")

	// 测试删除不存在的键
	ctx.DeleteMetadata("non-existent-key")
	_, ok = ctx.GetMetadata("non-existent-key")
	assert.False(t, ok, "删除不存在的键后，应该无法获取")
}

// 测试无效值类型
func TestInvalidValueType(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 设置一个非字符串的值
	ctx.SetMetadata("int-key", 123)

	// 尝试获取为字符串
	strVal, ok := baseCtx.GetStringValue(ctx, "int-key")
	assert.False(t, ok, "不应该成功获取非字符串值为字符串")
	assert.Equal(t, "", strVal, "获取非字符串值为字符串时应该返回空字符串")
}

// 测试 Metadata 方法返回的 map 是否是副本
func TestMetadataMapCopy(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 设置初始数据
	ctx.SetMetadata("key1", "value1")

	// 获取 metadata map
	metadataMap := ctx.Metadata()

	// 修改获取到的 map
	metadataMap["key2"] = "value2"

	// 验证原始上下文中的数据没有被修改
	_, exists := ctx.GetMetadata("key2")
	assert.False(t, exists, "修改返回的 map 不应影响原始数据")
}

// 测试克隆的完整性
func TestCompleteClone(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 设置所有可能的字段
	ctx.SetUserID("user123")
	ctx.SetUserName("username")
	ctx.SetRequestID("req123")
	ctx.SetClientIP("192.168.1.1")
	ctx.SetTraceID("trace123")
	ctx.SetSpanID("span123")
	ctx.SetError(errors.New("test error"))

	// 克隆上下文
	cloned := ctx.Clone()

	// 验证所有字段
	assert.Equal(t, ctx.GetUserID(), cloned.GetUserID())
	assert.Equal(t, ctx.GetUserName(), cloned.GetUserName())
	assert.Equal(t, ctx.GetRequestID(), cloned.GetRequestID())
	assert.Equal(t, ctx.GetClientIP(), cloned.GetClientIP())
	assert.Equal(t, ctx.GetTraceID(), cloned.GetTraceID())
	assert.Equal(t, ctx.GetSpanID(), cloned.GetSpanID())
	assert.Equal(t, ctx.GetError(), cloned.GetError())
}

// 测试上下文取消后的行为
func TestContextCancellation(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())
	cancelCtx, cancel := ctx.WithCancel()

	// 取消上下文
	cancel()

	// 验证取消后的行为
	assert.Error(t, cancelCtx.Err(), "取消后应返回错误")
	assert.Equal(t, context.Canceled, cancelCtx.Err(), "应该返回 context.Canceled 错误")

	// 验证在取消后设置和获取值是否正常
	cancelCtx.SetMetadata("key", "value")
	val, ok := cancelCtx.GetMetadata("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val, "取消后仍应能正常操作元数据")
}

// 测试超时边界情况
func TestTimeoutEdgeCases(t *testing.T) {
	ctx := baseCtx.NewContext(context.TODO())

	// 测试零超时
	t.Run("zero timeout", func(t *testing.T) {
		timeoutCtx, cancel := ctx.WithTimeout(0)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			assert.Error(t, timeoutCtx.Err())
		default:
			t.Error("context should timeout immediately")
		}
	})

	// 测试负超时
	t.Run("negative timeout", func(t *testing.T) {
		timeoutCtx, cancel := ctx.WithTimeout(-1 * time.Second)
		defer cancel()

		select {
		case <-timeoutCtx.Done():
			assert.Error(t, timeoutCtx.Err())
		default:
			t.Error("context should timeout immediately")
		}
	})

	// 测试过去的截止时间
	t.Run("past deadline", func(t *testing.T) {
		deadline := time.Now().Add(-1 * time.Hour)
		deadlineCtx, cancel := ctx.WithDeadline(deadline)
		defer cancel()

		select {
		case <-deadlineCtx.Done():
			assert.Error(t, deadlineCtx.Err())
		default:
			t.Error("context should be done immediately")
		}
	})
}
