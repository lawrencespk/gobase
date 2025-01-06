package mock

import (
	"context"
	"time"

	"gobase/pkg/client/redis"

	"github.com/stretchr/testify/mock"
)

type MockPipeline struct {
	mock.Mock
}

// 基础操作
func (m *MockPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockPipeline) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockPipeline) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// Hash操作
func (m *MockPipeline) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPipeline) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockPipeline) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

// Set操作
func (m *MockPipeline) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPipeline) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZSet操作
func (m *MockPipeline) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPipeline) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// 管道控制
func (m *MockPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	if cmders := args.Get(0); cmders != nil {
		return cmders.([]redis.Cmder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPipeline) Close() error {
	args := m.Called()
	return args.Error(0)
}

// 过期时间操作
func (m *MockPipeline) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	args := m.Called(ctx, key, tm)
	return args.Bool(0), args.Error(1)
}

func NewMockPipeline() *MockPipeline {
	return &MockPipeline{}
}
