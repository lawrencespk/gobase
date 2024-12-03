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
			BufferSize:    32 * 1024,      // 缓冲区大小
			FlushInterval: time.Second,    // 刷新间隔
			MaxOpenFiles:  100,            // 最大打开文件数
			DefaultPath:   "logs/app.log", // 默认路径
		})

		config := logrus.QueueConfig{
			MaxSize:       10000,       // 最大大小
			BatchSize:     100,         // 批处理大小
			FlushInterval: time.Second, // 刷新间隔
			Workers:       2,           // 工作线程数
		}

		options := &logrus.Options{
			Level:       types.InfoLevel, // 日志级别
			Development: false,           // 是否开发模式
		}

		logger, err := logrus.NewLogger(fm, config, options) // 创建日志实例
		if err != nil {
			defaultLogger = &types.BasicLogger{} // 设置默认日志实例
			return
		}
		defaultLogger = logger // 设置默认日志实例
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
		BufferSize:    32 * 1024,      // 缓冲区大小
		FlushInterval: time.Second,    // 刷新间隔
		MaxOpenFiles:  100,            // 最大打开文件数
		DefaultPath:   "logs/app.log", // 默认路径
	})

	config := logrus.QueueConfig{
		MaxSize:       10000,       // 最大大小
		BatchSize:     100,         // 批处理大小
		FlushInterval: time.Second, // 刷新间隔
		Workers:       2,           // 工作线程数
	}

	options := &logrus.Options{
		Level:       types.InfoLevel, // 日志级别
		Development: false,           // 是否开发模式
	}

	// 应用自定义选项
	for _, opt := range opts {
		opt(options) // 应用自定义选项
	}

	return logrus.NewLogger(fm, config, options) // 创建日志实例
}

// Factory 定义了日志工厂接口
type Factory interface {
	Create(opts ...logrus.Option) (types.Logger, error)
}

// factory 实现了日志工厂接口
type factory struct {
	typ string // 日志类型
}

// NewFactory 创建日志工厂实例
func NewFactory(typ string) Factory {
	return &factory{typ: typ} // 创建日志工厂实例
}

// Create 创建日志实例
func (f *factory) Create(opts ...logrus.Option) (types.Logger, error) {
	switch f.typ {
	case "logrus":
		return NewLogger(opts...) // 创建日志实例
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", f.typ) // 不支持的日志类型
	}
}

// InitializeLogger 初始化日志系统
func InitializeLogger() types.Logger {
	logger := GetLogger() // 获取默认日志实例
	SetLogger(logger)     // 设置默认日志实例
	return logger
}
