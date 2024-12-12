package integration

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/trace/jaeger"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
)

func TestJaegerIntegration(t *testing.T) {
	helper := setupTest(t)
	defer helper.cleanup()

	// 创建和测试追踪
	provider, err := jaeger.NewProvider()
	require.NoError(t, err)
	defer provider.Close()

	// 测试完整追踪流程
	t.Run("Full trace flow", func(t *testing.T) {
		rootSpan := provider.Tracer().StartSpan("root-operation")
		rootCtx := opentracing.ContextWithSpan(context.Background(), rootSpan)

		childSpan, _ := jaeger.StartSpanFromContext(rootCtx, "child-operation")
		childSpan.SetTag("custom.tag", "test-value")
		childSpan.LogKV("event", "test-event")

		time.Sleep(100 * time.Millisecond)
		childSpan.Finish()
		rootSpan.Finish()

		// 给 Jaeger 一些时间来处理 spans
		WaitForSpans(500 * time.Millisecond)
	})
}

func TestSamplingStrategies(t *testing.T) {
	helper := setupTest(t)
	defer helper.cleanup()

	tests := []struct {
		name       string
		validateFn func(*testing.T, *config.Config)
	}{
		{
			name: "Default constant sampler",
			validateFn: func(t *testing.T, cfg *config.Config) {
				require.Equal(t, "const", cfg.Jaeger.Sampler.Type)
				require.Equal(t, float64(1), cfg.Jaeger.Sampler.Param)
			},
		},
		// 可以添加更多采样策略测试用例，但使用配置文件中的值进行验证
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.GetConfig()
			require.NotNil(t, cfg)

			// 验证配置
			tt.validateFn(t, cfg)

			provider, err := jaeger.NewProvider()
			require.NoError(t, err)
			defer provider.Close()

			span := provider.Tracer().StartSpan("test-operation")
			span.Finish()

			WaitForSpans(500 * time.Millisecond)
		})
	}
}
