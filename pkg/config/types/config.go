package types

import "time"

// NacosConfig Nacos配置
type NacosConfig struct {
	Endpoint    string        // 服务端点
	NamespaceID string        // 命名空间ID
	Group       string        // 配置分组
	DataID      string        // 配置ID
	Username    string        // 用户名
	Password    string        // 密码
	Timeout     time.Duration // 超时时间
	LogDir      string        // 日志目录
	CacheDir    string        // 缓存目录
	LogLevel    string        // 日志级别
	// 认证相关
	AccessKey   string // AccessKey
	SecretKey   string // SecretKey
	EnableAuth  bool   // 是否启用认证
	AuthToken   string // 认证Token
	IdentityKey string // 身份Key
	IdentityVal string // 身份Value
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
		o.NacosConfig.Endpoint = endpoint
	}
}

// WithNacosNamespace 设置Nacos命名空间
func WithNacosNamespace(namespace string) ConfigOption {
	return func(o *ConfigOptions) {
		if o.NacosConfig == nil {
			o.NacosConfig = &NacosConfig{}
		}
		o.NacosConfig.NamespaceID = namespace
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
		o.NacosConfig.Timeout = timeout
	}
}
