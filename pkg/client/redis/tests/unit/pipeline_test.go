package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	// 创建 miniredis 实例
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	// 使用 miniredis 地址创建客户端
	client, err := redis.NewClient(
		redis.WithAddress(mr.Addr()),
	)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("basic pipeline", func(t *testing.T) {
		// 创建管道
		pipe := client.TxPipeline()
		assert.NotNil(t, pipe)

		// 添加命令到管道
		err := pipe.Set(ctx, "pipe_key1", "value1", time.Minute)
		assert.NoError(t, err)
		err = pipe.Set(ctx, "pipe_key2", "value2", time.Minute)
		assert.NoError(t, err)

		// 执行管道
		cmds, err := pipe.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 2)

		// 验证结果
		val1, err := client.Get(ctx, "pipe_key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val1)

		val2, err := client.Get(ctx, "pipe_key2")
		assert.NoError(t, err)
		assert.Equal(t, "value2", val2)

		// 清理
		_, err = client.Del(ctx, "pipe_key1", "pipe_key2")
		assert.NoError(t, err)
	})

	t.Run("pipeline with negative expiration", func(t *testing.T) {
		pipe := client.TxPipeline()
		assert.NotNil(t, pipe)

		// 添加一个负数过期时间的命令
		err := pipe.Set(ctx, "pipe_key3", "value3", -1*time.Second)
		assert.NoError(t, err) // pipeline.Set 不会立即执行，所以不会报错

		// 执行管道
		cmds, err := pipe.Exec(ctx)
		assert.NoError(t, err) // miniredis 不会对负数过期时间报错
		assert.NotNil(t, cmds)

		// 直接在 miniredis 上设置过期时间
		mr.SetTTL("pipe_key3", -1*time.Second)
		mr.FastForward(time.Second) // 推进时间以确保过期生效

		// 验证 key 是否存在
		exists := mr.Exists("pipe_key3")
		assert.False(t, exists, "key should not exist with negative expiration")

		// 通过客户端验证
		val, err := client.Get(ctx, "pipe_key3")
		assert.Error(t, err)     // 应该返回错误，因为 key 不存在
		assert.Equal(t, "", val) // 值应该为空

		// 清理，以防万一
		_, err = client.Del(ctx, "pipe_key3")
		assert.NoError(t, err)
	})

	t.Run("pipeline with mixed operations", func(t *testing.T) {
		pipe := client.TxPipeline()
		assert.NotNil(t, pipe)

		// 添加不同类型的操作，使用不同的 key
		err := pipe.Set(ctx, "pipe_str", "value", time.Minute)
		assert.NoError(t, err)

		_, err = pipe.HSet(ctx, "pipe_hash", "field1", "value1")
		assert.NoError(t, err)

		_, err = pipe.ZAdd(ctx, "pipe_zset", &redis.Z{
			Score:  1.0,
			Member: "member1",
		})
		assert.NoError(t, err)

		// 执行管道
		cmds, err := pipe.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 3)

		// 清理
		_, err = client.Del(ctx, "pipe_str", "pipe_hash", "pipe_zset")
		assert.NoError(t, err)
	})

	t.Run("pipeline discard", func(t *testing.T) {
		pipe := client.TxPipeline()
		assert.NotNil(t, pipe)

		// 添加命令
		err := pipe.Set(ctx, "pipe_discard", "value", time.Minute)
		assert.NoError(t, err)

		// 丢弃管道
		err = pipe.Close()
		assert.NoError(t, err)

		// 验证命令未执行
		exists, err := client.Exists(ctx, "pipe_discard")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
