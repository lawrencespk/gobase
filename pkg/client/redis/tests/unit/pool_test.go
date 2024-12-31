package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	// 创建 miniredis 实例
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	ctx := context.Background()

	t.Run("pool stats", func(t *testing.T) {
		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()), // 使用 miniredis 地址
			redis.WithPoolSize(10),
			redis.WithMinIdleConns(2),
		)
		assert.NoError(t, err)
		defer client.Close()

		// 执行一些操作来生成连接
		for i := 0; i < 5; i++ {
			err := client.Set(ctx, "key"+strconv.Itoa(i), "value", time.Minute)
			assert.NoError(t, err)
		}

		// 检查连接池统计信息
		stats := client.Pool().Stats()
		assert.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalConns, uint32(1))
		assert.GreaterOrEqual(t, stats.IdleConns, uint32(1))
	})

	t.Run("pool timeout", func(t *testing.T) {
		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithPoolSize(1),                   // 设置连接池大小为1
			redis.WithPoolTimeout(time.Millisecond), // 设置非常短的超时时间
			redis.WithMaxRetries(0),                 // 禁用重试
			redis.WithRetryBackoff(0),               // 禁用重试间隔
		)
		assert.NoError(t, err)
		defer client.Close()

		// 创建一个阻塞的连接
		done := make(chan struct{})
		go func() {
			defer close(done)
			// 使用一个长时间运行的操作来占用连接
			script := `
				local start = redis.call('TIME')
				redis.call('SET', KEYS[1], ARGV[1])
				while (redis.call('TIME')[1] - start[1]) < 1 do end
				return 1
			`
			_, err := client.Eval(ctx, script, []string{"blocking_key"}, "value")
			assert.NoError(t, err)
		}()

		// 等待一小段时间确保第一个操作开始
		time.Sleep(time.Millisecond * 10)

		// 尝试在池已满时获取连接
		err = client.Set(ctx, "timeout_key", "value", time.Minute)
		assert.Error(t, err)
		<-done
	})
}
