package logger

import (
	"fmt"
	"sync"
	"time"

	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

var (
	defaultLogger types.Logger
	once          sync.Once
)

// GetLogger 获取默认日志实例
func GetLogger() types.Logger {
	once.Do(func() {
		fm := logrus.NewFileManager(logrus.FileOptions{
			BufferSize:    32 * 1024,
			FlushInterval: time.Second,
			MaxOpenFiles:  100,
			DefaultPath:   "logs/app.log",
		})

		config := logrus.QueueConfig{
			MaxSize:       10000,
			BatchSize:     100,
			FlushInterval: time.Second,
			Workers:       2,
		}

		options := &logrus.Options{
			Level:       types.InfoLevel,
			Development: false,
		}

		logger, err := logrus.NewLogger(fm, config, options)
		if err != nil {
			defaultLogger = &types.BasicLogger{}
			return
		}
		defaultLogger = logger
	})
	return defaultLogger
}

// SetLogger 设置默认日志实例
func SetLogger(logger types.Logger) {
	defaultLogger = logger
}

// NewLogger 创建新的日志实例
func NewLogger(opts ...logrus.Option) (types.Logger, error) {
	fm := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/app.log",
	})

	config := logrus.QueueConfig{
		MaxSize:       10000,
		BatchSize:     100,
		FlushInterval: time.Second,
		Workers:       2,
	}

	options := &logrus.Options{
		Level:       types.InfoLevel,
		Development: false,
	}

	// 应用自定义选项
	for _, opt := range opts {
		opt(options)
	}

	return logrus.NewLogger(fm, config, options)
}

// Factory 定义了日志工厂接口
type Factory interface {
	Create(opts ...logrus.Option) (types.Logger, error)
}

// factory 实现了日志工厂接口
type factory struct {
	typ string
}

// NewFactory 创建日志工厂实例
func NewFactory(typ string) Factory {
	return &factory{typ: typ}
}

// Create 创建日志实例
func (f *factory) Create(opts ...logrus.Option) (types.Logger, error) {
	switch f.typ {
	case "logrus":
		return NewLogger(opts...)
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", f.typ)
	}
}

// InitializeLogger 初始化日志系统
func InitializeLogger() types.Logger {
	logger := GetLogger()
	SetLogger(logger)
	return logger
}
