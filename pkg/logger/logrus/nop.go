package logrus

import (
	"context"
	"time"

	"gobase/pkg/logger/types"
)

// NewNopLogger 创建一个新的空操作日志记录器
func NewNopLogger() types.Logger {
	return &nopLogger{}
}

// nopLogger 实现一个空操作的日志记录器
type nopLogger struct{}

// Debug implements types.Logger
func (l *nopLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {}

// Info implements types.Logger
func (l *nopLogger) Info(ctx context.Context, msg string, fields ...types.Field) {}

// Warn implements types.Logger
func (l *nopLogger) Warn(ctx context.Context, msg string, fields ...types.Field) {}

// Error implements types.Logger
func (l *nopLogger) Error(ctx context.Context, msg string, fields ...types.Field) {}

// Fatal implements types.Logger
func (l *nopLogger) Fatal(ctx context.Context, msg string, fields ...types.Field) {}

// Debugf implements types.Logger
func (l *nopLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}

// Infof implements types.Logger
func (l *nopLogger) Infof(ctx context.Context, format string, args ...interface{}) {}

// Warnf implements types.Logger
func (l *nopLogger) Warnf(ctx context.Context, format string, args ...interface{}) {}

// Errorf implements types.Logger
func (l *nopLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}

// Fatalf implements types.Logger
func (l *nopLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}

// WithFields implements types.Logger
func (l *nopLogger) WithFields(fields ...types.Field) types.Logger { return l }

// WithError implements types.Logger
func (l *nopLogger) WithError(err error) types.Logger { return l }

// WithContext implements types.Logger
func (l *nopLogger) WithContext(ctx context.Context) types.Logger { return l }

// WithTime implements types.Logger
func (l *nopLogger) WithTime(t time.Time) types.Logger { return l }

// WithCaller implements types.Logger
func (l *nopLogger) WithCaller(skip int) types.Logger { return l }

// SetLevel implements types.Logger
func (l *nopLogger) SetLevel(level types.Level) {}

// GetLevel implements types.Logger
func (l *nopLogger) GetLevel() types.Level { return types.InfoLevel }

// Sync implements types.Logger
func (l *nopLogger) Sync() error { return nil }

// ensure nopLogger implements types.Logger
var _ types.Logger = (*nopLogger)(nil)
