package nacos

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	config_client "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
)

// Options Nacos配置源选项
type Options struct {
	Endpoints     []string                // Nacos服务器地址列表
	NamespaceID   string                  // 命名空间ID
	Group         string                  // 配置分组
	DataID        string                  // 配置ID
	Username      string                  // 用户名
	Password      string                  // 密码
	TimeoutMs     uint64                  // 超时时间(毫秒)
	LogDir        string                  // 日志目录
	CacheDir      string                  // 缓存目录
	LogLevel      string                  // 日志级别
	Scheme        string                  // 协议(http/https)
	ConfigType    string                  // 配置类型
	OnChange      func(configTypes.Event) // 配置变更回调函数
	RetryTimes    int                     // 重试次数
	RetryInterval time.Duration           // 重试间隔

	// 日志相关配置
	LogConfig struct {
		Dir           string
		Level         string
		MaxAge        int  // 日志保留天数
		MaxSize       int  // 单个日志文件最大尺寸(MB)
		MaxBackups    int  // 最大备份数
		DisableStdout bool // 是否禁用标准输出
	}
}

// NacosSource 实现基于Nacos的配置源
type NacosSource struct {
	Client     config_client.IConfigClient
	V          *viper.Viper
	Opts       *Options
	OnChange   func(configTypes.Event)
	mu         sync.RWMutex
	logAdapter *NacosLogAdapter
}

// NewSourceWithClient 创建一个带有预配置客户端的 NacosSource
func NewSourceWithClient(client config_client.IConfigClient, opts *Options) (*NacosSource, error) {
	if client == nil {
		return nil, errors.NewConfigError("client cannot be nil", nil)
	}
	if opts == nil {
		return nil, errors.NewConfigError("options cannot be nil", nil)
	}
	if opts.DataID == "" {
		return nil, errors.NewConfigError("dataId is required", nil)
	}

	// 设置默认配置类型
	if opts.ConfigType == "" {
		opts.ConfigType = "yaml"
	}

	v := viper.New()
	v.SetConfigType(opts.ConfigType)

	// 创建日志适配器
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}
	logAdapter := NewNacosLogAdapter(logger, opts.LogDir)

	return &NacosSource{
		Client:     client,
		V:          v,
		Opts:       opts,
		OnChange:   opts.OnChange,
		logAdapter: logAdapter,
	}, nil
}

// NewSource 创建新的Nacos配置源
func NewSource(opts *Options) (*NacosSource, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// 创建Nacos客户端配置
	clientConfig := &constant.ClientConfig{
		NamespaceId:         opts.NamespaceID,
		TimeoutMs:           opts.TimeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              opts.LogDir,
		CacheDir:            opts.CacheDir,
		LogLevel:            opts.LogLevel,
		Username:            opts.Username,
		Password:            opts.Password,
		ContextPath:         "/nacos",
	}

	// 创建服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(opts.Endpoints))
	for _, endpoint := range opts.Endpoints {
		host, port := parseEndpoint(endpoint)
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      host,
			Port:        port,
			Scheme:      opts.Scheme,
			ContextPath: "/nacos",
		})
	}

	// 使用 CreateConfigClient
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, errors.NewConfigError("failed to create nacos client", err)
	}

	return &NacosSource{
		Client:   client,
		V:        viper.New(),
		Opts:     opts,
		OnChange: opts.OnChange,
	}, nil
}

// validateOptions 验证配置选项
func validateOptions(opts *Options) error {
	if opts == nil {
		return errors.NewConfigValidateError("options is nil", nil)
	}

	// 验证必需字段
	if len(opts.Endpoints) == 0 {
		return errors.NewConfigValidateError("no endpoints configured", nil)
	}
	if opts.DataID == "" {
		return errors.NewConfigValidateError("dataID is required", nil)
	}

	// 设置默认值
	if opts.NamespaceID == "" {
		opts.NamespaceID = "public" // 使用默认命名空间
	}
	if opts.Group == "" {
		opts.Group = "DEFAULT_GROUP" // 使用默认分组
	}
	if opts.TimeoutMs == 0 {
		opts.TimeoutMs = 5000 // 默认5秒超时
	}
	if opts.ConfigType == "" {
		opts.ConfigType = "yaml" // 默认使用 yaml 格式
	}
	if opts.LogDir == "" {
		opts.LogDir = "logs/nacos" // 默认日志目录
	}
	if opts.CacheDir == "" {
		opts.CacheDir = "cache/nacos" // 默认缓存目录
	}
	if opts.RetryTimes == 0 {
		opts.RetryTimes = 3 // 默认重试3次
	}
	if opts.RetryInterval == 0 {
		opts.RetryInterval = time.Second // 默认重试间隔1秒
	}
	if opts.Scheme == "" {
		opts.Scheme = "http"
	}

	// 验证超时设置
	if opts.TimeoutMs < 1000 {
		return errors.NewConfigValidateError("timeout must be at least 1000ms", nil)
	}

	// 验证重试设置
	if opts.RetryTimes < 0 {
		return errors.NewConfigValidateError("retry times cannot be negative", nil)
	}
	if opts.RetryInterval < time.Millisecond*100 {
		return errors.NewConfigValidateError("retry interval must be at least 100ms", nil)
	}

	return nil
}

