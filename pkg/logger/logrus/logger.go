package logrus

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger      *logrus.Logger  // logrus日志实例
	fields      []types.Field   // 字段
	opts        *Options        // 配置
	ctx         context.Context // 上下文
	compressor  *LogCompressor  // 压缩器
	cleaner     *LogCleaner     // 清理器
	asyncWriter *AsyncWriter    // 异步写入器
	fileManager *FileManager    // 文件管理器
	writeQueue  *WriteQueue     // 写入队列
	writers     []io.Writer     // 保存writers以便后续清理
	closed      bool            // 标记是否已关闭
	mu          sync.Mutex      // 保护 closed 字段
}

// NewLogger 创建新的logrus日志实例
func NewLogger(fm *FileManager, config QueueConfig, options *Options) (*logrusLogger, error) {
	l := &logrusLogger{
		logger:      logrus.New(),         // 创建新的logrus日志实例
		opts:        options,              // 配置
		ctx:         context.Background(), // 上下文
		fileManager: fm,                   // 文件管理器
	}

	// 设置多输出
	var writers []io.Writer

	// 处理自定义writers
	if len(options.writers) > 0 {
		writers = append(writers, options.writers...) // 添加自定义写入器
	}

	// 处理输出路径
	if len(options.OutputPaths) > 0 {
		for _, path := range options.OutputPaths { // 遍历输出路径
			var w io.Writer // 写入器
			switch path {
			case "stdout": // 标准输出
				w = os.Stdout
			case "stderr": // 标准错误
				w = os.Stderr
			default: // 默认输出到文件
				// 确保目录存在
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					return nil, errors.NewFileOperationError("failed to create log directory", err) // 创建日志目录失败
				}

				file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return nil, errors.NewFileOperationError("failed to open log file", err) // 打开日志文件失败
				}
				w = file
			}
			writers = append(writers, w) // 添加写入器
		}
	}

	// 如果没有配置任何输出，默认输出到标准输出
	if len(writers) == 0 {
		writers = append(writers, os.Stdout) // 默认输出到标准输出
	}

	// 设置输出
	l.logger.SetOutput(io.MultiWriter(writers...)) // 设置输出
	l.writers = writers                            // 设置写入器

	// 初始化写入队列
	queue, err := NewWriteQueue(fm, config) // 创建写入队列
	if err != nil {
		return nil, errors.Wrap(err, "failed to create write queue")
	}
	l.writeQueue = queue // 设置写入队列

	// 配置 logrus
	l.logger.SetLevel(convertLevel(options.Level)) // 设置日志级别
	l.logger.SetFormatter(newFormatter(options))   // 设置日志格式
	l.logger.SetReportCaller(true)                 // 设置调用者信息

	// 初始化压缩器
	if options.CompressConfig.Enable {
		// 确保压缩配置包含日志路径
		compressConfig := options.CompressConfig      // 压缩配置
		compressConfig.LogPaths = options.OutputPaths // 日志路径

		compressor := NewLogCompressor(compressConfig)                    // 创建压缩器
		l.compressor = compressor                                         // 设置压缩器
		compressor.Start()                                                // 启动压缩器
		log.Printf("Compressor started with config: %+v", compressConfig) // 打印压缩器启动成功
	}

	// 初始化清理器
	if options.CleanupConfig.Enable { // 如果启用清理
		cleaner := NewLogCleaner(options.CleanupConfig) // 创建清理器
		l.cleaner = cleaner                             // 设置清理器
		cleaner.Start()                                 // 启动清理器
	}

	return l, nil
}

// WithTime 添加时间字段
func (l *logrusLogger) WithTime(t time.Time) types.Logger {
	newLogger := l.clone()                                   // 克隆logger实例
	newLogger.fields = append(newLogger.fields, types.Field{ // 添加时间字段
		Key:   "time", // 时间字段名
		Value: t,      // 时间字段值
	})
	return newLogger
}

