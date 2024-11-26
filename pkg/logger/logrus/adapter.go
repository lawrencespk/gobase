package logrus

import (
	"context"
	"fmt"
	"time"

	"gobase/pkg/logger/types"

	"github.com/sirupsen/logrus"
)

// Adapter logrus适配器
type Adapter struct {
	log    *logrus.Logger
	fields types.Fields
}

// NewAdapter 创建新的 logrus 适配器
func NewAdapter(cfg types.Config) (types.Logger, error) {
	// 创建 logrus 实例
	log := logrus.New()

	// 配置日志级别
	level, err := logrus.ParseLevel(string(cfg.Level))
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	// 配置日志格式
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: cfg.TimeFormat,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: cfg.TimeFormat,
			FullTimestamp:   true,
		})
	}

	// 是否记录调用者信息
	log.SetReportCaller(cfg.Caller)

	return &Adapter{
		log:    log,
		fields: cfg.DefaultFields,
	}, nil
}

// Debug 实现 Debug 级别日志
func (a *Adapter) Debug(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Debug(args...)
}

// Debugf 实现格式化的 Debug 级别日志
func (a *Adapter) Debugf(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Debugf(format, args...)
}

// Info 实现 Info 级别日志
func (a *Adapter) Info(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Info(args...)
}

// Infof 实现格式化的 Info 级别日志
func (a *Adapter) Infof(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Infof(format, args...)
}

// Warn 实现 Warn 级别日志
func (a *Adapter) Warn(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Warn(args...)
}

// Warnf 实现格式化的 Warn 级别日志
func (a *Adapter) Warnf(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Warnf(format, args...)
}

// Error 实现 Error 级别日志
func (a *Adapter) Error(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Error(args...)
}

// Errorf 实现格式化的 Error 级别日志
func (a *Adapter) Errorf(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Errorf(format, args...)
}

// Fatal 实现 Fatal 级别日志
func (a *Adapter) Fatal(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Fatal(args...)
}

// Fatalf 实现格式化的 Fatal 级别日志
func (a *Adapter) Fatalf(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Fatalf(format, args...)
}

// Panic 实现 Panic 级别日志
func (a *Adapter) Panic(args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Panic(args...)
}

// Panicf 实现格式化的 Panic 级别日志
func (a *Adapter) Panicf(format string, args ...interface{}) {
	a.log.WithFields(logrus.Fields(a.fields)).Panicf(format, args...)
}

// WithFields 实现字段追加
func (a *Adapter) WithFields(fields types.Fields) types.Logger {
	newAdapter := *a
	newFields := make(types.Fields)

	// 复制现有字段
	for k, v := range a.fields {
		newFields[k] = v
	}
	// 添加新字段
	for k, v := range fields {
		newFields[k] = v
	}

	newAdapter.fields = newFields
	return &newAdapter
}

// WithContext 实现上下文追加
func (a *Adapter) WithContext(ctx context.Context) types.Logger {
	return a.WithFields(types.Fields{
		"context": ctx,
	})
}

// WithError 实现错误信息追加
func (a *Adapter) WithError(err error) types.Logger {
	return a.WithFields(types.Fields{
		"error": err.Error(),
	})
}

// WithTime 实现时间信息追加
func (a *Adapter) WithTime(t time.Time) types.Logger {
	return a.WithFields(types.Fields{
		"custom_time": t.Format(time.RFC3339),
	})
}

// AddHook 添加 logrus hook
func (a *Adapter) AddHook(hook interface{}) error {
	if h, ok := hook.(logrus.Hook); ok {
		a.log.AddHook(h)
		return nil
	}
	return fmt.Errorf("invalid hook type: %T", hook)
}
