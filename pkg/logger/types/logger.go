package types

import (
	"context"
	"time"
)

// Field 定义日志字段
type Field struct {
	Key   string      // 字段名
	Value interface{} // 字段值
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{
		Key:   "error", // 字段名
		Value: err,     // 字段值
	}
}

// BasicLogger 提供基础的日志实现
type BasicLogger struct{}

func (l *BasicLogger) Debug(ctx context.Context, msg string, fields ...Field)         {}
func (l *BasicLogger) Info(ctx context.Context, msg string, fields ...Field)          {}
func (l *BasicLogger) Warn(ctx context.Context, msg string, fields ...Field)          {}
func (l *BasicLogger) Error(ctx context.Context, msg string, fields ...Field)         {}
func (l *BasicLogger) Fatal(ctx context.Context, msg string, fields ...Field)         {}
func (l *BasicLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}
func (l *BasicLogger) Infof(ctx context.Context, format string, args ...interface{})  {}
func (l *BasicLogger) Warnf(ctx context.Context, format string, args ...interface{})  {}
func (l *BasicLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (l *BasicLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}
func (l *BasicLogger) WithContext(ctx context.Context) Logger                         { return l }
func (l *BasicLogger) WithFields(fields ...Field) Logger                              { return l }
func (l *BasicLogger) WithError(err error) Logger                                     { return l }
func (l *BasicLogger) WithTime(t time.Time) Logger                                    { return l }
func (l *BasicLogger) WithCaller(skip int) Logger                                     { return l }
func (l *BasicLogger) SetLevel(level Level)                                           {}
func (l *BasicLogger) GetLevel() Level                                                { return 0 }
func (l *BasicLogger) Sync() error                                                    { return nil }

// Logger 定义日志接口
type Logger interface {
	// 基础日志方法
	Debug(ctx context.Context, msg string, fields ...Field) // 调试日志
	Info(ctx context.Context, msg string, fields ...Field)  // 信息日志
	Warn(ctx context.Context, msg string, fields ...Field)  // 警告日志
	Error(ctx context.Context, msg string, fields ...Field) // 错误日志
	Fatal(ctx context.Context, msg string, fields ...Field) // 严重日志

	// 格式化日志方法
	Debugf(ctx context.Context, format string, args ...interface{}) // 调试日志
	Infof(ctx context.Context, format string, args ...interface{})  // 信息日志
	Warnf(ctx context.Context, format string, args ...interface{})  // 警告日志
	Errorf(ctx context.Context, format string, args ...interface{}) // 错误日志
	Fatalf(ctx context.Context, format string, args ...interface{}) // 严重日志

	// 链式调用方法
	WithContext(ctx context.Context) Logger // 上下文
	WithFields(fields ...Field) Logger      // 字段
	WithError(err error) Logger             // 错误
	WithTime(t time.Time) Logger            // 时间
	WithCaller(skip int) Logger             // 调用者

	// 配置方法
	SetLevel(level Level) // 设置日志级别
	GetLevel() Level      // 获取日志级别

	// 同步方法
	Sync() error
}
