package provider

import (
	"context"
	"sync"
	"time"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
	loggerTypes "gobase/pkg/logger/types"

	"github.com/mitchellh/mapstructure"
)

// BaseProvider 提供基础实现
type BaseProvider struct {
	mu       sync.RWMutex
	loaded   bool
	watching bool
	logger   loggerTypes.Logger
	opts     *configTypes.Options
	cache    sync.Map
	onChange func(configTypes.Event)

	// 抽象数据源接口
	source ConfigSource
}

// ConfigSource 定义配置源接口
type ConfigSource interface {
	// 加载配置内容
	Load(ctx context.Context) (map[string]interface{}, error)
	// 开始监听配置
	Watch(ctx context.Context, onChange func(configTypes.Event)) error
	// 停止监听配置
	StopWatch(ctx context.Context) error
	// 关闭并清理资源
	Close(ctx context.Context) error
}

// New 创建基础提供者
func New(source ConfigSource, opts *configTypes.Options) *BaseProvider {
	if opts == nil {
		opts = &configTypes.Options{
			Name:     "base",
			AutoLoad: true,
		}
	}

	return &BaseProvider{
		source: source,
		opts:   opts,
		logger: opts.Logger,
	}
}

// Load 加载配置
func (p *BaseProvider) Load(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 从配置源加载
	data, err := p.source.Load(ctx)
	if err != nil {
		if p.logger != nil {
			p.logger.Error(ctx, "Failed to load config", loggerTypes.Error(err))
		}
		return errors.NewConfigLoadError("failed to load config", err)
	}

	// 更新缓存
	for k, v := range data {
		p.cache.Store(k, v)
	}

	p.loaded = true
	return nil
}

// Get 获取配置值
func (p *BaseProvider) Get(key string) (interface{}, error) {
	if !p.IsLoaded() {
		return nil, errors.NewConfigNotLoadedError("config not loaded", nil)
	}

	if val, ok := p.cache.Load(key); ok {
		return val, nil
	}
	return nil, errors.NewConfigKeyNotFoundError("key not found", nil)
}

// GetString 获取字符串配置
func (p *BaseProvider) GetString(key string) (string, error) {
	val, err := p.Get(key)
	if err != nil {
		return "", err
	}

	str, ok := val.(string)
	if !ok {
		return "", errors.NewConfigTypeError("value is not string", nil)
	}
	return str, nil
}

// GetInt 获取整数配置
func (p *BaseProvider) GetInt(key string) (int, error) {
	val, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, errors.NewConfigTypeError("value cannot be converted to int", nil)
	}
}

// GetInt64 获取64位整数配置
func (p *BaseProvider) GetInt64(key string) (int64, error) {
	val, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, errors.NewConfigTypeError("value cannot be converted to int64", nil)
	}
}

// GetFloat64 获取浮点数配置
func (p *BaseProvider) GetFloat64(key string) (float64, error) {
	val, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, errors.NewConfigTypeError("value cannot be converted to float64", nil)
	}
}

// GetBool 获取布尔值配置
func (p *BaseProvider) GetBool(key string) (bool, error) {
	val, err := p.Get(key)
	if err != nil {
		return false, err
	}

	b, ok := val.(bool)
	if !ok {
		return false, errors.NewConfigTypeError("value is not bool", nil)
	}
	return b, nil
}

// GetTime 获取时间配置
func (p *BaseProvider) GetTime(key string) (time.Time, error) {
	val, err := p.GetString(key)
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, errors.NewConfigTypeError("value cannot be parsed as time", err)
	}
	return t, nil
}

// GetDuration 获取时间间隔配置
func (p *BaseProvider) GetDuration(key string) (time.Duration, error) {
	val, err := p.GetString(key)
	if err != nil {
		return 0, err
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, errors.NewConfigTypeError("value cannot be parsed as duration", err)
	}
	return d, nil
}

// GetStringSlice 获取字符串切片配置
func (p *BaseProvider) GetStringSlice(key string) ([]string, error) {
	val, err := p.Get(key)
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case []string:
		return v, nil
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, errors.NewConfigTypeError("slice item is not string", nil)
			}
			strs[i] = str
		}
		return strs, nil
	default:
		return nil, errors.NewConfigTypeError("value is not string slice", nil)
	}
}

