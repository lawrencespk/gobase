package logger

import (
	"errors"

	"github.com/spf13/viper"
)

const (
	// DefaultLevel 默认日志级别
	DefaultLevel = "info"
	// DefaultFormat 默认日志格式
	DefaultFormat = "json"
	// DefaultSampleRate 默认采样率
	DefaultSampleRate = 1.0
	// DefaultSlowThreshold 默认慢请求阈值(毫秒)
	DefaultSlowThreshold = 200
	// DefaultRequestBodyLimit 默认请求体记录阈值(字节)
	DefaultRequestBodyLimit = 1024
	// DefaultResponseBodyLimit 默认响应体记录阈值(字节)
	DefaultResponseBodyLimit = 1024
)

// Config 日志中间件配置
type Config struct {
	// 是否启用
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`

	// 日志级别: debug, info, warn, error, fatal
	Level string `mapstructure:"level" json:"level" yaml:"level"`

	// 日志格式: text, json
	Format string `mapstructure:"format" json:"format" yaml:"format"`

	// 采样率 0.0-1.0
	SampleRate float64 `mapstructure:"sample_rate" json:"sample_rate" yaml:"sample_rate"`

	// 慢请求阈值(毫秒)
	SlowThreshold int64 `mapstructure:"slow_threshold" json:"slow_threshold" yaml:"slow_threshold"`

	// 请求体记录阈值(字节)
	RequestBodyLimit int64 `mapstructure:"request_body_limit" json:"request_body_limit" yaml:"request_body_limit"`

	// 响应体记录阈值(字节)
	ResponseBodyLimit int64 `mapstructure:"response_body_limit" json:"response_body_limit" yaml:"response_body_limit"`

	// 跳过的路径
	SkipPaths []string `mapstructure:"skip_paths" json:"skip_paths" yaml:"skip_paths"`

	// 指标配置(为Prometheus准备)
	Metrics MetricsConfig `mapstructure:"metrics" json:"metrics" yaml:"metrics"`

	// 追踪配置(为Jaeger准备)
	Trace TraceConfig `mapstructure:"trace" json:"trace" yaml:"trace"`

	// 日志轮转配置
	Rotate RotateConfig `mapstructure:"rotate" json:"rotate" yaml:"rotate"`

	// 缓冲配置
	Buffer BufferConfig `mapstructure:"buffer" json:"buffer" yaml:"buffer"`
}

// MetricsConfig Prometheus指标配置
type MetricsConfig struct {
	// 是否启用指标收集
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`

	// 指标前缀
	Prefix string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`

	// 标签
	Labels map[string]string `mapstructure:"labels" json:"labels" yaml:"labels"`

	// 是否收集请求时长直方图
	EnableLatencyHistogram bool `mapstructure:"enable_latency_histogram" json:"enable_latency_histogram" yaml:"enable_latency_histogram"`

	// 是否收集请求大小直方图
	EnableSizeHistogram bool `mapstructure:"enable_size_histogram" json:"enable_size_histogram" yaml:"enable_size_histogram"`

	// 直方图桶
	Buckets []float64 `mapstructure:"buckets" json:"buckets" yaml:"buckets"`
}

// TraceConfig Jaeger追踪配置
type TraceConfig struct {
	// 是否启用追踪
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`

	// 采样类型: const, probabilistic, rateLimiting, remote
	SamplerType string `mapstructure:"sampler_type" json:"sampler_type" yaml:"sampler_type"`

	// 采样参数
	SamplerParam float64 `mapstructure:"sampler_param" json:"sampler_param" yaml:"sampler_param"`

	// 标签
	Tags map[string]string `mapstructure:"tags" json:"tags" yaml:"tags"`
}

// RotateConfig 日志轮转配置
type RotateConfig struct {
	// 是否启用日志轮转
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`

	// 单个文件最大尺寸(MB)
	MaxSize int `mapstructure:"max_size" json:"max_size" yaml:"max_size"`

	// 文件保留天数
	MaxAge int `mapstructure:"max_age" json:"max_age" yaml:"max_age"`

	// 最大备份数
	MaxBackups int `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`

	// 是否压缩
	Compress bool `mapstructure:"compress" json:"compress" yaml:"compress"`

	// 日志文件路径
	FilePath string `mapstructure:"file_path" json:"file_path" yaml:"file_path"`
}

// BufferConfig 缓冲配置
type BufferConfig struct {
	// 是否启用缓冲
	Enable bool `mapstructure:"enable" json:"enable" yaml:"enable"`

	// 缓冲区大小(字节)
	Size int `mapstructure:"size" json:"size" yaml:"size"`

	// 刷新超时时间(毫秒)
	FlushInterval int `mapstructure:"flush_interval" json:"flush_interval" yaml:"flush_interval"`

	// 是否在错误时立即刷新
	FlushOnError bool `mapstructure:"flush_on_error" json:"flush_on_error" yaml:"flush_on_error"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Enable:            true,
		Level:             DefaultLevel,
		Format:            DefaultFormat,
		SampleRate:        DefaultSampleRate,
		SlowThreshold:     DefaultSlowThreshold,
		RequestBodyLimit:  DefaultRequestBodyLimit,
		ResponseBodyLimit: DefaultResponseBodyLimit,
		SkipPaths:         []string{"/health", "/metrics"},
		Metrics: MetricsConfig{
			Enable:                 true,
			Prefix:                 "http_request",
			EnableLatencyHistogram: true,
			EnableSizeHistogram:    true,
			Buckets:                []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		Trace: TraceConfig{
			Enable:       true,
			SamplerType:  "const",
			SamplerParam: 1,
		},
		Rotate: RotateConfig{
			Enable:     true,
			MaxSize:    100, // 100MB
			MaxAge:     7,   // 7天
			MaxBackups: 5,   // 保留5个备份
			Compress:   true,
			FilePath:   "./logs/app.log",
		},
		Buffer: BufferConfig{
			Enable:        true,
			Size:          4096, // 4KB
			FlushInterval: 1000, // 1秒
			FlushOnError:  true,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.SampleRate < 0 || c.SampleRate > 1 {
		return errors.New("sample_rate must be between 0 and 1")
	}

	if c.SlowThreshold < 0 {
		return errors.New("slow_threshold must be positive")
	}

	if c.RequestBodyLimit < 0 {
		return errors.New("request_body_limit must be positive")
	}

	if c.ResponseBodyLimit < 0 {
		return errors.New("response_body_limit must be positive")
	}

	// 验证轮转配置
	if c.Rotate.Enable {
		if c.Rotate.MaxSize <= 0 {
			return errors.New("rotate max_size must be positive")
		}
		if c.Rotate.MaxAge <= 0 {
			return errors.New("rotate max_age must be positive")
		}
		if c.Rotate.MaxBackups <= 0 {
			return errors.New("rotate max_backups must be positive")
		}
		if c.Rotate.FilePath == "" {
			return errors.New("rotate file_path is required")
		}
	}

	// 验证缓冲配置
	if c.Buffer.Enable {
		if c.Buffer.Size <= 0 {
			return errors.New("buffer size must be positive")
		}
		if c.Buffer.FlushInterval <= 0 {
			return errors.New("buffer flush_interval must be positive")
		}
	}

	return nil
}

// LoadConfig 从viper加载配置
func LoadConfig(v *viper.Viper) (*Config, error) {
	config := DefaultConfig()

	if err := v.UnmarshalKey("middleware.logger", config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}
