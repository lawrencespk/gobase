package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestTracing(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)

	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	client, err := redis.NewClient(
		redis.WithAddress(addr),
		redis.WithTracer(tracer),
	)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("basic operation tracing", func(t *testing.T) {
		tracer.Reset()

		err := client.Set(ctx, "trace_key", "value", time.Minute)
		assert.NoError(t, err)

		spans := tracer.FinishedSpans()
		assert.Len(t, spans, 1)
		assert.Equal(t, "redis.Set", spans[0].OperationName)
	})

	t.Run("pipeline tracing", func(t *testing.T) {
		tracer.Reset()

		pipe := client.TxPipeline()
		err := pipe.Set(ctx, "trace_pipe_key", "value", time.Minute)
		assert.NoError(t, err)
		_, err = pipe.Exec(ctx)
		assert.NoError(t, err)

		spans := tracer.FinishedSpans()
		assert.GreaterOrEqual(t, len(spans), 1)
	})

	t.Run("error tracing", func(t *testing.T) {
		tracer.Reset()

		_, err := client.Get(ctx, "non_existent_key")
		assert.Error(t, err)

		spans := tracer.FinishedSpans()
		assert.Len(t, spans, 1)
		assert.Equal(t, "redis.Get", spans[0].OperationName)
	})
}
