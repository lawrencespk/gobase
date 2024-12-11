package jaeger

import (
	"fmt"
	"io"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

// Provider Jaeger提供者
type Provider struct {
	config  *Config
	tracer  opentracing.Tracer
	closer  io.Closer
	metrics metrics.Factory
}

// NewProvider 创建Jaeger提供者
func NewProvider() (*Provider, error) {
	// 获取配置
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if !config.Enable {
		return &Provider{config: config}, nil
	}

	// 将 map[string]string 转换为 []opentracing.Tag
	tags := make([]opentracing.Tag, 0, len(config.Tags))
	for k, v := range config.Tags {
		tags = append(tags, opentracing.Tag{Key: k, Value: v})
	}

	// 创建Jaeger配置
	cfg := jaegercfg.Configuration{
		ServiceName: config.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:                    config.Sampler.Type,
			Param:                   config.Sampler.Param,
			SamplingServerURL:       config.Sampler.ServerURL,
			MaxOperations:           config.Sampler.MaxOperations,
			SamplingRefreshInterval: time.Duration(config.Sampler.RefreshInterval) * time.Second,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			LocalAgentHostPort:  fmt.Sprintf("%s:%s", config.Agent.Host, config.Agent.Port),
			CollectorEndpoint:   config.Collector.Endpoint,
			User:                config.Collector.Username,
			Password:            config.Collector.Password,
			QueueSize:           config.Buffer.Size,
			BufferFlushInterval: config.Buffer.FlushInterval,
		},
		Tags: tags,
	}

	// 创建metrics工厂
	metrics := metrics.NullFactory

	// 创建tracer
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Metrics(metrics),
		jaegercfg.Logger(jaeger.StdLogger),
	)
	if err != nil {
		return nil, errors.NewThirdPartyError(fmt.Sprintf("[%s] failed to create tracer", codes.JaegerInitError), err)
	}

	// 设置全局tracer
	opentracing.SetGlobalTracer(tracer)

	return &Provider{
		config:  config,
		tracer:  tracer,
		closer:  closer,
		metrics: metrics,
	}, nil
}

// Tracer 获取tracer实例
func (p *Provider) Tracer() opentracing.Tracer {
	if !p.config.Enable {
		return opentracing.NoopTracer{}
	}
	return p.tracer
}

// Close 关闭provider
func (p *Provider) Close() error {
	if !p.config.Enable || p.closer == nil {
		return nil
	}

	if err := p.closer.Close(); err != nil {
		return errors.NewThirdPartyError(fmt.Sprintf("[%s] failed to close provider", codes.JaegerShutdownError), err)
	}
	return nil
}
