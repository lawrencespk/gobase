package unit

import (
	"testing"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/config/types"
	"gobase/pkg/trace/jaeger"

	"github.com/opentracing/opentracing-go"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantError bool
	}{
		{
			name: "valid config with jaeger enabled",
			config: &config.Config{
				Jaeger: types.JaegerConfig{
					Enable:      true,
					ServiceName: "test-service",
					Agent: types.JaegerAgentConfig{
						Host: "localhost",
						Port: "6831",
					},
					Collector: types.JaegerCollectorConfig{
						Endpoint: "http://localhost:14268/api/traces",
						Timeout:  5 * time.Second,
					},
					Sampler: types.JaegerSamplerConfig{
						Type:  "const",
						Param: 1,
					},
					Buffer: types.JaegerBufferConfig{
						Enable:        true,
						Size:          1000,
						FlushInterval: time.Second,
					},
					Tags: map[string]string{
						"env": "test",
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid config with jaeger disabled",
			config: &config.Config{
				Jaeger: types.JaegerConfig{
					Enable:      false,
					ServiceName: "test-service",
				},
			},
			wantError: false,
		},
		{
			name: "invalid config - missing service name",
			config: &config.Config{
				Jaeger: types.JaegerConfig{
					Enable: true,
					Agent: types.JaegerAgentConfig{
						Host: "localhost",
						Port: "6831",
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置测试配置
			config.SetConfig(tt.config)

			// 创建provider
			provider, err := jaeger.NewProvider()
			if (err != nil) != tt.wantError {
				t.Errorf("NewProvider() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil {
				return
			}

			// 验证provider
			if provider == nil {
				t.Error("NewProvider() returned nil provider")
				return
			}

			// 验证tracer
			tracer := provider.Tracer()
			if tt.config.Jaeger.Enable {
				if _, ok := tracer.(opentracing.NoopTracer); ok {
					t.Error("Provider.Tracer() returned NoopTracer when Jaeger is enabled")
				}
			} else {
				if _, ok := tracer.(opentracing.NoopTracer); !ok {
					t.Error("Provider.Tracer() did not return NoopTracer when Jaeger is disabled")
				}
			}

			// 测试关闭
			if err := provider.Close(); err != nil {
				t.Errorf("Provider.Close() error = %v", err)
			}
		})
	}
}

func TestProviderGlobalTracer(t *testing.T) {
	// 设置测试配置
	config.SetConfig(&config.Config{
		Jaeger: types.JaegerConfig{
			Enable:      true,
			ServiceName: "test-service",
			Agent: types.JaegerAgentConfig{
				Host: "localhost",
				Port: "6831",
			},
			Collector: types.JaegerCollectorConfig{
				Endpoint: "http://localhost:14268/api/traces",
				Timeout:  5 * time.Second,
			},
			Sampler: types.JaegerSamplerConfig{
				Type:  "const",
				Param: 1,
			},
		},
	})

	// 创建provider
	provider, err := jaeger.NewProvider()
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	defer provider.Close()

	// 验证全局tracer是否被设置
	globalTracer := opentracing.GlobalTracer()
	if globalTracer == nil {
		t.Error("Global tracer not set")
	}

	// 验证是否可以创建span
	span := globalTracer.StartSpan("test-operation")
	if span == nil {
		t.Error("Failed to create span from global tracer")
	}
	span.Finish()
}

func TestProviderWithTags(t *testing.T) {
	// 设置带标签的测试配置
	testTags := map[string]string{
		"env":     "test",
		"version": "1.0.0",
	}

	config.SetConfig(&config.Config{
		Jaeger: types.JaegerConfig{
			Enable:      true,
			ServiceName: "test-service",
			Agent: types.JaegerAgentConfig{
				Host: "localhost",
				Port: "6831",
			},
			Collector: types.JaegerCollectorConfig{
				Endpoint: "http://localhost:14268/api/traces",
				Timeout:  5 * time.Second,
			},
			Sampler: types.JaegerSamplerConfig{
				Type:  "const",
				Param: 1,
			},
			Tags: testTags,
		},
	})

	// 创建provider
	provider, err := jaeger.NewProvider()
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	defer provider.Close()

	// 创建span
	span := provider.Tracer().StartSpan("test-operation")
	if span == nil {
		t.Error("Failed to create span")
	}
	span.Finish()

	// 由于我们无法直接访问span的标签，我们只能验证provider是否成功创建
	if provider.Tracer() == nil {
		t.Error("Provider.Tracer() returned nil")
	}
}
