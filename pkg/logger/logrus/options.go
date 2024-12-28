package logrus

import (
	"compress/gzip"
	"gobase/pkg/config"
	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/types"
	"io"
	"time"
)

// DefaultOptions 返回默认选项
func DefaultOptions() *Options {
	conf := config.GetConfig()
	opts := &Options{
		Development: false,           // 默认不是开发模式
		Level:       types.InfoLevel, // 默认日志级别为 Info

		ReportCaller:     true,               // 默认报告调用者信息
		TimeFormat:       time.RFC3339,       // 默认时间格式为 RFC3339
		MaxAge:           7 * 24 * time.Hour, // 默认日志保留时间为 7 天
		RotationTime:     24 * time.Hour,     // 默认日志轮转时间为 24 小时
		MaxSize:          100,                // 默认单个日志文件最大大小为 100MB
		Compress:         true,               // 默认压缩旧日志
		OutputPaths:      []string{"stdout"}, // 默认输出路径为 stdout
		ErrorOutputPaths: []string{"stderr"}, // 默认错误输出路径为 stderr
		CompressConfig: CompressConfig{
			Enable:       true,                 // 默认启用压缩
			Algorithm:    "gzip",               // 默认压缩算法为 gzip
			Level:        gzip.BestCompression, // 默认压缩级别为 BestCompression
			DeleteSource: true,                 // 默认压缩后删除源文件
			Interval:     1 * time.Hour,        // 默认压缩检查间隔为 1 小时
		},
		CleanupConfig: CleanupConfig{
			Enable:     true,
			MaxBackups: 3,              // 保留的旧日志文件个数
			MaxAge:     7,              // 日志文件的最大保留天数
			Interval:   24 * time.Hour, // 清理检查间隔
		},
		AsyncConfig: AsyncConfig{
			Enable:        true,        // 默认启用异步写入
			BufferSize:    8192,        // 默认缓冲区大小为 8192
			FlushInterval: time.Second, // 默认刷新间隔为 1 秒
			BlockOnFull:   false,       // 默认缓冲区满时阻塞
			DropOnFull:    true,        // 默认缓冲区满时丢弃
			FlushOnExit:   true,        // 默认退出时刷新缓冲区
		},
		RecoveryConfig: RecoveryConfig{
			Enable:           true,        // 默认启用恢复
			MaxRetries:       3,           // 默认最大重试次数为 3
			RetryInterval:    time.Second, // 默认重试间隔为 1 秒
			EnableStackTrace: true,        // 默认启用堆栈跟踪
			MaxStackSize:     4096,        // 默认最大堆栈大小为 4096
		},
		QueueConfig: QueueConfig{
			MaxSize:         1000,             // 默认队列最大大小为 1000
			BatchSize:       100,              // 默认批处理大小为 100
			Workers:         1,                // 默认工作协程数量为 1
			FlushInterval:   time.Second,      // 默认刷新间隔为 1 秒
			RetryCount:      3,                // 默认重试次数为 3
			RetryInterval:   time.Second,      // 默认重试间隔为 1 秒
			MaxBatchWait:    time.Second * 5,  // 默认最大批处理等待时间为 5 秒
			ShutdownTimeout: time.Second * 10, // 默认关闭超时时间为 10 秒
		},
		writers:   []io.Writer{},          // 添加 writers 字段
		ElkConfig: elk.DefaultElkConfig(), // 添加 ElkConfig 字段
	}

	// 如果存在配置文件，则使用配置文件中的值覆盖默认值
	if conf != nil && conf.Logger.Level != "" {
		// 设置日志级别
		switch conf.Logger.Level {
		case "debug":
			opts.Level = types.DebugLevel
		case "info":
			opts.Level = types.InfoLevel
		case "warn":
			opts.Level = types.WarnLevel
		case "error":
			opts.Level = types.ErrorLevel
		}

		// 其他配置项如果需要从配置文件读取，可以在这里添加
	}

	return opts
}

// Options 定义日志选项
type Options struct {
	ElasticURLs      []string       // Elasticsearch URL 列表
	ElasticIndex     string         // Elasticsearch 索引名称
	Development      bool           // 是否为开发模式
	Level            types.Level    // 日志级别
	ReportCaller     bool           // 是否报告调用者信息
	TimeFormat       string         // 时间格式
	MaxAge           time.Duration  // 日志保留时间
	RotationTime     time.Duration  // 日志轮转时间
	MaxSize          int64          // 单个日志文件最大大小(MB)
	Compress         bool           // 是否压缩旧日志
	OutputPaths      []string       // 输出路径
	ErrorOutputPaths []string       // 错误输出路径
	CompressConfig   CompressConfig // 压缩配置
	CleanupConfig    CleanupConfig  // 清理配置
	AsyncConfig      AsyncConfig    // 异步配置
	RecoveryConfig   RecoveryConfig // 恢复配置
	QueueConfig      QueueConfig    // 队列配置 (使用 queue.go 中的定义)
	writers          []io.Writer    // 添加 writers 字段
	ElkConfig        *elk.ElkConfig // 添加 ElkConfig 字段
}