// WithCaller 添加调用者信息
func (l *logrusLogger) WithCaller(skip int) types.Logger {
	newLogger := l.clone()                              // 克隆logger实例
	if pc, file, line, ok := runtime.Caller(skip); ok { // 获取调用者信息
		f := runtime.FuncForPC(pc)                               // 获取函数信息
		newLogger.fields = append(newLogger.fields, types.Field{ // 添加调用者信息
			Key: "caller", // 调用者字段名
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
	if l.logger.IsLevelEnabled(logrus.DebugLevel) { // 如果启用调试级别
		l.log(ctx, logrus.DebugLevel, msg, fields...) // 记录调试日志
	}
}

// Info 信息日志
func (l *logrusLogger) Info(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.InfoLevel) { // 如果启用信息级别
		l.log(ctx, logrus.InfoLevel, msg, fields...) // 记录信息日志
	}
}

// Warn 警告日志
func (l *logrusLogger) Warn(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.WarnLevel) { // 如果启用警告级别
		l.log(ctx, logrus.WarnLevel, msg, fields...) // 记录警告日志
	}
}

// Error 错误日志
func (l *logrusLogger) Error(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.ErrorLevel) { // 如果启用错误级别
		l.log(ctx, logrus.ErrorLevel, msg, fields...) // 记录错误日志
	}
}

// Fatal 严重日志
func (l *logrusLogger) Fatal(ctx context.Context, msg string, fields ...types.Field) {
	if l.logger.IsLevelEnabled(logrus.FatalLevel) { // 如果启用严重级别
		l.log(ctx, logrus.FatalLevel, msg, fields...) // 记录严重日志
	}
}

// Debugf 格式化调试日志
func (l *logrusLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.DebugLevel) { // 如果启用调试级别
		l.log(ctx, logrus.DebugLevel, fmt.Sprintf(format, args...)) // 记录调试日志
	}
}

// Infof 格式化信息日志
func (l *logrusLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.InfoLevel) { // 如果启用信息级别
		l.log(ctx, logrus.InfoLevel, fmt.Sprintf(format, args...)) // 记录信息日志
	}
}

// Warnf 格式化警告日志
func (l *logrusLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.WarnLevel) { // 如果启用警告级别
		l.log(ctx, logrus.WarnLevel, fmt.Sprintf(format, args...)) // 记录警告日志
	}
}

// Errorf 格式化错误日志
func (l *logrusLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.ErrorLevel) { // 如果启用错误级别
		l.log(ctx, logrus.ErrorLevel, fmt.Sprintf(format, args...)) // 记录错误日志
	}
}

// Fatalf 格式化严重日志
func (l *logrusLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	if l.logger.IsLevelEnabled(logrus.FatalLevel) { // 如果启用严重级别
		l.log(ctx, logrus.FatalLevel, fmt.Sprintf(format, args...)) // 记录严重日志
	}
}

// WithContext 添加上下文
func (l *logrusLogger) WithContext(ctx context.Context) types.Logger {
	if ctx == nil { // 如果上下文为空
		return l
	}
	fields := extractContextFields(ctx) // 提取上下文字段
	return l.WithFields(fields...)      // 添加字段
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
	return l.WithFields(types.Error(err)) // 添加错误字段
}

// SetLevel 设置日志级别
func (l *logrusLogger) SetLevel(level types.Level) {
	l.logger.SetLevel(logrus.Level(level)) // 设置日志级别
}

// GetLevel 取日志级别
func (l *logrusLogger) GetLevel() types.Level {
	return types.Level(l.logger.GetLevel()) // 取日志级别
}

