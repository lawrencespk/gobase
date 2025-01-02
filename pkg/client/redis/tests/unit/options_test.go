package unit

import (
	"gobase/pkg/client/redis"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := redis.DefaultOptions()
		assert.NotNil(t, opts)
		assert.Equal(t, 10, opts.PoolSize)
		assert.Equal(t, 0, opts.DB)
		assert.Equal(t, 3, opts.MaxRetries)
	})

	t.Run("with address", func(t *testing.T) {
		opts := redis.DefaultOptions()
		redis.WithAddress("localhost:6379")(opts)
		assert.Equal(t, []string{"localhost:6379"}, opts.Addresses)
	})

	t.Run("with credentials", func(t *testing.T) {
		opts := redis.DefaultOptions()
		redis.WithUsername("user")(opts)
		redis.WithPassword("pass")(opts)
		assert.Equal(t, "user", opts.Username)
		assert.Equal(t, "pass", opts.Password)
	})

	t.Run("with pool config", func(t *testing.T) {
		opts := redis.DefaultOptions()
		redis.WithPoolSize(20)(opts)
		redis.WithMinIdleConns(5)(opts)
		assert.Equal(t, 20, opts.PoolSize)
		assert.Equal(t, 5, opts.MinIdleConns)
	})

	t.Run("with timeouts", func(t *testing.T) {
		opts := redis.DefaultOptions()
		redis.WithDialTimeout(time.Second)(opts)
		redis.WithReadTimeout(time.Second)(opts)
		redis.WithWriteTimeout(time.Second)(opts)
		assert.Equal(t, time.Second, opts.DialTimeout)
		assert.Equal(t, time.Second, opts.ReadTimeout)
		assert.Equal(t, time.Second, opts.WriteTimeout)
	})

	t.Run("with features", func(t *testing.T) {
		opts := redis.DefaultOptions()
		redis.WithTLS(true)(opts)
		redis.WithCluster(true)(opts)
		redis.WithMetrics(true)(opts)
		redis.WithMetricsNamespace("test_metrics")(opts)

		assert.True(t, opts.EnableTLS)
		assert.True(t, opts.EnableCluster)
		assert.True(t, opts.EnableMetrics)
		assert.Equal(t, "test_metrics", opts.MetricsNamespace)
	})

	t.Run("metrics options", func(t *testing.T) {
		metrics := redis.NewRedisMetrics("test_redis")

		opts := redis.DefaultOptions()

		redis.WithEnableMetrics(true)(opts)
		assert.True(t, opts.EnableMetrics)

		redis.WithMetricsNamespace("test_namespace")(opts)
		assert.Equal(t, "test_namespace", opts.MetricsNamespace)

		redis.WithCollector(metrics)(opts)
		assert.NotNil(t, opts.Collector)
		assert.Equal(t, metrics, opts.Collector)
	})
}

func TestOptionsValidation(t *testing.T) {
	t.Run("metrics validation", func(t *testing.T) {
		// 创建 miniredis 实例
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		metrics := redis.NewRedisMetrics("test_redis")

		// 使用 miniredis 地址
		client, err := redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithEnableMetrics(true),
			redis.WithMetricsNamespace("test_namespace"),
			redis.WithCollector(metrics),
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// 测试不带 collector 的情况
		client, err = redis.NewClient(
			redis.WithAddress(mr.Addr()),
			redis.WithEnableMetrics(true),
		)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()
	})
}
