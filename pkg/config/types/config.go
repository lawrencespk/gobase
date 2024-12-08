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
