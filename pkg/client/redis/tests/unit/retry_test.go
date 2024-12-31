package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRetryMechanism(t *testing.T) {
	ctx := context.Background()

	t.Run("retry_success", func(t *testing.T) {
		// 创建 miniredis 实例
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatal(err)
		}
		defer mr.Close()

		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithMaxRetries(3),
			redis.WithRetryBackoff(time.Millisecond*100),
		)
		assert.NoError(t, err)
		defer client.Close()

		// 测试基本重试
		err = client.Set(ctx, "retry_key", "value", time.Minute)
		assert.NoError(t, err)

		val, err := client.Get(ctx, "retry_key")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)

		// 清理
		_, err = client.Del(ctx, "retry_key")
		assert.NoError(t, err)
	})

	t.Run("retry_exhausted", func(t *testing.T) {
		// 创建一个临时的 miniredis 实例并立即关闭它
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatal(err)
		}
		addr := mr.Addr()
		mr.Close() // 立即关闭服务器

		startTime := time.Now()

		// 使用已关闭的服务器地址
		client, err := redis.NewClient(
			redis.WithAddress(addr),
			redis.WithMaxRetries(2),
			redis.WithRetryBackoff(time.Millisecond*100),
			redis.WithConnTimeout(time.Millisecond*500),
			redis.WithDialTimeout(time.Millisecond*500),
		)

		if err != nil {
			duration := time.Since(startTime)
			// 验证错误
			assert.True(t, errors.HasErrorCode(err, codes.CacheError),
				"expected CacheError but got: %v", err)

			// 验证重试行为
			expectedMinDuration := time.Millisecond * 200 // 2次重试，每次100ms
			assert.True(t, duration >= expectedMinDuration,
				"operation should take at least %v, but took %v",
				expectedMinDuration, duration)
			return
		}

		defer func() {
			if client != nil {
				client.Close()
			}
		}()

		// 如果成功创建了客户端（不应该发生），则标记测试失败
		t.Fatal("expected client creation to fail")
	})

	t.Run("retry_with_timeout", func(t *testing.T) {
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatal(err)
		}
		defer mr.Close()

		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithMaxRetries(3),
			redis.WithRetryBackoff(time.Millisecond*100),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		// 使用短超时的上下文
		shortCtx, cancel := context.WithTimeout(ctx, time.Millisecond*50)
		defer cancel()

		// 设置一个会阻塞的命令
		mr.SetError("LOADING Redis is loading the dataset in memory")

		err = client.Set(shortCtx, "timeout_key", "value", time.Minute)
		assert.Error(t, err, "expected timeout error")
		assert.True(t, errors.HasErrorCode(err, codes.TimeoutError),
			"expected timeout error but got: %v", err)

		// 清理错误状态
		mr.SetError("")
	})

	t.Run("retry_with_pipeline", func(t *testing.T) {
		// 创建 miniredis 实例
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatal(err)
		}
		defer mr.Close()

		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithMaxRetries(3),
			redis.WithRetryBackoff(time.Millisecond*100),
		)
		assert.NoError(t, err)
		defer client.Close()

		pipe := client.TxPipeline()
		assert.NotNil(t, pipe)

		// 添加多个命令到管道
		err = pipe.Set(ctx, "pipe_retry_1", "value1", time.Minute)
		assert.NoError(t, err)
		err = pipe.Set(ctx, "pipe_retry_2", "value2", time.Minute)
		assert.NoError(t, err)

		// 执行管道
		cmds, err := pipe.Exec(ctx)
		assert.NoError(t, err)
		assert.Len(t, cmds, 2)

		// 清理
		_, err = client.Del(ctx, "pipe_retry_1", "pipe_retry_2")
		assert.NoError(t, err)
	})

	t.Run("retry_backoff", func(t *testing.T) {
		// 创建一个临时的 miniredis 实例并立即关闭它
		mr, err := miniredis.Run()
		if err != nil {
			t.Fatal(err)
		}
		addr := mr.Addr()
		mr.Close() // 立即关闭服务器

		startTime := time.Now()

		client, err := redis.NewClient(
			redis.WithAddress(addr),
			redis.WithMaxRetries(2),
			redis.WithRetryBackoff(time.Millisecond*100),
			redis.WithConnTimeout(time.Millisecond*500),
			redis.WithDialTimeout(time.Millisecond*500),
		)

		if err != nil {
			duration := time.Since(startTime)
			// 验证错误
			assert.True(t, errors.HasErrorCode(err, codes.CacheError),
				"expected CacheError but got: %v", err)

			// 验证重试间隔
			assert.True(t, duration >= time.Millisecond*200,
				"operation should take at least %v, but took %v",
				time.Millisecond*200, duration)
			return
		}

		defer func() {
			if client != nil {
				client.Close()
			}
		}()

		// 如果成功创建了客户端（不应该发生），则标记测试失败
		t.Fatal("expected client creation to fail")
	})
}
