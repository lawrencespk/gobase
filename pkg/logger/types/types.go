package types

import (
	"context"
	"time"
)

// Fields 定义日志字段类型
type Fields map[string]interface{}

// Level 定义日志级别类型
type Level string

// 定义所有支持的日志级别
const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// Config 定义日志配置结构
type Config struct {
	Type       string `json:"type"`        // 日志类型：logrus 或 elk
	Level      Level  `json:"level"`       // 日志级别
	Format     string `json:"format"`      // 日志格式：text 或 json
	TimeFormat string `json:"time_format"` // 时间格式
	Caller     bool   `json:"caller"`      // 是否记录调用者信息

	// ELK 相关配置
	ElkEndpoint string `json:"elk_endpoint"` // ELK服务地址
	ElkIndex    string `json:"elk_index"`    // ELK索引名称
	ElkType     string `json:"elk_type"`     // ELK文档类型

	// 默认字段
	DefaultFields Fields `json:"default_fields"` // 默认携带的字段
}

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
	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
	WithTime(t time.Time) Logger

	// Hook相关方法
	AddHook(hook interface{}) error
}