// GetStringMap 获取字符串映射配置
func (p *BaseProvider) GetStringMap(key string) (map[string]interface{}, error) {
	val, err := p.Get(key)
	if err != nil {
		return nil, err
	}

	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, errors.NewConfigTypeError("value is not string map", nil)
	}
	return m, nil
}

// GetStringMapString 获取字符串-字符串映射配置
func (p *BaseProvider) GetStringMapString(key string) (map[string]string, error) {
	val, err := p.GetStringMap(key)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for k, v := range val {
		str, ok := v.(string)
		if !ok {
			return nil, errors.NewConfigTypeError("map value is not string", nil)
		}
		result[k] = str
	}
	return result, nil
}

// Unmarshal 将配置解析到结构体
func (p *BaseProvider) Unmarshal(val interface{}) error {
	if !p.IsLoaded() {
		return errors.NewConfigNotLoadedError("config not loaded", nil)
	}

	config := make(map[string]interface{})
	p.cache.Range(func(key, value interface{}) bool {
		config[key.(string)] = value
		return true
	})

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           val,
		WeaklyTypedInput: true,
		TagName:          "yaml",
	})
	if err != nil {
		return errors.NewConfigParseError("failed to create decoder", err)
	}

	if err := decoder.Decode(config); err != nil {
		return errors.NewConfigParseError("failed to decode config", err)
	}

	return nil
}

// UnmarshalKey 将指定键的配置解析到结构体
func (p *BaseProvider) UnmarshalKey(key string, val interface{}) error {
	value, err := p.Get(key)
	if err != nil {
		return err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           val,
		WeaklyTypedInput: true,
		TagName:          "yaml",
	})
	if err != nil {
		return errors.NewConfigParseError("failed to create decoder", err)
	}

	if err := decoder.Decode(value); err != nil {
		return errors.NewConfigParseError("failed to decode config", err)
	}

	return nil
}

// WatchConfig 监听配置变更
func (p *BaseProvider) WatchConfig(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.watching {
		return nil
	}

	if err := p.source.Watch(ctx, p.handleConfigChange); err != nil {
		if p.logger != nil {
			p.logger.Error(ctx, "Failed to watch config", loggerTypes.Error(err))
		}
		return errors.NewConfigWatchError("failed to watch config", err)
	}

	p.watching = true
	return nil
}

// StopWatchConfig 停止监听配置
func (p *BaseProvider) StopWatchConfig(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.watching {
		return nil
	}

	if err := p.source.StopWatch(ctx); err != nil {
		if p.logger != nil {
			p.logger.Error(ctx, "Failed to stop watching config", loggerTypes.Error(err))
		}
		return errors.NewConfigWatchError("failed to stop watching config", err)
	}

	p.watching = false
	return nil
}

// Close 关闭提供者
func (p *BaseProvider) Close(ctx context.Context) error {
	if err := p.StopWatchConfig(ctx); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.source.Close(ctx); err != nil {
		if p.logger != nil {
			p.logger.Error(ctx, "Failed to close config source", loggerTypes.Error(err))
		}
		return errors.NewConfigCloseError("failed to close config source", err)
	}

	p.loaded = false
	p.cache = sync.Map{}

	return nil
}

// SetLogger 设置日志记录器
func (p *BaseProvider) SetLogger(logger loggerTypes.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger = logger
}

// OnConfigChange 注册配置变更回调函数
func (p *BaseProvider) OnConfigChange(fn func(configTypes.Event)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onChange = fn
}

// IsLoaded 检查配置是否已加载
func (p *BaseProvider) IsLoaded() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loaded
}

// handleConfigChange 处理配置变更
func (p *BaseProvider) handleConfigChange(event configTypes.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.onChange != nil {
		// 添加时间戳
		event.Timestamp = time.Now()
		// 添加提供者信息
		event.Provider = p.opts.Name
		p.onChange(event)
	}
}
