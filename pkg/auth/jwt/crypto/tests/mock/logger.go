package mock

import (
	"context"
	"gobase/pkg/logger/types"
	"time"
)

type MockLogger struct {
	level      types.Level   // 存储日志级别
	withCaller int           // 存储调用者信息级别
	fields     []types.Field // 存储日志字段
	timestamp  time.Time     // 存储时间戳
}

func NewMockLogger() types.Logger {
	return &MockLogger{
		level:      types.InfoLevel, // 默认使用 Info 级别
		withCaller: 0,               // 默认不记录调用者信息
		fields:     make([]types.Field, 0),
		timestamp:  time.Now(),
	}
}

// Debug 实现 Debug 级别日志
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {}

// Info 实现 Info 级别日志
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...types.Field) {}

// Warn 实现 Warn 级别日志
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...types.Field) {}

// Error 实现 Error 级别日志
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...types.Field) {}

// Fatal 实现 Fatal 级别日志
func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...types.Field) {}

// Debugf 实现格式化的 Debug 级别日志
func (m *MockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}

// Infof 实现格式化的 Info 级别日志
func (m *MockLogger) Infof(ctx context.Context, format string, args ...interface{}) {}

// Warnf 实现格式化的 Warn 级别日志
func (m *MockLogger) Warnf(ctx context.Context, format string, args ...interface{}) {}

// Errorf 实现格式化的 Error 级别日志
func (m *MockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}

// Fatalf 实现格式化的 Fatal 级别日志
func (m *MockLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}

// GetLevel 返回日志级别
func (m *MockLogger) GetLevel() types.Level {
	return m.level
}

// SetLevel 设置日志级别
func (m *MockLogger) SetLevel(level types.Level) {
	m.level = level
}

// Sync 同步日志
func (m *MockLogger) Sync() error {
	return nil // Mock实现，直接返回nil
}

// WithCaller 设置调用者信息级别
func (m *MockLogger) WithCaller(skip int) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: skip,
		fields:     append([]types.Field{}, m.fields...),
	}
	return newLogger
}

// WithField 添加单个字段
func (m *MockLogger) WithField(key string, value interface{}) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: m.withCaller,
		fields:     append(m.fields, types.Field{Key: key, Value: value}),
	}
	return newLogger
}

// WithFields 添加多个字段
func (m *MockLogger) WithFields(fields ...types.Field) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: m.withCaller,
		fields:     append(m.fields, fields...),
	}
	return newLogger
}

// WithError 添加错误信息
func (m *MockLogger) WithError(err error) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: m.withCaller,
		fields:     append(m.fields, types.Field{Key: "error", Value: err}),
	}
	return newLogger
}

// WithContext 添加上下文信息
func (m *MockLogger) WithContext(ctx context.Context) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: m.withCaller,
		fields:     append([]types.Field{}, m.fields...),
	}
	return newLogger
}

// WithTime 设置日志时间
func (m *MockLogger) WithTime(t time.Time) types.Logger {
	newLogger := &MockLogger{
		level:      m.level,
		withCaller: m.withCaller,
		fields:     append([]types.Field{}, m.fields...),
		timestamp:  t,
	}
	return newLogger
}
