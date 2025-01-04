package testutils

import (
	"context"

	"gobase/pkg/logger/types"
)

// NoopLogger 是一个空实现的logger
type NoopLogger struct{}

// Debug level
func (n *NoopLogger) Debug(ctx context.Context, msg string, fields ...types.Field)   {}
func (n *NoopLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}

// Info level
func (n *NoopLogger) Info(ctx context.Context, msg string, fields ...types.Field)   {}
func (n *NoopLogger) Infof(ctx context.Context, format string, args ...interface{}) {}

// Warn level
func (n *NoopLogger) Warn(ctx context.Context, msg string, fields ...types.Field)   {}
func (n *NoopLogger) Warnf(ctx context.Context, format string, args ...interface{}) {}

// Error level
func (n *NoopLogger) Error(ctx context.Context, msg string, fields ...types.Field)   {}
func (n *NoopLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}

// Fatal level
func (n *NoopLogger) Fatal(ctx context.Context, msg string, fields ...types.Field)   {}
func (n *NoopLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}

// Level operations
func (n *NoopLogger) GetLevel() types.Level      { return types.InfoLevel }
func (n *NoopLogger) SetLevel(level types.Level) {}

// Factory functions
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func NewNoopLoggerWithLevel(level types.Level) *NoopLogger {
	n := &NoopLogger{}
	n.SetLevel(level)
	return n
}

func NewNoopLoggerWithContext(ctx context.Context) *NoopLogger {
	return &NoopLogger{}
}
