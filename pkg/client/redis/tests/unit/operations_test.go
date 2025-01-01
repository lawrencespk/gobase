package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/client/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisOperations(t *testing.T) {
	// 创建 miniredis 实例
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// 创建 Redis 客户端
	client, err := redis.NewClient(
		redis.WithAddress(mr.Addr()),
	)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("string operations", func(t *testing.T) {
		// 测试 Set 和 Get
		err := client.Set(ctx, "test_key", "test_value", time.Minute)
		assert.NoError(t, err)

		val, err := client.Get(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)

		// 测试过期
		mr.FastForward(time.Minute + time.Second)
		_, err = client.Get(ctx, "test_key")
		assert.Error(t, err)
	})

	t.Run("hash operations", func(t *testing.T) {
		// 测试 HSet 和 HGet
		n, err := client.HSet(ctx, "test_hash", "field1", "value1", "field2", "value2")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)

		val, err := client.HGet(ctx, "test_hash", "field1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)

		// 测试不存在的字段
		_, err = client.HGet(ctx, "test_hash", "non_existent")
		assert.Error(t, err)
	})

	t.Run("sorted set operations", func(t *testing.T) {
		// 测试 ZAdd 和 ZRem
		members := []*redis.Z{
			{Score: 1.0, Member: "member1"},
			{Score: 2.0, Member: "member2"},
		}
		n, err := client.ZAdd(ctx, "test_zset", members...)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)

		// 测试删除成员
		n, err = client.ZRem(ctx, "test_zset", "member1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("publish/subscribe", func(t *testing.T) {
		// 测试发布消息
		err := client.Publish(ctx, "test_channel", "test_message")
		assert.NoError(t, err)

		// 测试空频道名
		err = client.Publish(ctx, "", "test_message")
		assert.Error(t, err)
	})

	t.Run("lua script", func(t *testing.T) {
		script := `return ARGV[1]`
		result, err := client.Eval(ctx, script, []string{}, "test_value")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", result)
	})

	t.Run("key operations", func(t *testing.T) {
		// 测试键是否存在
		err := client.Set(ctx, "exists_key", "value", time.Minute)
		assert.NoError(t, err)

		exists, err := client.Exists(ctx, "exists_key")
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = client.Exists(ctx, "non_existent_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("error cases", func(t *testing.T) {
		// 测试获取不存在的键
		_, err := client.Get(ctx, "non_existent")
		assert.Error(t, err)

		// 测试设置空键
		err = client.Set(ctx, "", "value", time.Minute)
		assert.Error(t, err)

		// 测试 Hash 操作参数验证
		_, err = client.HGet(ctx, "", "field") // 空键
		assert.Error(t, err)

		_, err = client.HGet(ctx, "key", "") // 空字段
		assert.Error(t, err)

		// 添加 Publish 方法的错误测试
		err = client.Publish(ctx, "", "message") // 空channel
		assert.Error(t, err)

		err = client.Publish(ctx, "test_channel", nil) // nil message
		assert.Error(t, err)
	})

	// 添加 Publish 的正常功能测试
	t.Run("publish/subscribe operations", func(t *testing.T) {
		channel := "test_channel"
		message := "test_message"

		// 测试发布消息
		err := client.Publish(ctx, channel, message)
		assert.NoError(t, err)
	})

	t.Run("publish/subscribe operations", func(t *testing.T) {
		channel := "test_channel"
		message := "test_message"

		// 创建订阅者
		pubsub := client.Subscribe(ctx, channel)
		require.NotNil(t, pubsub)
		defer pubsub.Close()

		// 在goroutine中接收消息
		msgChan := make(chan string)
		errChan := make(chan error)
		go func() {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				errChan <- err
				return
			}
			msgChan <- msg.Payload
		}()

		// 发布消息
		err := client.Publish(ctx, channel, message)
		assert.NoError(t, err)

		// 等待接收消息或错误
		select {
		case receivedMsg := <-msgChan:
			assert.Equal(t, message, receivedMsg)
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for message")
		}
	})

	t.Run("subscribe error cases", func(t *testing.T) {
		// 测试订阅空频道
		pubsub := client.Subscribe(ctx)
		_, err := pubsub.ReceiveMessage(ctx)
		assert.Error(t, err)
		assert.NoError(t, pubsub.Close())

		// 测试订阅后关闭
		pubsub = client.Subscribe(ctx, "test_channel")
		require.NotNil(t, pubsub)

		err = pubsub.Close()
		assert.NoError(t, err)

		// 关闭后接收消息应该返回错误
		_, err = pubsub.ReceiveMessage(ctx)
		assert.Error(t, err)
	})

	t.Run("multiple subscribers", func(t *testing.T) {
		channel := "test_multi_channel"
		message := "test_message"

		// 创建多个订阅者
		sub1 := client.Subscribe(ctx, channel)
		sub2 := client.Subscribe(ctx, channel)
		defer sub1.Close()
		defer sub2.Close()

		// 在goroutine中接收消息
		msgChan1 := make(chan string)
		msgChan2 := make(chan string)
		go func() {
			msg, err := sub1.ReceiveMessage(ctx)
			if err == nil {
				msgChan1 <- msg.Payload
			}
		}()
		go func() {
			msg, err := sub2.ReceiveMessage(ctx)
			if err == nil {
				msgChan2 <- msg.Payload
			}
		}()

		// 发布消息
		err := client.Publish(ctx, channel, message)
		assert.NoError(t, err)

		// 验证两个订阅者都收到消息
		for i := 0; i < 2; i++ {
			select {
			case msg1 := <-msgChan1:
				assert.Equal(t, message, msg1)
			case msg2 := <-msgChan2:
				assert.Equal(t, message, msg2)
			case <-time.After(time.Second):
				assert.Fail(t, "timeout waiting for message")
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		channel := "test_cancel_channel"

		// 创建带取消的上下文
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 创建订阅者
		pubsub := client.Subscribe(ctx, channel)
		defer pubsub.Close()

		// 取消上下文
		cancel()

		// 验证接收消息会返回上下文取消错误
		_, err := pubsub.ReceiveMessage(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
