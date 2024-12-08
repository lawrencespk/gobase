package nacos

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
)

// NacosSource 实现基于Nacos的配置源
type NacosSource struct {
	mu       sync.RWMutex
	client   config_client.IConfigClient
	v        *viper.Viper // 用于解析配置内容
	onChange func(configTypes.Event)

	// Nacos特定配置
	namespace string
	group     string
	dataID    string
}

// Options Nacos配置源选项
type Options struct {
	Endpoints  []string // 服务端点
	Namespace  string   // 命名空间
	Group      string   // 配置分组
	DataID     string   // 配置ID
	Username   string   // 用户名
	Password   string   // 密码
	TimeoutMs  uint64   // 超时时间(毫秒)
	LogDir     string   // 日志目录
	CacheDir   string   // 缓存目录
	LogLevel   string   // 日志级别
	Scheme     string   // 协议(http/https)
	EnableAuth bool     // 是否启用认证
	AccessKey  string   // 访问密钥
	SecretKey  string   // 密钥
	FileType   string   // 配置文件类型(yaml/json等)
}

// New 创建Nacos配置源
func New(opts *Options) (*NacosSource, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// 创建Nacos客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         opts.Namespace,
		TimeoutMs:           opts.TimeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              opts.LogDir,
		CacheDir:            opts.CacheDir,
		LogLevel:            opts.LogLevel,
		Username:            opts.Username,
		Password:            opts.Password,
		AccessKey:           opts.AccessKey,
		SecretKey:           opts.SecretKey,
	}

	// 创建服务端配置
	serverConfigs := make([]constant.ServerConfig, 0, len(opts.Endpoints))
	for _, endpoint := range opts.Endpoints {
		host, port, err := parseEndpoint(endpoint)
		if err != nil {
			return nil, errors.NewConfigValidateError("invalid endpoint", err)
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			Scheme:      opts.Scheme,
			ContextPath: "/nacos",
			IpAddr:      host,
			Port:        uint64(port),
		})
	}

	// 创建配置客户端
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, errors.NewConfigProviderError("failed to create nacos client", err)
	}

	// 创建viper实例用于解析配置
	v := viper.New()
	if opts.FileType != "" {
		v.SetConfigType(opts.FileType)
	}

	return &NacosSource{
		client:    client,
		v:         v,
		namespace: opts.Namespace,
		group:     opts.Group,
		dataID:    opts.DataID,
	}, nil
}

// validateOptions 验证配置选项
func validateOptions(opts *Options) error {
	if opts == nil {
		return errors.NewConfigValidateError("options is nil", nil)
	}
	if len(opts.Endpoints) == 0 {
		return errors.NewConfigValidateError("endpoints is empty", nil)
	}
	if opts.Namespace == "" {
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

// parseEndpoint 解析端点配置
func parseEndpoint(endpoint string) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", 0, err
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

// Load 加载配置
func (s *NacosSource) Load(ctx context.Context) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := s.client.GetConfig(vo.ConfigParam{
		DataId: s.dataID,
		Group:  s.group,
	})
	if err != nil {
		return nil, errors.NewConfigLoadError("failed to get nacos config", err)
	}

	// 解析配置内容
	if err := s.v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, errors.NewConfigParseError("failed to parse config content", err)
	}

	return s.v.AllSettings(), nil
}

// Watch 监听配置变更
func (s *NacosSource) Watch(ctx context.Context, onChange func(configTypes.Event)) error {
	s.mu.Lock()
	s.onChange = onChange
	s.mu.Unlock()

	return s.client.ListenConfig(vo.ConfigParam{
		DataId: s.dataID,
		Group:  s.group,
		OnChange: func(namespace, group, dataId, data string) {
			// 获取旧配置
			oldConfig := s.v.AllSettings()

			// 解析新配置
			if err := s.v.ReadConfig(strings.NewReader(data)); err != nil {
				if s.onChange != nil {
					s.onChange(configTypes.Event{
						Key:   s.dataID,
						Type:  configTypes.EventUpdate,
						Error: err,
					})
				}
				return
			}

			// 触发变更事件
			if s.onChange != nil {
				s.onChange(configTypes.Event{
					Key:      s.dataID,
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

	if s.onChange == nil {
		return nil
	}

	err := s.client.CancelListenConfig(vo.ConfigParam{
		DataId: s.dataID,
		Group:  s.group,
	})
	if err != nil {
		return errors.NewConfigWatchError("failed to cancel config listening", err)
	}

	s.onChange = nil
	return nil
}

// Close 关闭配置源
func (s *NacosSource) Close(ctx context.Context) error {
	return s.StopWatch(ctx)
}
