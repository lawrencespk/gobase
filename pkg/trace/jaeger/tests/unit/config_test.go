package unit

import (
	"testing"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/config/types"
	"gobase/pkg/trace/jaeger"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *jaeger.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &jaeger.Config{
				Enable:      true,
				ServiceName: "test-service",
				Agent: jaeger.AgentConfig{
					Host: "localhost",
					Port: "6831",
				},
				Collector: jaeger.CollectorConfig{
					Endpoint: "http://localhost:14268/api/traces",
					Timeout:  5 * time.Second,
				},
				Sampler: jaeger.SamplerConfig{
					Type:  "const",
					Param: 1,
				},
				Buffer: jaeger.BufferConfig{
					Enable:        true,
					Size:          1000,
					FlushInterval: 1 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: &jaeger.Config{
				Enable: true,
				Agent: jaeger.AgentConfig{
					Host: "localhost",
					Port: "6831",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sampler type",
			config: &jaeger.Config{
				Enable:      true,
				ServiceName: "test-service",
				Agent: jaeger.AgentConfig{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: jaeger.SamplerConfig{
					Type:  "invalid",
					Param: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid probabilistic param",
			config: &jaeger.Config{
				Enable:      true,
				ServiceName: "test-service",
				Agent: jaeger.AgentConfig{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: jaeger.SamplerConfig{
					Type:  "probabilistic",
					Param: 2.0, // 超出有效范围
				},
			},
			wantErr: true,
		},
		{
			name: "disabled config",
			config: &jaeger.Config{
				Enable:      false,
				ServiceName: "test-service",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jaeger.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// 设置测试配置
	testConfig := &config.Config{
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
	}
	config.SetConfig(testConfig)

	// 测试加载配置
	cfg, err := jaeger.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 验证加载的配置
	if cfg.ServiceName != "test-service" {
		t.Errorf("LoadConfig() ServiceName = %v, want %v", cfg.ServiceName, "test-service")
	}
	if cfg.Agent.Host != "localhost" {
		t.Errorf("LoadConfig() Agent.Host = %v, want %v", cfg.Agent.Host, "localhost")
	}
}
