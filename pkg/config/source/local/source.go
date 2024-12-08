package local

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
)

// LocalSource 实现基于本地文件的配置源
type LocalSource struct {
	mu       sync.RWMutex
	v        *viper.Viper
	filePath string
	onChange func(configTypes.Event)
}

// Options 本地配置源选项
type Options struct {
	FilePath string                 // 配置文件路径
	FileType string                 // 配置文件类型(yaml/json等)
	Defaults map[string]interface{} // 默认值
}

// New 创建本地配置源
func New(opts *Options) (*LocalSource, error) {
	if opts == nil || opts.FilePath == "" {
		return nil, errors.NewConfigValidateError("file path is required", nil)
	}

	// 创建viper实例
	v := viper.New()
	v.SetConfigFile(opts.FilePath)
	if opts.FileType != "" {
		v.SetConfigType(opts.FileType)
	}

	// 设置默认值
	if opts.Defaults != nil {
		for k, val := range opts.Defaults {
			v.SetDefault(k, val)
		}
	}

	return &LocalSource{
		v:        v,
		filePath: opts.FilePath,
	}, nil
}

// Load 加载配置
func (s *LocalSource) Load(ctx context.Context) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.v.ReadInConfig(); err != nil {
		return nil, errors.NewConfigLoadError("failed to read config file", err)
	}

	return s.v.AllSettings(), nil
}

// Watch 监听配置变更
func (s *LocalSource) Watch(ctx context.Context, onChange func(configTypes.Event)) error {
	s.mu.Lock()
	s.onChange = onChange
	s.mu.Unlock()

	// 监听文件变更
	s.v.OnConfigChange(func(in fsnotify.Event) {
		// 获取旧配置
		oldConfig := s.v.AllSettings()

		// 重新加载配置
		if err := s.v.ReadInConfig(); err != nil {
			if s.onChange != nil {
				s.onChange(configTypes.Event{
					Key:   s.filePath,
					Type:  configTypes.EventUpdate,
					Error: err,
				})
			}
			return
		}

		// 触发变更事件
		if s.onChange != nil {
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
