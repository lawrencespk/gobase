package unit

import (
	"gobase/pkg/client/redis"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
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
		metricsPrefix := "test_metrics"
		opts := redis.DefaultOptions()
		redis.WithMetrics(true)(opts)
		redis.WithMetricsNamespace(metricsPrefix)(opts)

		assert.True(t, opts.EnableMetrics)
		assert.Equal(t, metricsPrefix, opts.MetricsNamespace)
	})
}

func TestMetricsOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	registry := prometheus.NewRegistry()
	metricsPrefix := "test_redis_metrics_" + time.Now().Format("150405")

	client, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithMetrics(true),
		redis.WithMetricsNamespace(metricsPrefix),
		redis.WithRegistry(registry),
	)

	if err == nil {
		defer client.Close()
		metrics, err := registry.Gather()
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics)
	}
}
