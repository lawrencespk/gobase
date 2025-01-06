package mock

import (
	"context"
	"gobase/pkg/logger/types"
)

// MockLogger 用于测试的 logger 实现
type MockLogger struct {
	types.Logger
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields ...types.Field) {}
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {}
