package local

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
)

// Options 本地配置源选项
type Options struct {
	Path     string                  // 配置文件路径
	OnChange func(configTypes.Event) // 配置变更回调函数
}

// LocalSource 实现基于本地文件的配置源
type LocalSource struct {
	mu       sync.RWMutex
	v        *viper.Viper
	filePath string
	onChange func(configTypes.Event)
}

// NewSource 创建新的本地配置源
func NewSource(opts *Options) (*LocalSource, error) {
	if opts == nil || opts.Path == "" {
		return nil, errors.NewConfigError("invalid options", nil)
	}

	s := &LocalSource{
		v:        viper.New(),
		filePath: opts.Path,
		onChange: opts.OnChange,
	}

	s.v.SetConfigFile(opts.Path)
	return s, nil
}

// Load 加载配置
func (s *LocalSource) Load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.v.ReadInConfig(); err != nil {
		return errors.NewConfigError("failed to read config", err)
	}
	return nil
}

// Get 获取配置值
func (s *LocalSource) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.v.IsSet(key) {
		return nil, errors.NewConfigError("key not found", nil)
	}
	return s.v.Get(key), nil
}

// Watch 开始监听配置变更
func (s *LocalSource) Watch(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.v.OnConfigChange(func(in fsnotify.Event) {
		if s.onChange != nil {
			// 保存旧配置
			oldConfig := s.v.AllSettings()

			// 重新加载配置
			if err := s.v.ReadInConfig(); err != nil {
				return
			}

			// 触发变更事件
			s.onChange(configTypes.Event{
				Key:      s.filePath,
				Value:    s.v.AllSettings(),
				OldValue: oldConfig,
				Type:     configTypes.EventUpdate,
			})
		}
	})

	s.v.WatchConfig()
	return nil
}

// StopWatch 停止监听配置
func (s *LocalSource) StopWatch(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onChange = nil
	return nil
}

// Close 关闭配置源
func (s *LocalSource) Close(ctx context.Context) error {
	return s.StopWatch(ctx)
}
