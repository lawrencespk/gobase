package mock

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockLimiter 模拟限流器
type MockLimiter struct {
	mock.Mock
}

// Allow 模拟Allow方法
func (m *MockLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Error(1)
}

// AllowN 模拟AllowN方法
func (m *MockLimiter) AllowN(ctx context.Context, key string, n int64, limit int64, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, n, limit, window)
	return args.Bool(0), args.Error(1)
}

// Wait 模拟Wait方法
func (m *MockLimiter) Wait(ctx context.Context, key string, limit int64, window time.Duration) error {
	args := m.Called(ctx, key, limit, window)
	return args.Error(0)
}

// Reset 模拟Reset方法
func (m *MockLimiter) Reset(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
