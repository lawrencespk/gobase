package nacos

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
)

// Options Nacos配置源选项
type Options struct {
	Endpoint      string                  // Nacos服务器地址
	NamespaceID   string                  // 命名空间ID
	Group         string                  // 配置分组
	DataID        string                  // 配置ID
	Username      string                  // 用户名
	Password      string                  // 密码
	OnChange      func(configTypes.Event) // 配置变更回调函数
	RetryTimes    int                     // 重试次数
	RetryInterval time.Duration           // 重试间隔
}

// NacosSource 实现基于Nacos的配置源
type NacosSource struct {
	mu       sync.RWMutex
	client   config_client.IConfigClient
	v        *viper.Viper
	opts     *Options
	onChange func(configTypes.Event)
}

// NewSourceWithClient 使用指定的客户端创建配置源
func NewSourceWithClient(opts *Options, client config_client.IConfigClient) (*NacosSource, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return &NacosSource{
		client:   client,
		v:        viper.New(),
		opts:     opts,
		onChange: opts.OnChange,
	}, nil
}

// NewSource 创建新的Nacos配置源
func NewSource(opts *Options) (*NacosSource, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// 创建Nacos客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         opts.NamespaceID,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		Username:            opts.Username,
		Password:            opts.Password,
	}

	// 创建服务器配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      opts.Endpoint,
			ContextPath: "/nacos",
			Port:        8848,
		},
	}

	// 创建Nacos配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, errors.NewConfigError("failed to create nacos client", err)
	}

	return &NacosSource{
		client:   client,
		v:        viper.New(),
		opts:     opts,
		onChange: opts.OnChange,
	}, nil
}

// validateOptions 验证配置选项
func validateOptions(opts *Options) error {
	if opts == nil {
		return errors.NewConfigValidateError("options is nil", nil)
	}
	if opts.Endpoint == "" {
		return errors.NewConfigValidateError("endpoint is empty", nil)
	}
	if opts.NamespaceID == "" {
		return errors.NewConfigValidateError("namespace is empty", nil)
	}
	if opts.Group == "" {
		return errors.NewConfigValidateError("group is empty", nil)
	}
	if opts.DataID == "" {
		return errors.NewConfigValidateError("dataID is empty", nil)
	}
	return nil
}

// Load 加载配置
func (s *NacosSource) Load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := s.client.GetConfig(vo.ConfigParam{
		DataId: s.opts.DataID,
		Group:  s.opts.Group,
	})
	if err != nil {
		return errors.NewConfigError("failed to get config from nacos", err)
	}

	s.v.SetConfigType("yaml")
	if err := s.v.ReadConfig(strings.NewReader(content)); err != nil {
		return errors.NewConfigError("failed to parse config", err)
	}

	return nil
}

// Get 获取配置值
func (s *NacosSource) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.v.IsSet(key) {
		return nil, errors.NewConfigError("key not found", nil)
	}
	return s.v.Get(key), nil
}

// Watch 开始监听配置变更
func (s *NacosSource) Watch(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.ListenConfig(vo.ConfigParam{
		DataId: s.opts.DataID,
		Group:  s.opts.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if s.onChange != nil {
				// 保存旧配置
				oldConfig := s.v.AllSettings()

				// 更新配置
				s.v.SetConfigType("yaml")
				if err := s.v.ReadConfig(strings.NewReader(data)); err != nil {
					return
				}

				// 触发变更事件
				s.onChange(configTypes.Event{
					Key:      dataId,
					Value:    s.v.AllSettings(),
					OldValue: oldConfig,
					Type:     configTypes.EventUpdate,
				})
			}
		},
	})
}

// StopWatch 停止监听配置
func (s *NacosSource) StopWatch(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.CancelListenConfig(vo.ConfigParam{
		DataId: s.opts.DataID,
		Group:  s.opts.Group,
	})
}

// Close 关闭配置源
func (s *NacosSource) Close(ctx context.Context) error {
	return s.StopWatch(ctx)
}