// Sync 同步日志
func (l *logrusLogger) Sync() error {
	l.mu.Lock()   // 锁定
	if l.closed { // 如果已关闭
		l.mu.Unlock() // 解锁
		return nil
	}
	l.mu.Unlock() // 解锁

	var errs []error

	// 停止压缩器
	if l.compressor != nil {
		l.compressor.Stop() // 停止压缩器
	}

	// 停止清理器
	if l.cleaner != nil {
		l.cleaner.Stop() // 停止清理器
	}

	// 停止异步写入器
	if l.asyncWriter != nil {
		if err := l.asyncWriter.Stop(); err != nil { // 停止异步写入器
			errs = append(errs, errors.Wrap(err, "failed to stop async writer")) // 添加错误
		}
	}

	// 关闭文件管理器
	if l.fileManager != nil {
		if err := l.fileManager.Close(); err != nil { // 关闭文件管理器
			errs = append(errs, errors.Wrap(err, "failed to close file manager")) // 添加错误
		}
	}

	// 关闭写入队列
	if l.writeQueue != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := l.writeQueue.Close(ctx); err != nil { // 关闭写入队列
			errs = append(errs, errors.Wrap(err, "failed to close write queue")) // 添加错误
		}
	}

	if len(errs) > 0 {
		return errors.NewSystemError("sync errors", fmt.Errorf("%v", errs)) // 返回同步错误
	}
	return nil
}

// log 内部日志方法
func (l *logrusLogger) log(ctx context.Context, level logrus.Level, msg string, fields ...types.Field) {
	if !l.logger.IsLevelEnabled(level) { // 如果未启用日志级别
		return
	}

	// 使用上下文创建日志条目
	entry := l.logger.WithContext(ctx)

	// 合并字段
	allFields := make(logrus.Fields)

	// 添加基础字段
	allFields["timestamp"] = time.Now().Format(time.RFC3339) // 添加时间字段
	allFields["level"] = level.String()                      // 添加日志级别字段

	// 添加上下文字段
	if ctx != nil {
		contextFields := extractContextFields(ctx) // 提取上下文字段
		for _, field := range contextFields {
			allFields[field.Key] = field.Value // 添加上下文字段
		}
	}

	// 添加预设字段
	for _, field := range l.fields {
		allFields[field.Key] = field.Value // 添加预设字段
	}

	// 添加当前字段
	for _, field := range fields {
		allFields[field.Key] = field.Value // 添加当前字段
	}

	// 添加调用者信息
	if l.opts.ReportCaller {
		if pc, file, line, ok := runtime.Caller(2); ok { // 获取调用者信息
			f := runtime.FuncForPC(pc) // 获取函数信息
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
	fields := make([]types.Field, 0) // 创建字段列表

	// 从上下文中提取标准字段
	if ctx == nil {
		return fields
	}

	// 提取请求ID
	if requestID, ok := ctx.Value("request_id").(string); ok { // 提取请求ID
		fields = append(fields, types.Field{Key: "request_id", Value: requestID}) // 添加请求ID字段
	}

	// 提取追踪ID
	if traceID, ok := ctx.Value("trace_id").(string); ok { // 提取追踪ID
		fields = append(fields, types.Field{Key: "trace_id", Value: traceID}) // 添加追踪ID字段
	}

	// 提取用户ID
	if userID, ok := ctx.Value("user_id").(string); ok { // 提取用户ID
		fields = append(fields, types.Field{Key: "user_id", Value: userID}) // 添加用户ID字段
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
	case types.WarnLevel: // 警告级别
		return logrus.WarnLevel
	case types.ErrorLevel: // 错误级别
		return logrus.ErrorLevel
	case types.FatalLevel: // 严重级别
		return logrus.FatalLevel
	default: // 默认信息级别
		return logrus.InfoLevel
	}
}

// Close 实现 io.Closer 接口
func (l *logrusLogger) Close() error {
	l.mu.Lock()   // 锁定
	if l.closed { // 如果已关闭
		l.mu.Unlock() // 解锁
		return nil
	}
	l.closed = true // 设置已关闭
	l.mu.Unlock()   // 解锁

	var errs []error

	// 先同步日志
	if err := l.Sync(); err != nil { // 同步日志
		errs = append(errs, errors.Wrap(err, "sync error")) // 添加错误
	}

	// 关闭所有writers
	for _, w := range l.writers {
		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil { // 关闭writer
				errs = append(errs, errors.Wrap(err, "writer close error")) // 添加错误
			}
		}
	}

	// 清空writers列表
	l.writers = nil

	if len(errs) > 0 {
		return errors.NewSystemError("close errors", fmt.Errorf("%v", errs)) // 返回关闭错误
	}
	return nil
}
