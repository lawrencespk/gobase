package mock

import (
	"context"
	"time"

	"gobase/pkg/client/redis"

	"github.com/stretchr/testify/mock"
)

// MockRedisClient Redis客户端的mock实现
type MockRedisClient struct {
	mock.Mock
}

// Get 实现 Get 方法
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// Set 实现 Set 方法
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// Del 实现 Del 方法
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// HGet 实现 HGet 方法
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

// HSet 实现 HSet 方法
func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

// HDel 实现 HDel 方法
func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

// LPush 实现 LPush 方法
func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

// LPop 实现 LPop 方法
func (m *MockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// SAdd 实现 SAdd 方法
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// SRem 实现 SRem 方法
func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZAdd 实现 ZAdd 方法
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZRem 实现 ZRem 方法
func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// TxPipeline 实现 TxPipeline 方法
func (m *MockRedisClient) TxPipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

// Eval 实现 Eval 方法
func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	mockArgs := m.Called(ctx, script, keys, args)
	return mockArgs.Get(0), mockArgs.Error(1)
}

// Close 实现 Close 方法
func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ping 实现 Ping 方法
func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// PoolStats 实现 PoolStats 方法
func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	if stats := args.Get(0); stats != nil {
		return stats.(*redis.PoolStats)
	}
	return nil
}

// Exists 实现 Exists 方法
func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Pool 实现 Pool 方法
func (m *MockRedisClient) Pool() redis.Pool {
	args := m.Called()
	return args.Get(0).(redis.Pool)
}

// Publish 实现 Publish 方法
func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

// Subscribe 实现 Subscribe 方法
func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(redis.PubSub)
}

// NewMockClient 创建一个新的mock客户端
func NewMockClient() *MockRedisClient {
	return &MockRedisClient{}
}
