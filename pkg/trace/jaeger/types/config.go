package types

import "time"

// Config Jaeger配置
type Config struct {
	Enable      bool              `mapstructure:"enable" json:"enable" yaml:"enable"`
	ServiceName string            `mapstructure:"service_name" json:"service_name" yaml:"service_name"`
	Agent       AgentConfig       `mapstructure:"agent" json:"agent" yaml:"agent"`
	Collector   CollectorConfig   `mapstructure:"collector" json:"collector" yaml:"collector"`
	Sampler     SamplerConfig     `mapstructure:"sampler" json:"sampler" yaml:"sampler"`
	Tags        map[string]string `mapstructure:"tags" json:"tags" yaml:"tags"`
	Buffer      BufferConfig      `mapstructure:"buffer" json:"buffer" yaml:"buffer"`
}

// AgentConfig Jaeger Agent配置
type AgentConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
	Port string `mapstructure:"port" json:"port" yaml:"port"`
}

// CollectorConfig Jaeger Collector配置
type CollectorConfig struct {
	Endpoint string        `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	Username string        `mapstructure:"username" json:"username" yaml:"username"`
	Password string        `mapstructure:"password" json:"password" yaml:"password"`
	Timeout  time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// SamplerConfig 采样配置
type SamplerConfig struct {
	Type            string  `mapstructure:"type" json:"type" yaml:"type"`
	Param           float64 `mapstructure:"param" json:"param" yaml:"param"`
	ServerURL       string  `mapstructure:"server_url" json:"server_url" yaml:"server_url"`
	MaxOperations   int     `mapstructure:"max_operations" json:"max_operations" yaml:"max_operations"`
	RefreshInterval int     `mapstructure:"refresh_interval" json:"refresh_interval" yaml:"refresh_interval"`
	RateLimit       float64 `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
	Adaptive        bool    `mapstructure:"adaptive" json:"adaptive" yaml:"adaptive"`
}

// BufferConfig 缓冲区配置
type BufferConfig struct {
	Enable        bool          `mapstructure:"enable" json:"enable" yaml:"enable"`
	Size          int           `mapstructure:"size" json:"size" yaml:"size"`
	FlushInterval time.Duration `mapstructure:"flush_interval" json:"flush_interval" yaml:"flush_interval"`
}