// Option 定义选项函数类型
type Option func(*Options)

// WithLevel 设置日志级别
func WithLevel(level types.Level) Option {
	return func(o *Options) {
		o.Level = level // 设置日志级别
	}
}

// WithFormat 设置日志格式
func WithFormat(format string) Option {
	return func(o *Options) {
		o.TimeFormat = format // 设置时间格式
	}
}

// WithElastic 设置Elasticsearch配置
func WithElastic(urls []string, index string) Option {
	return func(o *Options) {
		o.ElasticURLs = urls   // Elasticsearch URL 列表
		o.ElasticIndex = index // Elasticsearch 索引名称
	}
}

// WithDevelopment 设置开发模式
func WithDevelopment(enabled bool) Option {
	return func(o *Options) {
		o.Development = enabled // 设置开发模式
	}
}

// WithRotation 设置日志轮转配置
func WithRotation(maxAge, rotationTime time.Duration, maxSize int64) Option {
	return func(o *Options) {
		o.MaxAge = maxAge             // 设置日志保留时间
		o.RotationTime = rotationTime // 设置日志轮转时间
		o.MaxSize = maxSize           // 设置单个日志文件最大大小
	}
}

// WithOutputPaths 设置输出路径
func WithOutputPaths(outputs []string) Option {
	return func(o *Options) {
		o.OutputPaths = outputs // 设置输出路径
	}
}

// WithErrorOutputPaths 设置错误输出路径
func WithErrorOutputPaths(errorOutputs []string) Option {
	return func(o *Options) {
		o.ErrorOutputPaths = errorOutputs // 设置错误输出路径
	}
}

// WithCompressConfig 设置压缩配置
func WithCompressConfig(config CompressConfig) Option {
	return func(o *Options) {
		o.CompressConfig = config // 设置压缩配置
	}
}

// WithCleanupConfig 设置清理配置
func WithCleanupConfig(config CleanupConfig) Option {
	return func(o *Options) {
		o.CleanupConfig = config // 设置清理配置
	}
}

// WithAsyncConfig 设置异步配置
func WithAsyncConfig(config AsyncConfig) Option {
	return func(o *Options) {
		o.AsyncConfig = config // 设置异步配置
	}
}

// WithRecoveryConfig 设置恢复配置
func WithRecoveryConfig(config RecoveryConfig) Option {
	return func(o *Options) {
		o.RecoveryConfig = config // 设置恢复配置
	}
}

// WithAsync 设置异步配置
func WithAsync(config AsyncConfig) Option {
	return func(o *Options) {
		o.AsyncConfig = config // 设置异步配置
	}
}

// WithCompress 设置压缩配置
func WithCompress(config CompressConfig) Option {
	return func(o *Options) {
		o.CompressConfig = config // 设置压缩配置
	}
}

// WithCleanup 设置清理配置
func WithCleanup(config CleanupConfig) Option {
	return func(o *Options) {
		o.CleanupConfig = config // 设置清理配置
	}
}

// WithRecovery 设置恢复配置
func WithRecovery(config RecoveryConfig) Option {
	return func(o *Options) {
		o.RecoveryConfig = config // 设置恢复配置
	}
}

// WithOutput 设置输出
func WithOutput(w io.Writer) Option {
	return func(o *Options) {
		if o.writers == nil { // 如果writers为空
			o.writers = make([]io.Writer, 0) // 创建新的writers
		}
		o.writers = append(o.writers, w) // 添加writer
	}
}

// WithElkConfig 设置 Elasticsearch 配置
func WithElkConfig(config *elk.ElkConfig) Option {
	return func(opts *Options) {
		opts.ElkConfig = config
	}
}

// AddWriter 添加输出writer
func (o *Options) AddWriter(w io.Writer) {
	if o.writers == nil {
		o.writers = make([]io.Writer, 0)
	}
	o.writers = append(o.writers, w)
}

// GetWriters 获取所有writers
func (o *Options) GetWriters() []io.Writer {
	return o.writers
}
