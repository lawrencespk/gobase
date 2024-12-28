package logrus

import (
	"context"
	"fmt"
	"io"
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
	asyncWriter *AsyncWriter    // 异步写入器
	fileManager *FileManager    // 文件管理器
	writers     []io.Writer     // 保存writers以便后续清理
	closed      bool            // 标记是否已关闭
	mu          sync.Mutex      // 保护 closed 字段
}

// NewLogger 创建新的logrus日志实例
func NewLogger(fm *FileManager, config QueueConfig, options *Options) (*logrusLogger, error) {
	l := &logrusLogger{
		logger:      logrus.New(),
		opts:        options,
		ctx:         context.Background(),
		fileManager: fm,
	}

	// 设置格式化器
	formatter := newFormatter(options)
	l.logger.SetFormatter(formatter)

	// 设置一个空的输出
	l.logger.SetOutput(io.Discard)

	// 处理所有的 writers
	var logPaths []string

	// 处理自定义 writers
	for _, w := range options.writers {
		if w != nil {
			// 如果启用了异步写入且不是标准输出/错误
			var writer io.Writer = w
			if options.AsyncConfig.Enable && w != os.Stdout && w != os.Stderr {
				if ww, ok := w.(Writer); ok {
					aw := NewAsyncWriter(ww, options.AsyncConfig)
					if aw != nil {
						writer = aw
						l.asyncWriter = aw
					}
				}
			}

			// 添加到 writers 列表
			l.writers = append(l.writers, writer)

			// 创建并添加 hook
			hook := &writerHook{
				writer:    writer,
				formatter: formatter,
			}
			l.logger.AddHook(hook)
		}
	}

	// 处理输出路径
	for _, path := range options.OutputPaths {
		var w io.Writer
		switch path {
		case "stdout":
			w = os.Stdout
		case "stderr":
			w = os.Stderr
		default:
			// 确保目录存在
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return nil, errors.NewFileOperationError("failed to create log directory", err)
			}

			file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return nil, errors.NewFileOperationError("failed to open log file", err)
			}
			w = file
			logPaths = append(logPaths, path)
		}

		// 如果启用了异步写入且不是标准输出/错误
		if options.AsyncConfig.Enable && w != os.Stdout && w != os.Stderr {
			if ww, ok := w.(Writer); ok {
				aw := NewAsyncWriter(ww, options.AsyncConfig)
				if aw != nil {
					w = aw
					l.asyncWriter = aw
				}
			}
		}

		// 添加到 writers 列表
		l.writers = append(l.writers, w)

		// 创建并添加 hook
		hook := &writerHook{
			writer:    w,
			formatter: formatter,
		}
		l.logger.AddHook(hook)
	}

	// 如果启用了压缩
	if options.CompressConfig.Enable && len(logPaths) > 0 {
		options.CompressConfig.LogPaths = logPaths
		l.compressor = NewLogCompressor(options.CompressConfig)
		if l.compressor != nil {
			l.compressor.Start()
		}
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
		logger:      l.logger,
		fields:      append([]types.Field{}, l.fields...),
		opts:        l.opts,
		ctx:         l.ctx,
		compressor:  l.compressor,
		asyncWriter: l.asyncWriter,
		fileManager: l.fileManager,
		writers:     l.writers,
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
		logger:      l.logger,
		fields:      append(l.fields, fields...),
		opts:        l.opts,
		ctx:         l.ctx,
		compressor:  l.compressor,
		asyncWriter: l.asyncWriter,
		fileManager: l.fileManager,
		writers:     l.writers,
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

// GetLevel 获取日志级别
func (l *logrusLogger) GetLevel() types.Level {
	return types.Level(l.logger.GetLevel()) // 取日志级别
}

// Sync 同步日志
func (l *logrusLogger) Sync() error {
	// 如果启用了异步写入，等待写入完成
	if l.asyncWriter != nil {
		if err := l.asyncWriter.Flush(); err != nil {
			return errors.NewLogFlushError("failed to flush async writer", err)
		}
	}

	// 同步所有 writers
	for _, w := range l.writers {
		if syncer, ok := w.(interface{ Sync() error }); ok {
			if err := syncer.Sync(); err != nil {
				return errors.NewLogFlushError("failed to sync writer", err)
			}
		}
	}

	// 如果启用了压缩，等待压缩完成
	if l.compressor != nil {
		// 给文件写入一些时间
		time.Sleep(time.Second * 2)

		// 确保所有文件都已经同步到磁盘
		for _, w := range l.writers {
			if file, ok := w.(*os.File); ok {
				if err := file.Sync(); err != nil {
					return errors.NewLogFlushError("failed to sync file", err)
				}
			}
		}

		// 停止压缩器
		l.compressor.Stop()

		// 等待压缩完成
		time.Sleep(time.Second)
	}

	return nil
}

// log 内部日志方法
func (l *logrusLogger) log(ctx context.Context, level logrus.Level, msg string, fields ...types.Field) {
	if !l.logger.IsLevelEnabled(level) {
		return
	}

	// 创建新的日志条目
	entry := &logrus.Entry{
		Logger:  l.logger,
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	// 合并字段
	allFields := make(logrus.Fields)

	// 添加基础字段
	allFields["timestamp"] = entry.Time.Format(time.RFC3339)
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
				"function": f.Name(),
				"file":     file,
				"line":     line,
			}
		}
	}

	// 设置所有字段
	entry.Data = allFields

	// 直接调用 Log 方法，这会触发所有的 hooks
	entry.Log(level, msg)
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