// Load 加载配置
func (s *NacosSource) Load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error
	for i := 0; i <= s.Opts.RetryTimes; i++ {
		content, err := s.Client.GetConfig(vo.ConfigParam{
			DataId: s.Opts.DataID,
			Group:  s.Opts.Group,
		})

		if err == nil {
			// 重置 viper 实例
			s.V.SetConfigType(s.Opts.ConfigType)
			if err := s.V.ReadConfig(strings.NewReader(content)); err != nil {
				return errors.NewConfigError("failed to parse config content", err)
			}
			return nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.Opts.RetryInterval):
			continue
		}
	}
	return errors.NewConfigError("failed to load config after retries", lastErr)
}

// Get 获取配置值
func (s *NacosSource) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查是否已加载配置
	if s.V == nil {
		return nil, errors.NewConfigError("configuration not initialized", nil)
	}

	if !s.V.IsSet(key) {
		return nil, errors.NewConfigError("key not found", nil)
	}

	val := s.V.Get(key)
	if val == nil {
		return nil, errors.NewConfigError("value is nil", nil)
	}

	return val, nil
}

// Watch 开始监听配置变更
func (s *NacosSource) Watch(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Client.ListenConfig(vo.ConfigParam{
		DataId: s.Opts.DataID,
		Group:  s.Opts.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if s.OnChange != nil {
				// 保存旧配置
				oldConfig := s.V.AllSettings()

				// 更新配置
				s.V.SetConfigType("yaml")
				if err := s.V.ReadConfig(strings.NewReader(data)); err != nil {
					return
				}

				// 触发变更事件
				s.OnChange(configTypes.Event{
					Key:      dataId,
					Value:    s.V.AllSettings(),
					OldValue: oldConfig,
					Type:     configTypes.EventUpdate,
				})
			}
		},
	})
}

// StopWatch 停止监听配置
func (s *NacosSource) StopWatch(ctx context.Context) error {
	// 避免死锁，使用 tryLock
	if !s.mu.TryLock() {
		return nil // 已经在关闭过程中
	}
	defer s.mu.Unlock()

	s.OnChange = nil
	return nil
}

// Close 关闭配置源
func (s *NacosSource) Close(ctx context.Context) error {
	// 在关闭后执行额外的清理
	defer func() {
		// 强制关闭所有日志文件
		if s.logAdapter != nil {
			s.logAdapter.Close()
			s.logAdapter = nil
		}

		// 在 Windows 上运行垃圾回收可能有助于释放文件句柄
		runtime.GC()
	}()

	// 使用 context 控制超时
	done := make(chan error, 1)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		var errs []error

		// 1. 停止监听
		if err := s.StopWatch(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop watch: %v", err))
		}

		// 2. 关闭日志适配器
		if s.logAdapter != nil {
			if err := s.logAdapter.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close log adapter: %v", err))
			}
		}

		// 3. 关闭客户端
		if s.Client != nil {
			s.Client.CloseClient()
		}

		// 4. 等待资源释放
		time.Sleep(100 * time.Millisecond)

		if len(errs) > 0 {
			done <- fmt.Errorf("close errors: %v", errs)
			return
		}
		done <- nil
	}()

	// 设置超时
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("close timeout")
	}
}

func parseEndpoint(endpoint string) (string, uint64) {
	parts := strings.Split(endpoint, ":")
	if len(parts) == 2 {
		port, _ := strconv.ParseUint(parts[1], 10, 64)
		return parts[0], port
	}
	return endpoint, 8848
}
