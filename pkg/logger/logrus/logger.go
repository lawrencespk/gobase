package logrus

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"gobase/pkg/logger/types"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger         *logrus.Logger  // logrus日志实例
	fields         []types.Field   // 字段
	opts           *Options        // 配置
	ctx            context.Context // 上下文
	compressor     *LogCompressor  // 压缩器
	cleaner        *LogCleaner     // 清理器
	asyncWriter    *AsyncWriter    // 异步写入器
	recoveryWriter *RecoveryWriter // 错误恢复写入器
	fileManager    *FileManager    // 文件管理器
	writeQueue     *WriteQueue     // 写入队列
}

// NewLogger 创建新的logrus日志实例
func NewLogger(fm *FileManager, config QueueConfig, options *Options) (*logrusLogger, error) {
	queue, err := NewWriteQueue(fm, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create write queue: %w", err)
	}

	l := &logrusLogger{
		logger:     logrus.New(),         // 创建新的logrus日志实例
		opts:       options,              // 确保 options 被正确赋值
		ctx:        context.Background(), // 上下文
		writeQueue: queue,                // 使用正确的变量名
	}

	// 配置logrus
	l.logger.SetLevel(convertLevel(options.Level)) // 设置日志级别
	l.logger.SetFormatter(newFormatter(options))   // 使用自定义格式化器
	l.logger.SetReportCaller(true)                 // 设置调用者

	// 设置输出
	l.logger.SetOutput(l.getLogOutput())

	// 添加hooks
	if len(options.ElasticURLs) > 0 {
		hook, err := newElasticHook(options)
		if err == nil {
			l.logger.AddHook(hook)
		}
	}

	// 初始化压缩器
	l.compressor = NewLogCompressor(options.CompressConfig)
	l.compressor.Start()

	// 初始化清理器
	l.cleaner = NewLogCleaner(options.CleanupConfig)
	l.cleaner.Start()

	// 初始化文件管理器
	l.fileManager = NewFileManager(FileOptions{
		BufferSize:    32 * 1024,              // 32KB
		FlushInterval: time.Second,            // 1秒
		MaxOpenFiles:  100,                    // 最大打开文件数
		DefaultPath:   options.OutputPaths[0], // 使用配置的第一个输出路径作为默认路径
	})

	// 初始化写入队列
	l.writeQueue, err = NewWriteQueue(l.fileManager, QueueConfig{
		MaxSize:       10000,       // 最大大小
		BatchSize:     100,         // 批量大小
		FlushInterval: time.Second, // 刷新间隔
		Workers:       2,           // 工作线程数
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize write queue: %w", err)
	}

	return l, nil
}

// getLogOutput 获取日志输出
func (l *logrusLogger) getLogOutput() io.Writer {
	var output io.Writer
	if len(l.opts.OutputPaths) == 0 {
		output = os.Stdout
	} else {
		queue, err := NewWriteQueue(l.fileManager, QueueConfig{
			MaxSize:       10000,       // 最大大小
			BatchSize:     100,         // 批量大小
			FlushInterval: time.Second, // 刷新间隔
			Workers:       2,           // 工作线程数
		})
		if err != nil {
			return os.Stdout
		}
		output = queue
	}

	// 添加错误恢复
	if l.opts.RecoveryConfig.Enable {
		l.recoveryWriter = NewRecoveryWriter(output, l.opts.RecoveryConfig)
		output = l.recoveryWriter
	}

	// 如果启用异步写入，包装输出
	if l.opts.AsyncConfig.Enable {
		l.asyncWriter = NewAsyncWriter(output, l.opts.AsyncConfig)
		return l.asyncWriter
	}

	return output
}

// WithTime 添加时间字段
func (l *logrusLogger) WithTime(t time.Time) types.Logger {
	newLogger := l.clone()
	newLogger.fields = append(newLogger.fields, types.Field{
		Key:   "time", // 时间字段名
		Value: t,      // 时间字段值
	})
	return newLogger
}

// WithCaller 添加调用者信息
func (l *logrusLogger) WithCaller(skip int) types.Logger {
	newLogger := l.clone()
	if pc, file, line, ok := runtime.Caller(skip); ok {
		f := runtime.FuncForPC(pc)
		newLogger.fields = append(newLogger.fields, types.Field{
			Key: "caller",
			Value: map[string]interface{}{
				"function": f.Name(), // 函数名
				"file":     file,     // 文件名
				"line":     line,     // 行号
			},
		})
	}
	return newLogger
}

// clone 克隆logger实例
func (l *logrusLogger) clone() *logrusLogger {
	return &logrusLogger{
		logger: l.logger,                             // logrus日志实例
		fields: append([]types.Field{}, l.fields...), // 字段
		opts:   l.opts,                               // 配置
		ctx:    l.ctx,                                // 上下文
	}
}

// Debug 调试日志
func (l *logrusLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.DebugLevel) {
		l.log(ctx, logrus.DebugLevel, msg, fields...)
	}
}

// Info 信息日志
func (l *logrusLogger) Info(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.InfoLevel) {
		l.log(ctx, logrus.InfoLevel, msg, fields...)
	}
}

// Warn 警告日志
func (l *logrusLogger) Warn(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.WarnLevel) {
		l.log(ctx, logrus.WarnLevel, msg, fields...)
	}
}

