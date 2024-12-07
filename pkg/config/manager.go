package config

import (
	"sync"
	"time"

	"gobase/pkg/config/nacos"
	"gobase/pkg/config/types"
	"gobase/pkg/config/viper"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// Manager 配置管理器
type Manager struct {
	loader      types.Loader
	parser      types.Parser
	nacosClient *nacos.Client
	mutex       sync.RWMutex
	watchers    map[string][]func(string, interface{})
}

// NewManager 创建配置管理器（使用Viper作为默认实现）
func NewManager(opts ...types.ConfigOption) (*Manager, error) {
	// 默认配置选项
	options := &types.ConfigOptions{
		ConfigFile:  "config/config.yaml",
		ConfigType:  "yaml",
		Environment: "development",
		EnableEnv:   true,
		EnvPrefix:   "GOBASE",
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(options)
	}

	// 创建配置管理器
	m := &Manager{
		watchers: make(map[string][]func(string, interface{})),
	}

	// 初始化loader和parser
	if err := m.initLoaderAndParser(options); err != nil {
		return nil, err
	}

	// 如果配置了Nacos，则初始化Nacos客户端
	if m.loader.IsSet("nacos") {
		// 从配置文件读取Nacos配置
		nacosConfig := &nacos.Config{
			Endpoint:    m.loader.GetString("nacos.endpoint"),
			NamespaceID: m.loader.GetString("nacos.namespace_id"),
			Group:       m.loader.GetString("nacos.group"),
			DataID:      m.loader.GetString("nacos.data_id"),
			Username:    m.loader.GetString("nacos.username"),
			Password:    m.loader.GetString("nacos.password"),
			Timeout:     m.loader.GetDuration("nacos.timeout"),
			LogDir:      m.loader.GetString("nacos.log_dir"),
			CacheDir:    m.loader.GetString("nacos.cache_dir"),
			LogLevel:    m.loader.GetString("nacos.log_level"),
			EnableAuth:  m.loader.GetBool("nacos.enable_auth"),
			AccessKey:   m.loader.GetString("nacos.access_key"),
			SecretKey:   m.loader.GetString("nacos.secret_key"),
			AuthToken:   m.loader.GetString("nacos.auth_token"),
			IdentityKey: m.loader.GetString("nacos.identity_key"),
			IdentityVal: m.loader.GetString("nacos.identity_val"),
		}

		// 如果通过选项函数提供了Nacos配置，则覆盖配置文件中的值
		if options.NacosConfig != nil {
			if options.NacosConfig.Endpoint != "" {
				nacosConfig.Endpoint = options.NacosConfig.Endpoint
			}
			if options.NacosConfig.NamespaceID != "" {
				nacosConfig.NamespaceID = options.NacosConfig.NamespaceID
			}
			if options.NacosConfig.Group != "" {
				nacosConfig.Group = options.NacosConfig.Group
			}
			if options.NacosConfig.DataID != "" {
				nacosConfig.DataID = options.NacosConfig.DataID
			}
			if options.NacosConfig.Username != "" {
				nacosConfig.Username = options.NacosConfig.Username
			}
			if options.NacosConfig.Password != "" {
				nacosConfig.Password = options.NacosConfig.Password
			}
			if options.NacosConfig.Timeout != 0 {
				nacosConfig.Timeout = options.NacosConfig.Timeout
			}
			if options.NacosConfig.LogDir != "" {
				nacosConfig.LogDir = options.NacosConfig.LogDir
			}
			if options.NacosConfig.CacheDir != "" {
				nacosConfig.CacheDir = options.NacosConfig.CacheDir
			}
			if options.NacosConfig.LogLevel != "" {
				nacosConfig.LogLevel = options.NacosConfig.LogLevel
			}
			nacosConfig.EnableAuth = options.NacosConfig.EnableAuth
			if options.NacosConfig.AccessKey != "" {
				nacosConfig.AccessKey = options.NacosConfig.AccessKey
			}
			if options.NacosConfig.SecretKey != "" {
				nacosConfig.SecretKey = options.NacosConfig.SecretKey
			}
			if options.NacosConfig.AuthToken != "" {
				nacosConfig.AuthToken = options.NacosConfig.AuthToken
			}
			if options.NacosConfig.IdentityKey != "" {
				nacosConfig.IdentityKey = options.NacosConfig.IdentityKey
			}
			if options.NacosConfig.IdentityVal != "" {
				nacosConfig.IdentityVal = options.NacosConfig.IdentityVal
			}
		}

		// 初始化Nacos客户端
		nacosClient, err := nacos.NewClient(nacosConfig)
		if err != nil {
			return nil, err
		}
		m.nacosClient = nacosClient
	}

	return m, nil
}

// initLoaderAndParser 初始化加载器和解析器
func (m *Manager) initLoaderAndParser(options *types.ConfigOptions) error {
	if options == nil {
		return errors.NewInvalidParamsError("config options is nil", nil)
	}

	// 创建Viper加载器
	loader := viper.NewLoader(options.ConfigFile, options.EnableEnv, options.EnvPrefix)
	if err := loader.Load(); err != nil {
		return errors.NewError(codes.ConfigLoadError, "failed to load config", err)
	}

	// 创建Viper解析器
	parser := viper.NewParser(loader)

	m.loader = loader
	m.parser = parser
	return nil
}

// NewManagerWithLoader 使用自定义加载器和解析器创建配置管理器
func NewManagerWithLoader(loader types.Loader, parser types.Parser) *Manager {
	return &Manager{
		loader:   loader,
		parser:   parser,
		watchers: make(map[string][]func(string, interface{})),
	}
}

// Get 获取配置值
func (m *Manager) Get(key string) interface{} {
	return m.loader.Get(key)
}

// GetString 获取字符串配置
func (m *Manager) GetString(key string) string {
	return m.loader.GetString(key)
}

// GetInt 获取整数配置
func (m *Manager) GetInt(key string) int {
	return m.loader.GetInt(key)
}

// GetBool 获取布尔配置
func (m *Manager) GetBool(key string) bool {
	return m.loader.GetBool(key)
}

// GetFloat64 获取浮点数配置
func (m *Manager) GetFloat64(key string) float64 {
	return m.loader.GetFloat64(key)
}

// GetDuration 获取时间间隔配置
func (m *Manager) GetDuration(key string) time.Duration {
	return m.loader.GetDuration(key)
}

// GetStringSlice 获取字符串切片配置
func (m *Manager) GetStringSlice(key string) []string {
	return m.loader.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func (m *Manager) GetStringMap(key string) map[string]interface{} {
	return m.loader.GetStringMap(key)
}

// GetStringMapString 获取字符串-字符串映射配置
func (m *Manager) GetStringMapString(key string) map[string]string {
	return m.loader.GetStringMapString(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.loader.Set(key, value)

	// 手动触发 watchers
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if callbacks, ok := m.watchers[key]; ok {
		for _, cb := range callbacks {
			cb(key, value)
		}
	}
}

// IsSet 检查配置是否存在
func (m *Manager) IsSet(key string) bool {
	return m.loader.IsSet(key)
}

// Watch 监听配置变化
func (m *Manager) Watch(key string, callback func(key string, value interface{})) error {
	m.mutex.Lock()
	if _, exists := m.watchers[key]; !exists {
		m.watchers[key] = make([]func(string, interface{}), 0)
	}
	m.watchers[key] = append(m.watchers[key], callback)
	m.mutex.Unlock()

	// 如果使用Nacos，则设置Nacos配置监听
	if m.nacosClient != nil {
		// 从key中提取group和dataId
		group := m.GetString("nacos.group") // 默认从配置中获取group
		if group == "" {
			group = "DEFAULT_GROUP" // 如果未配置，使用默认值
		}

		return m.nacosClient.WatchConfig(key, group, func(dataId, data string) {
			m.mutex.RLock()
			defer m.mutex.RUnlock()
			if callbacks, ok := m.watchers[dataId]; ok {
				for _, cb := range callbacks {
					cb(dataId, data)
				}
			}
		})
	}

	// 否则使用本地文件监听，但使用自己的回调处理
	return m.loader.Watch(key, func(k string, v interface{}) {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		if callbacks, ok := m.watchers[k]; ok {
			for _, cb := range callbacks {
				cb(k, v)
			}
		}
	})
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 清理 watchers
	m.watchers = make(map[string][]func(string, interface{}))

	// 关闭 Nacos 客户端
	if m.nacosClient != nil {
		if err := m.nacosClient.Close(); err != nil {
			return errors.NewError(
				codes.ConfigUpdateError,
				"failed to close nacos client",
				err,
			)
		}
	}

	return nil
}

// Parse 解析配置到结构体
func (m *Manager) Parse(key string, out interface{}) error {
	return m.parser.Parse(key, out)
}
