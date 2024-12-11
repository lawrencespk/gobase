package logger

import (
	"errors"
	"gobase/pkg/logger/types"
	"math/rand/v2"
	"time"

	"github.com/gin-gonic/gin"
)

// Options 中间件选项
type Options struct {
	// 配置
	Config *Config

	// 日志实例
	Logger types.Logger

	// 格式化器
	Formatter LogFormatter

	// 采样器
	Sampler Sampler

	// 指标收集器(为Prometheus准备)
	MetricsCollector MetricsCollector

	// 追踪器(为Jaeger准备)
	Tracer Tracer
}

// Option 选项函数
type Option func(*Options)

// DefaultOptions 返回默认选项
func DefaultOptions() *Options {
	return &Options{
		Config:    DefaultConfig(),
		Logger:    nil, // 需要外部设置
		Formatter: NewDefaultFormatter(),
		Sampler:   NewDefaultSampler(DefaultConfig().SampleRate),
	}
}

// WithConfig 设置配置
func WithConfig(cfg *Config) Option {
	return func(o *Options) {
		o.Config = cfg
	}
}

// WithLogger 设置日志实例
func WithLogger(logger types.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithFormatter 设置格式化器
func WithFormatter(formatter LogFormatter) Option {
	return func(o *Options) {
		o.Formatter = formatter
	}
}

// WithSampler 设置采样器
func WithSampler(sampler Sampler) Option {
	return func(o *Options) {
		o.Sampler = sampler
	}
}

// WithMetricsCollector 设置指标收集器
func WithMetricsCollector(collector MetricsCollector) Option {
	return func(o *Options) {
		o.MetricsCollector = collector
	}
}

// WithTracer 设置追踪器
func WithTracer(tracer Tracer) Option {
	return func(o *Options) {
		o.Tracer = tracer
	}
}

// defaultSampler 默认采样器
type defaultSampler struct {
	rate float64
}

// NewDefaultSampler 创建默认采样器
func NewDefaultSampler(rate float64) Sampler {
	return &defaultSampler{rate: rate}
}

// Sample 实现采样接口
func (s *defaultSampler) Sample(ctx *gin.Context) bool {
	return rand.Float64() < s.rate
}

// defaultFormatter 默认格式化器
type defaultFormatter struct{}

// NewDefaultFormatter 创建默认格式化器
func NewDefaultFormatter() LogFormatter {
	return &defaultFormatter{}
}

// Format 实现格式化接口
func (f *defaultFormatter) Format(c *gin.Context, param *LogFormatterParam) map[string]interface{} {
	data := map[string]interface{}{
		"time":       param.StartTime.Format(time.RFC3339),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"ip":         c.ClientIP(),
		"user-agent": c.Request.UserAgent(),
		"latency":    param.Latency.Milliseconds(),
		"status":     param.StatusCode,
		"req_size":   param.RequestSize,
		"resp_size":  param.ResponseSize,
	}

	// 添加请求ID
	if requestID := c.GetString("request_id"); requestID != "" {
		data["request_id"] = requestID
	}

	// 添加错误信息
	if param.Error != nil {
		data["error"] = param.Error.Error()
	}

	// 添加额外字段
	for k, v := range param.Fields {
		data[k] = v
	}

	return data
}

// Validate 验证选项
func (o *Options) Validate() error {
	if o.Logger == nil {
		return errors.New("logger is required")
	}

	if o.Config == nil {
		return errors.New("config is required")
	}

	if err := o.Config.Validate(); err != nil {
		return err
	}

	if o.Formatter == nil {
		return errors.New("formatter is required")
	}

	if o.Sampler == nil {
		return errors.New("sampler is required")
	}

	return nil
}
