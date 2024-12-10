package tests

import (
	"context"
	"testing"
	"time"

	basecontext "gobase/pkg/context"
	"gobase/pkg/context/types"

	"github.com/stretchr/testify/assert"
)

func TestBaseContext(t *testing.T) {
	// 测试基本的元数据操作
	t.Run("metadata operations", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		// 测试设置和获取
		ctx.SetValue("key", "value")
		val := ctx.GetValue("key")
		assert.Equal(t, "value", val)

		// 测试删除
		ctx.DeleteValue("key")
		val = ctx.GetValue("key")
		assert.Nil(t, val)
	})

	// 测试批量设置
	t.Run("batch metadata operations", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		data := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		for k, v := range data {
			ctx.SetValue(k, v)
		}

		// 验证所有值都被正确设置
		val1 := ctx.GetValue("key1")
		assert.Equal(t, "value1", val1)

		val2 := ctx.GetValue("key2")
		assert.Equal(t, "value2", val2)
	})

	// 测试用户信息
	t.Run("user info", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		ctx.SetValue(basecontext.KeyUserID, "123")
		userID := ctx.GetValue(basecontext.KeyUserID)
		assert.Equal(t, "123", userID)

		ctx.SetValue(basecontext.KeyUserName, "test")
		userName := ctx.GetValue(basecontext.KeyUserName)
		assert.Equal(t, "test", userName)
	})

	// 测试元数据类型转换
	t.Run("metadata type conversion", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		// 测试整数值
		ctx.SetValue("int-key", 123)
		intVal, ok := types.GetIntValue(ctx, "int-key")
		assert.True(t, ok)
		assert.Equal(t, 123, intVal)

		// 测试浮点值
		ctx.SetValue("float-key", 123.456)
		floatVal, ok := types.GetFloat64Value(ctx, "float-key")
		assert.True(t, ok)
		assert.Equal(t, 123.456, floatVal)

		// 测试布尔值
		ctx.SetValue("bool-key", true)
		boolVal, ok := types.GetBoolValue(ctx, "bool-key")
		assert.True(t, ok)
		assert.True(t, boolVal)

		// 测试时间值
		now := time.Now()
		ctx.SetValue("time-key", now)
		timeVal, ok := types.GetTimeValue(ctx, "time-key")
		assert.True(t, ok)
		assert.Equal(t, now, timeVal)
	})

	// 测试元数据边界情况
	t.Run("metadata edge cases", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		// 测试空值
		ctx.SetValue("empty-key", nil)
		val := ctx.GetValue("empty-key")
		assert.Nil(t, val)

		// 测试类型不匹配
		ctx.SetValue("type-mismatch", "not-an-int")
		intVal, ok := types.GetIntValue(ctx, "type-mismatch")
		assert.False(t, ok)
		assert.Equal(t, 0, intVal)
	})

	// 测试 Metadata() 方法
	t.Run("metadata map copy", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())

		// 设置初始数据
		ctx.SetValue("key1", "value1")

		// 获取并修改 map
		metadataMap := ctx.Metadata()
		metadataMap["key2"] = "value2"

		// 验证原始数据未被修改
		val := ctx.GetValue("key2")
		assert.Nil(t, val)
	})

	// 测试并发安全性
	t.Run("metadata concurrency", func(t *testing.T) {
		ctx := basecontext.NewContext(context.TODO())
		done := make(chan bool)

		// 并发写入
		go func() {
			for i := 0; i < 100; i++ {
				ctx.SetValue("key", i)
			}
			done <- true
		}()

		// 并发读取
		go func() {
			for i := 0; i < 100; i++ {
				_ = ctx.GetValue("key")
			}
			done <- true
		}()

		// 等待完成
		<-done
		<-done
	})
}