// Close 实现 io.Closer 接口
func (l *logrusLogger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	var errs []error

	// 1. 先停止异步写入器（它会关闭底层的文件句柄）
	if l.asyncWriter != nil {
		if err := l.asyncWriter.Stop(); err != nil {
			errs = append(errs, errors.Wrap(err, "async writer stop error"))
		}
	}

	// 2. 然后停止压缩器
	if l.compressor != nil {
		l.compressor.Stop()
		// 给压缩器一些时间完成最后的工作
		time.Sleep(time.Second)
	}

	// 3. 关闭其他非异步的 writers
	for _, w := range l.writers {
		// 跳过已经被异步写入器包装的 writer
		if _, ok := w.(*AsyncWriter); ok {
			continue
		}

		// 如果是文件，先同步再关闭
		if file, ok := w.(*os.File); ok {
			if err := file.Sync(); err != nil {
				errs = append(errs, errors.Wrap(err, "file sync error"))
			}
			if err := file.Close(); err != nil {
				errs = append(errs, errors.Wrap(err, "file close error"))
			}
		} else if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, errors.Wrap(err, "writer close error"))
			}
		}
	}

	// 清空 writers 列表
	l.writers = nil

	if len(errs) > 0 {
		return errors.NewSystemError("close errors", fmt.Errorf("%v", errs))
	}
	return nil
}

// GetLogrusLogger 返回原始的 logrus logger
func (l *logrusLogger) GetLogrusLogger() *logrus.Logger {
	return l.logger
}

// AddWriter 添加一个输出 writer
func (l *logrusLogger) AddWriter(w io.Writer) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return errors.NewSystemError("logger is closed", nil)
	}

	// 添加到 writers 列表
	l.writers = append(l.writers, w)

	// 添加到 logrus logger，使用相同的 formatter
	l.logger.AddHook(&writerHook{
		writer:    w,
		formatter: l.logger.Formatter,
	})

	return nil
}

// writerHook 实现 logrus.Hook 接口
type writerHook struct {
	writer    io.Writer
	formatter logrus.Formatter
	mu        sync.Mutex
}

func (h *writerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// 添加接口定义
type flusher interface {
	Flush() error
}

type syncer interface {
	Sync() error
}

func (h *writerHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return errors.NewLogFormatError("failed to format log entry", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := h.writer.Write(line); err != nil {
		return errors.NewLogWriteError("failed to write log entry", err)
	}

	// 如果 writer 支持 flush 和 sync
	if f, ok := h.writer.(flusher); ok {
		if err := f.Flush(); err != nil {
			return errors.NewLogFlushError("failed to flush log buffer", err)
		}
	}

	if s, ok := h.writer.(syncer); ok {
		if err := s.Sync(); err != nil {
			return errors.NewLogFlushError("failed to sync log file", err)
		}
	}

	return nil
}
