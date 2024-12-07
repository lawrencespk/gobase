package config

import (
	"fmt"
	"sync"
	"time"

	"gobase/pkg/config/nacos"
	"gobase/pkg/config/types"
	"gobase/pkg/config/viper"
	"gobase/pkg/errors"
)

// Manager 配置管理器
type Manager struct {
	loader      *viper.Loader
	parser      *viper.Parser
	nacosClient *nacos.Client
	mutex       sync.RWMutex
	watchers    map[string][]func(string, interface{})
}

// NewManager 创建配置管理器
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

	// 创建Viper加载器
	loader := viper.NewLoader(options.ConfigFile, options.EnableEnv, options.EnvPrefix)
	if err := loader.Load(); err != nil {
		return nil, err
	}

	// 创建配置解析器
	parser := viper.NewParser(loader)

	// 创建配置管理器
	m := &Manager{
		loader:   loader,
		parser:   parser,
		watchers: make(map[string][]func(string, interface{})),
	}

	// 如果配置了Nacos，则初始化Nacos客户端
	if loader.IsSet("nacos") {
		// 从配置文件读取Nacos配置
		nacosConfig := &nacos.Config{
			Endpoint:    loader.GetString("nacos.endpoint"),
			NamespaceID: loader.GetString("nacos.namespace_id"),
			Group:       loader.GetString("nacos.group"),
			DataID:      loader.GetString("nacos.data_id"),
			Username:    loader.GetString("nacos.username"),
			Password:    loader.GetString("nacos.password"),
			Timeout:     loader.GetDuration("nacos.timeout"),
			LogDir:      loader.GetString("nacos.log_dir"),
			CacheDir:    loader.GetString("nacos.cache_dir"),
			LogLevel:    loader.GetString("nacos.log_level"),
			EnableAuth:  loader.GetBool("nacos.enable_auth"),
			AccessKey:   loader.GetString("nacos.access_key"),
			SecretKey:   loader.GetString("nacos.secret_key"),
			AuthToken:   loader.GetString("nacos.auth_token"),
			IdentityKey: loader.GetString("nacos.identity_key"),
			IdentityVal: loader.GetString("nacos.identity_val"),
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

	// 否则使用本地文件监听
	return m.loader.Watch(key, callback)
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	if m.nacosClient != nil {
		if err := m.nacosClient.Close(); err != nil {
			return errors.NewConfigUpdateError(
				fmt.Sprintf("failed to close nacos client: %v", err),
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
