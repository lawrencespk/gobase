package types

// Config Prometheus配置
type Config struct {
	// 是否启用Prometheus监控
	Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`

	// metrics接口端口
	Port int `mapstructure:"port" json:"port" yaml:"port"`

	// metrics接口路径
	Path string `mapstructure:"path" json:"path" yaml:"path"`

	// 指标命名空间
	Namespace string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`

	// 全局标签
	Labels map[string]string `mapstructure:"labels" json:"labels" yaml:"labels"`

	// 启用的收集器列表
	Collectors []string `mapstructure:"collectors" json:"collectors" yaml:"collectors"`

	// 采样配置
	Sampling struct {
		// 是否启用采样
		Enabled bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
		// 采样率 0.0-1.0
		Rate float64 `mapstructure:"rate" json:"rate" yaml:"rate"`
	} `mapstructure:"sampling" json:"sampling" yaml:"sampling"`

	// Prometheus配置
	Prometheus *PrometheusConfig `mapstructure:"prometheus" json:"prometheus" yaml:"prometheus"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Port    int    `mapstructure:"port" json:"port" yaml:"port"`
	Path    string `mapstructure:"path" json:"path" yaml:"path"`
}
