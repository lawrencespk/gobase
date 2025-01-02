package unit

import (
	"testing"
	"time"

	"gobase/pkg/client/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	// 创建 miniredis 实例
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	t.Run("default pool settings", func(t *testing.T) {
		opts := redis.DefaultOptions()
		assert.Equal(t, 10, opts.PoolSize)
		assert.Equal(t, 0, opts.MinIdleConns)
		assert.Equal(t, 5*time.Minute, opts.IdleTimeout)
	})

	t.Run("custom pool settings", func(t *testing.T) {
		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithPoolSize(20),
			redis.WithMinIdleConns(5),
			redis.WithIdleTimeout(time.Minute),
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// 等待连接池初始化完成
		maxRetries := 5
		var stats *redis.PoolStats
		for i := 0; i < maxRetries; i++ {
			stats = client.PoolStats()
			if stats.TotalConns == uint32(20) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 验证连接池设置
		assert.Equal(t, uint32(20), stats.TotalConns)
		assert.GreaterOrEqual(t, stats.IdleConns, uint32(5))
	})

	t.Run("pool timeouts", func(t *testing.T) {
		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithDialTimeout(time.Second),
			redis.WithReadTimeout(2*time.Second),
			redis.WithWriteTimeout(2*time.Second),
			redis.WithIdleTimeout(time.Minute),
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// 验证连接池是否正常工作
		stats := client.PoolStats()
		assert.Greater(t, stats.TotalConns, uint32(0))
	})

	t.Run("invalid pool settings", func(t *testing.T) {
		// 测试负数连接池大小
		_, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithPoolSize(-1),
		)
		assert.Error(t, err)

		// 测试负数最小空闲连接
		_, err = redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithMinIdleConns(-1),
		)
		assert.Error(t, err)

		// 测试零值超时
		_, err = redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithIdleTimeout(0),
		)
		assert.Error(t, err)
	})
}
