package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fmt"
	"gobase/pkg/errors"
	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

var (
	defaultLogger types.Logger // 默认日志实例
	once          sync.Once    // 确保单例模式
)

// getLogPath 获取日志路径
func getLogPath() string {
	if path := os.Getenv("LOG_PATH"); path != "" { // 从环境变量获取日志路径
		return path
	}
	return "logs/app.log" // 默认日志路径
}

// GetLogger 获取默认日志实例
func GetLogger() types.Logger {
	once.Do(func() {
		// 从环境变量获取日志路径
		logPath := getLogPath() // 获取日志路径
		ensureLogDir(logPath)   // 确保日志目录存在

		fm := logrus.NewFileManager(logrus.FileOptions{
			BufferSize:    32 * 1024,   // 缓冲区大小
			FlushInterval: time.Second, // 刷新间隔
			MaxOpenFiles:  100,         // 最大打开文件数
			DefaultPath:   logPath,     // 默认路径
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

// ensureLogDir 确保日志目录存在
func ensureLogDir(logPath string) {
	dir := filepath.Dir(logPath) // 获取日志目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		// 如果创建目录失败，可以打印错误或使用其他方式处理
		panic(err)
	}
}

// SetLogger 设置默认日志实例
func SetLogger(logger types.Logger) {
	defaultLogger = logger // 设置默认日志实例
}

// Option 定义配置选项函数类型
type Option func(*Options)

// Options 日志配置选项
type Options struct {
	Level          types.Level
	OutputPaths    []string
	FileConfig     *logrus.FileOptions
	EnableELK      bool
	ELKConfig      *elk.ElkConfig
	RecoveryConfig logrus.RecoveryConfig
	AsyncConfig    logrus.AsyncConfig
	CompressConfig logrus.CompressConfig
	CleanupConfig  logrus.CleanupConfig
	writers        []io.Writer
}

// DefaultOptions 返回默认配置
func DefaultOptions() *Options {
	return &Options{
		Level:       types.InfoLevel,
		OutputPaths: []string{"stdout"},
		EnableELK:   false,
		ELKConfig:   nil,
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(opts ...Option) (types.Logger, error) {
	// 创建并应用选项
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 创建默认的队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:       10000,       // 默认队列大小
		BatchSize:     100,         // 默认批处理大小
		FlushInterval: time.Second, // 默认刷新间隔
		Workers:       2,           // 默认工作线程数
		RetryCount:    3,           // 默认重试次数
		RetryInterval: time.Second, // 默认重试间隔
	}

	// 创建 logrus 选项对象
	logrusOptions := &logrus.Options{
		Level:          options.Level,
		OutputPaths:    options.OutputPaths,
		CompressConfig: options.CompressConfig, // 添加压缩配置
		AsyncConfig:    options.AsyncConfig,    // 添加异步配置
	}

	// 添加自定义 writers
	for _, w := range options.writers {
		if w != nil {
			logrusOptions.AddWriter(w) // 使用公开的方法添加 writer
		}
	}

	// 创建 logrus logger
	logrusLogger, err := logrus.NewLogger(nil, queueConfig, logrusOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	return logrusLogger, nil
}

// WithLevel 设置日志级别
func WithLevel(level types.Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

// WithOutputPaths 设置输出路径
func WithOutputPaths(paths []string) Option {
	return func(o *Options) {
		o.OutputPaths = paths
	}
}

// WithFileConfig 设置文件配置
func WithFileConfig(config *logrus.FileOptions) Option {
	return func(o *Options) {
		o.FileConfig = config
	}
}

// WithELK 启用 ELK 并设置配置
func WithELK(config *elk.ElkConfig) Option {
	return func(o *Options) {
		o.EnableELK = true
		o.ELKConfig = config
	}
}

// WithRecoveryConfig 设置恢复配置
func WithRecoveryConfig(config *logrus.RecoveryConfig) Option {
	return func(o *Options) {
		if config != nil {
			o.RecoveryConfig = *config
		}
	}
}

// WithAsyncConfig 设置异步配置
func WithAsyncConfig(config logrus.AsyncConfig) Option {
	return func(o *Options) {
		o.AsyncConfig = config
	}
}

// WithCompressConfig 设置压缩配置
func WithCompressConfig(config logrus.CompressConfig) Option {
	return func(o *Options) {
		o.CompressConfig = config
	}
}

// WithCleanupConfig 设置清理配置
func WithCleanupConfig(config logrus.CleanupConfig) Option {
	return func(o *Options) {
		o.CleanupConfig = config
	}
}

// WithOutput 添加一个输出 writer
func WithOutput(w io.Writer) Option {
	return func(o *Options) {
		if o.writers == nil {
			o.writers = make([]io.Writer, 0)
		}
		o.writers = append(o.writers, w)
	}
}

// InitializeLogger 初始化日志系统
func InitializeLogger() types.Logger {
	logger := GetLogger() // 获取默认日志实例
	SetLogger(logger)     // 设置默认日志实例
	return logger
}

// LoggerWithHooks 定义带有 hooks 的 logger 接口
type LoggerWithHooks interface {
	types.Logger
	WaitForHooks() // 等待所有 hooks 处理完成
}

type loggerWithHooks struct {
	types.Logger
	wg sync.WaitGroup
}

func (l *loggerWithHooks) WaitForHooks() {
	l.wg.Wait()
}

// Factory 定义日志工厂接口
type Factory interface {
	Create(opts ...Option) (types.Logger, error)
	CreateWithHooks(opts ...Option) (LoggerWithHooks, error)
}

type loggerFactory struct {
	loggerType string
}

// NewFactory 创建新的日志工厂
func NewFactory(loggerType string) Factory {
	return &loggerFactory{
		loggerType: loggerType,
	}
}

// Create 创建新的日志实例
func (f *loggerFactory) Create(opts ...Option) (types.Logger, error) {
	switch f.loggerType {
	case "logrus":
		// 创建并应用选项
		options := DefaultOptions()
		for _, opt := range opts {
			opt(options)
		}

		// 创建 FileManager
		var fm *logrus.FileManager
		if options.FileConfig != nil {
			fm = logrus.NewFileManager(*options.FileConfig)
		} else if len(options.OutputPaths) > 0 {
			// 如果没有明确的 FileConfig 但有输出路径，创建默认的 FileManager
			fm = logrus.NewFileManager(logrus.FileOptions{
				BufferSize:    32 * 1024,   // 默认缓冲区大小
				FlushInterval: time.Second, // 默认刷新间隔
				MaxOpenFiles:  100,         // 默认最大打开文件数
			})
		}

		// 创建 logrus 选项对象
		logrusOptions := &logrus.Options{
			Level:          options.Level,
			OutputPaths:    options.OutputPaths,
			RecoveryConfig: options.RecoveryConfig,
			CompressConfig: options.CompressConfig,
			CleanupConfig:  options.CleanupConfig,
			AsyncConfig:    options.AsyncConfig,
		}

		// 添加自定义 writers
		if len(options.writers) > 0 {
			for _, w := range options.writers {
				// 确保每个 writer 都被正确添加
				if w != nil {
					logrusOptions.AddWriter(w)
				}
			}
			// 强制启用异步写入以确保所有 writers 都能正确处理
			if !logrusOptions.AsyncConfig.Enable {
				logrusOptions.AsyncConfig.Enable = true
				logrusOptions.AsyncConfig.FlushOnExit = true
			}
		}

		// 创建默认的队列配置
		queueConfig := logrus.QueueConfig{
			MaxSize:       10000,       // 默认队列大小
			BatchSize:     100,         // 默认批处理大小
			FlushInterval: time.Second, // 默认刷新间隔
			Workers:       2,           // 默认工作线程数
			RetryCount:    3,           // 默认重试次数
			RetryInterval: time.Second, // 默认重试间隔
		}

		// 使用 FileManager 创建 logger
		return logrus.NewLogger(fm, queueConfig, logrusOptions)
	default:
		return nil, errors.NewLogConfigError("unsupported logger type", fmt.Errorf("type: %s", f.loggerType))
	}
}

// CreateWithHooks 创建带有 hooks 的 logger
func (f *loggerFactory) CreateWithHooks(opts ...Option) (LoggerWithHooks, error) {
	switch f.loggerType {
	case "logrus":
		baseLogger, err := f.Create(opts...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create base logger")
		}

		return &loggerWithHooks{
			Logger: baseLogger,
		}, nil
	default:
		return nil, errors.NewLogConfigError("unsupported logger type", fmt.Errorf("type: %s", f.loggerType))
	}
}
