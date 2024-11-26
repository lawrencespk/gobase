package logger

import (
	"context"
	"time"

	"gobase/pkg/logger/types"
)

// Logger 定义统一的日志接口
type Logger interface {
	// 基础日志方法
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	// 格式化日志方法
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	// 结构化日志方法
	WithFields(fields types.Fields) Logger
	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
	WithTime(t time.Time) Logger

	// Hook相关方法
	AddHook(hook interface{}) error
}
