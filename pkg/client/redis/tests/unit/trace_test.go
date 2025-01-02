package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/config"
	"gobase/pkg/config/types"
	"gobase/pkg/trace/jaeger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracing(t *testing.T) {
	// 创建配置
	cfg := config.NewConfig()
	cfg.Jaeger = types.JaegerConfig{
		Enable:      true,
		ServiceName: "redis-test",
		Sampler: types.JaegerSamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Agent: types.JaegerAgentConfig{
			Host: "localhost",
			Port: "6831",
		},
		Collector: types.JaegerCollectorConfig{
			Endpoint: "http://localhost:14268/api/traces",
			Username: "",
			Password: "",
			Timeout:  30,
		},
		Buffer: types.JaegerBufferConfig{
			Size:          1000,
			FlushInterval: 1,
		},
	}

	// 设置全局配置
	config.SetConfig(cfg)

	// 创建测试用的 jaeger provider
	provider, err := jaeger.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	if provider == nil {
		t.Fatal("provider should not be nil")
	}
	defer func() {
		if provider != nil {
			provider.Close()
		}
	}()

	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	client, err := redis.NewClient(
		redis.WithAddress(addr),
		redis.WithTracer(provider),
	)
	assert.NoError(t, err)
	defer client.Close()

	t.Run("basic operation tracing", func(t *testing.T) {
		ctx := context.Background()

		err := client.Set(ctx, "trace_key", "value", time.Minute)
		assert.NoError(t, err)

		val, err := client.Get(ctx, "trace_key")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("pipeline tracing", func(t *testing.T) {
		ctx := context.Background()

		pipe := client.TxPipeline()
		err := pipe.Set(ctx, "trace_pipe_key", "value", time.Minute)
		assert.NoError(t, err)
		_, err = pipe.Exec(ctx)
		assert.NoError(t, err)

		val, err := client.Get(ctx, "trace_pipe_key")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("error tracing", func(t *testing.T) {
		ctx := context.Background()

		_, err := client.Get(ctx, "non_existent_key")
		assert.Error(t, err)
	})
}
