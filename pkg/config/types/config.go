package types

import "time"

// NacosConfig Nacos配置
type NacosConfig struct {
	Endpoints  []string `mapstructure:"endpoints"`   // 服务端点列表
	Namespace  string   `mapstructure:"namespace"`   // 命名空间
	Group      string   `mapstructure:"group"`       // 配置分组
	DataID     string   `mapstructure:"data_id"`     // 配置ID
	Username   string   `mapstructure:"username"`    // 用户名
	Password   string   `mapstructure:"password"`    // 密码
	TimeoutMs  int      `mapstructure:"timeout_ms"`  // 超时时间(毫秒)
	LogDir     string   `mapstructure:"log_dir"`     // 日志目录
	CacheDir   string   `mapstructure:"cache_dir"`   // 缓存目录
	LogLevel   string   `mapstructure:"log_level"`   // 日志级别
	Scheme     string   `mapstructure:"scheme"`      // 协议
	EnableAuth bool     `mapstructure:"enable_auth"` // 是否启用认证
	AccessKey  string   `mapstructure:"access_key"`  // 访问密钥
	SecretKey  string   `mapstructure:"secret_key"`  // 密钥
}

// ConfigOptions 配置选项
type ConfigOptions struct {
	ConfigFile  string       // 配置文件路径
	ConfigType  string       // 配置文件类型
	Environment string       // 环境
	EnableEnv   bool         // 是否启用环境变量
	EnvPrefix   string       // 环境变量前缀
	NacosConfig *NacosConfig // Nacos配置
}

// ConfigOption 配置选项函数
type ConfigOption func(*ConfigOptions)

// WithConfigFile 设置配置文件路径
func WithConfigFile(configFile string) ConfigOption {
	return func(o *ConfigOptions) {
		o.ConfigFile = configFile
	}
}

// WithConfigType 设置配置文件类型
func WithConfigType(configType string) ConfigOption {
	return func(o *ConfigOptions) {
		o.ConfigType = configType
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) ConfigOption {
	return func(o *ConfigOptions) {
		o.Environment = env
	}
}

// WithEnableEnv 设置是否启用环境变量
func WithEnableEnv(enable bool) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnableEnv = enable
	}
}

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnvPrefix = prefix
	}
}

// WithNacosConfig 设置Nacos配置
func WithNacosConfig(config *NacosConfig) ConfigOption {
	return func(o *ConfigOptions) {
		o.NacosConfig = config
	}
}

// WithNacosEndpoint 设置Nacos服务端点
func WithNacosEndpoint(endpoint string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		if o.NacosConfig.Endpoints == nil {
			o.NacosConfig.Endpoints = make([]string, 0, 1)
		}
		o.NacosConfig.Endpoints = append(o.NacosConfig.Endpoints, endpoint)
	}
}

// WithNacosNamespace 设置Nacos命名空间
func WithNacosNamespace(namespace string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.Namespace = namespace
	}
}

// WithNacosGroup 设置Nacos分组
func WithNacosGroup(group string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.Group = group
	}
}

// WithNacosDataID 设置Nacos数据ID
func WithNacosDataID(dataID string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.DataID = dataID
	}
}

// WithNacosAuth 设置Nacos认证信息
func WithNacosAuth(username, password string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.Username = username
		o.NacosConfig.Password = password
	}
}

// WithNacosTimeout 设置Nacos超时时间
func WithNacosTimeout(timeout time.Duration) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.TimeoutMs = int(timeout.Milliseconds())
	}
}

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	Enable      bool                  `mapstructure:"enable" json:"enable" yaml:"enable"`                   // 是否启用
	ServiceName string                `mapstructure:"service_name" json:"service_name" yaml:"service_name"` // 服务名称
	Agent       JaegerAgentConfig     `mapstructure:"agent" json:"agent" yaml:"agent"`                      // Agent配置
	Collector   JaegerCollectorConfig `mapstructure:"collector" json:"collector" yaml:"collector"`          // Collector配置
	Sampler     JaegerSamplerConfig   `mapstructure:"sampler" json:"sampler" yaml:"sampler"`                // 采样配置
	Tags        map[string]string     `mapstructure:"tags" json:"tags" yaml:"tags"`                         // 标签
	Buffer      JaegerBufferConfig    `mapstructure:"buffer" json:"buffer" yaml:"buffer"`                   // 缓冲区配置
}

// JaegerAgentConfig Jaeger Agent配置
type JaegerAgentConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"` // 主机
	Port string `mapstructure:"port" json:"port" yaml:"port"` // 端口
}

// JaegerCollectorConfig Jaeger Collector配置
type JaegerCollectorConfig struct {
	Endpoint string        `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"` // 端点
	Username string        `mapstructure:"username" json:"username" yaml:"username"` // 用户名
	Password string        `mapstructure:"password" json:"password" yaml:"password"` // 密码
	Timeout  time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`    // 超时时间
}

// JaegerSamplerConfig 采样配置
type JaegerSamplerConfig struct {
	Type            string  `mapstructure:"type" json:"type" yaml:"type"`                                     // 类型
	Param           float64 `mapstructure:"param" json:"param" yaml:"param"`                                  // 参数
	ServerURL       string  `mapstructure:"server_url" json:"server_url" yaml:"server_url"`                   // 服务器URL
	MaxOperations   int     `mapstructure:"max_operations" json:"max_operations" yaml:"max_operations"`       // 最大操作数
	RefreshInterval int     `mapstructure:"refresh_interval" json:"refresh_interval" yaml:"refresh_interval"` // 刷新间隔
	RateLimit       float64 `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`                   // 速率限制
	Adaptive        bool    `mapstructure:"adaptive" json:"adaptive" yaml:"adaptive"`                         // 是否自适应
}

// JaegerBufferConfig 缓冲区配置
type JaegerBufferConfig struct {
	Enable        bool          `mapstructure:"enable" json:"enable" yaml:"enable"`                         // 是否启用
	Size          int           `mapstructure:"size" json:"size" yaml:"size"`                               // 大小
	FlushInterval time.Duration `mapstructure:"flush_interval" json:"flush_interval" yaml:"flush_interval"` // 刷新间隔
}

type GrafanaConfig struct {
	Dashboards struct {
		HTTP    string `json:"http" yaml:"http"`       // HTTP仪表盘配置
		Logger  string `json:"logger" yaml:"logger"`   // 日志仪表盘配置
		Runtime string `json:"runtime" yaml:"runtime"` // 运行时仪表盘配置
		System  string `json:"system" yaml:"system"`   // 系统仪表盘配置
	} `json:"dashboards" yaml:"dashboards"`

	Alerts struct {
		Rules  string `json:"rules" yaml:"rules"`   // 通用告警规则
		Logger string `json:"logger" yaml:"logger"` // 日志告警规则
	} `json:"alerts" yaml:"alerts"`
}

type Config struct {
	Jaeger  JaegerConfig  `json:"jaeger" yaml:"jaeger" mapstructure:"jaeger"`    // Jaeger配置
	Grafana GrafanaConfig `json:"grafana" yaml:"grafana" mapstructure:"grafana"` // Grafana配置
}
