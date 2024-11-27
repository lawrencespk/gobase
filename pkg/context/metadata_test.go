package context

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseContext(t *testing.T) {
	// 测试基本的元数据操作
	t.Run("metadata operations", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		// 测试设置和获取
		ctx.SetMetadata("key", "value")
		val, ok := ctx.GetMetadata("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)

		// 测试删除
		ctx.DeleteMetadata("key")
		_, ok = ctx.GetMetadata("key")
		assert.False(t, ok)
	})

	// 测试批量设置
	t.Run("batch metadata operations", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		data := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		ctx.SetMetadataMap(data)

		// 验证所有值都被正确设置
		val1, ok := ctx.GetMetadata("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", val1)

		val2, ok := ctx.GetMetadata("key2")
		assert.True(t, ok)
		assert.Equal(t, "value2", val2)
	})

	// 测试用户信息
	t.Run("user info", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		ctx.SetUserID("123")
		assert.Equal(t, "123", ctx.GetUserID())

		ctx.SetUserName("test")
		assert.Equal(t, "test", ctx.GetUserName())
	})

	// 测试元数据类型转换
	t.Run("metadata type conversion", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		// 测试整数值
		ctx.SetMetadata("int-key", 123)
		intVal, ok := GetIntValue(ctx, "int-key")
		assert.True(t, ok)
		assert.Equal(t, 123, intVal)

		// 测试浮点值
		ctx.SetMetadata("float-key", 123.456)
		floatVal, ok := GetFloat64Value(ctx, "float-key")
		assert.True(t, ok)
		assert.Equal(t, 123.456, floatVal)

		// 测试布尔值
		ctx.SetMetadata("bool-key", true)
		boolVal, ok := GetBoolValue(ctx, "bool-key")
		assert.True(t, ok)
		assert.True(t, boolVal)

		// 测试时间值
		now := time.Now()
		ctx.SetMetadata("time-key", now)
		timeVal, ok := GetTimeValue(ctx, "time-key")
		assert.True(t, ok)
		assert.Equal(t, now, timeVal)
	})

	// 测试元数据边界情况
	t.Run("metadata edge cases", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		// 测试空值
		ctx.SetMetadata("empty-key", nil)
		val, ok := ctx.GetMetadata("empty-key")
		assert.True(t, ok)
		assert.Nil(t, val)

		// 测试类型不匹配
		ctx.SetMetadata("type-mismatch", "not-an-int")
		intVal, ok := GetIntValue(ctx, "type-mismatch")
		assert.False(t, ok)
		assert.Equal(t, 0, intVal)
	})

	// 测试 Metadata() 方法
	t.Run("metadata map copy", func(t *testing.T) {
		ctx := NewContext(context.TODO())

		// 设置初始数据
		ctx.SetMetadata("key1", "value1")

		// 获取并修改 map
		metadataMap := ctx.Metadata()
		metadataMap["key2"] = "value2"

		// 验证原始数据未被修改
		_, exists := ctx.GetMetadata("key2")
		assert.False(t, exists)
	})

	// 测试并发安全性
	t.Run("metadata concurrency", func(t *testing.T) {
		ctx := NewContext(context.TODO())
		done := make(chan bool)

		// 并发写入
		go func() {
			for i := 0; i < 100; i++ {
				ctx.SetMetadata("key", i)
			}
			done <- true
		}()

		// 并发读取
		go func() {
			for i := 0; i < 100; i++ {
				ctx.GetMetadata("key")
			}
			done <- true
		}()

		// 等待完成
		<-done
		<-done
	})
}