// Error 错误日志
func (l *logrusLogger) Error(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.ErrorLevel) {
		l.log(ctx, logrus.ErrorLevel, msg, fields...)
	}
}

// Fatal 严重日志
func (l *logrusLogger) Fatal(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.FatalLevel) {
		l.log(ctx, logrus.FatalLevel, msg, fields...)
	}
}

// Debugf 格式化调试日志
func (l *logrusLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.DebugLevel) {
		l.log(ctx, logrus.DebugLevel, fmt.Sprintf(format, args...))
	}
}

// Infof 格式化信息日志
func (l *logrusLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.InfoLevel) {
		l.log(ctx, logrus.InfoLevel, fmt.Sprintf(format, args...))
	}
}

// Warnf 格式化警告日志
func (l *logrusLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.WarnLevel) {
		l.log(ctx, logrus.WarnLevel, fmt.Sprintf(format, args...))
	}
}

// Errorf 格式化错误日志
func (l *logrusLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.ErrorLevel) {
		l.log(ctx, logrus.ErrorLevel, fmt.Sprintf(format, args...))
	}
}

// Fatalf 格式化严重日志
func (l *logrusLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.FatalLevel) {
		l.log(ctx, logrus.FatalLevel, fmt.Sprintf(format, args...))
	}
}

// WithContext 添加上下文
func (l *logrusLogger) WithContext(ctx context.Context) types.Logger {
	if ctx == nil {
		return l
	}
	fields := extractContextFields(ctx)
	return l.WithFields(fields...)
}

// WithFields 添加字段
func (l *logrusLogger) WithFields(fields ...types.Field) types.Logger {
	newLogger := &logrusLogger{
		logger: l.logger,                    // logrus日志实例
		fields: append(l.fields, fields...), // 字段
		opts:   l.opts,                      // 配置
	}
	return newLogger
}

// WithError 添加错误
func (l *logrusLogger) WithError(err error) types.Logger {
	return l.WithFields(types.Error(err))
}

// SetLevel 设置日志级别
func (l *logrusLogger) SetLevel(level types.Level) {
	l.logger.SetLevel(logrus.Level(level))
}

// GetLevel 获取日志级别
func (l *logrusLogger) GetLevel() types.Level {
	return types.Level(l.logger.GetLevel())
}

// Sync 同步日志
func (l *logrusLogger) Sync() error {
	var errs []error

	if l.compressor != nil {
		l.compressor.Stop()
	}
	if l.cleaner != nil {
		l.cleaner.Stop()
	}
	if l.asyncWriter != nil {
		if err := l.asyncWriter.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.fileManager != nil {
		if err := l.fileManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if l.writeQueue != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := l.writeQueue.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("sync errors: %v", errs)
	}
	return nil
}

// log 内部日志方法
func (l *logrusLogger) log(ctx context.Context, level logrus.Level, msg string, fields ...types.Field) {
	if !l.logger.IsLevelEnabled(level) {
		return
	}

	// 使用上下文创建日志条目
	entry := l.logger.WithContext(ctx)

	// 合并字段
	allFields := make(logrus.Fields)

	// 添加基础字段
	allFields["timestamp"] = time.Now().Format(time.RFC3339)
	allFields["level"] = level.String()

	// 添加上下文字段
	if ctx != nil {
		contextFields := extractContextFields(ctx)
		for _, field := range contextFields {
			allFields[field.Key] = field.Value
		}
	}

	// 添加预设字段
	for _, field := range l.fields {
		allFields[field.Key] = field.Value
	}

	// 添加当前字段
	for _, field := range fields {
		allFields[field.Key] = field.Value
	}

	// 添加调用者信息
	if l.opts.ReportCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			f := runtime.FuncForPC(pc)
			allFields["caller"] = map[string]interface{}{
				"function": f.Name(), // 函数名
				"file":     file,     // 文件名
				"line":     line,     // 行号
			}
		}
	}

	// 使用上下文和所有字段创建日志条目
	entry.WithFields(allFields).Log(level, msg)
}

// extractContextFields 从上下文中提取字段
func extractContextFields(ctx context.Context) []types.Field {
	fields := make([]types.Field, 0)

	// 从上下文中提取标准字段
	if ctx == nil {
		return fields
	}

	// 提取请求ID
	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields = append(fields, types.Field{Key: "request_id", Value: requestID})
	}

	// 提取追踪ID
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		fields = append(fields, types.Field{Key: "trace_id", Value: traceID})
	}

	// 提取用户ID
	if userID, ok := ctx.Value("user_id").(string); ok {
		fields = append(fields, types.Field{Key: "user_id", Value: userID})
	}

	return fields
}

// convertLevel 将 types.Level 转换为 logrus.Level
func convertLevel(level types.Level) logrus.Level {
	switch level {
	case types.DebugLevel: // 调试级别
		return logrus.DebugLevel
	case types.InfoLevel: // 信息级别
		return logrus.InfoLevel
	case types.WarnLevel:
		return logrus.WarnLevel
	case types.ErrorLevel: // 错误级别
		return logrus.ErrorLevel
	case types.FatalLevel: // 严重级别
		return logrus.FatalLevel
	default: // 默认信息级别
		return logrus.InfoLevel
	}
}
