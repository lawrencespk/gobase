package mock

import (
	"context"
	"time"

	"gobase/pkg/client/redis"

	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

// 已实现的方法
func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	ret := m.Called(mock.Anything, script, keys, args)
	return ret.Get(0), ret.Error(1)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := []interface{}{ctx, keys}
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

// 需要补充实现的方法
func (m *MockRedisClient) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	ret := m.Called(ctx, key)
	return ret.String(0), ret.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ret := m.Called(ctx, key, value, expiration)
	return ret.Error(0)
}

func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	ret := m.Called(ctx, key, field)
	return ret.String(0), ret.Error(1)
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := []interface{}{ctx, key}
	args = append(args, values...)
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := []interface{}{ctx, key}
	for _, field := range fields {
		args = append(args, field)
	}
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := []interface{}{ctx, key}
	args = append(args, values...)
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	ret := m.Called(ctx, key)
	return ret.String(0), ret.Error(1)
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := []interface{}{ctx, key}
	args = append(args, members...)
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := []interface{}{ctx, key}
	args = append(args, members...)
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := []interface{}{ctx, key}
	for _, member := range members {
		args = append(args, member)
	}
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := []interface{}{ctx, key}
	args = append(args, members...)
	ret := m.Called(args...)
	return ret.Get(0).(int64), ret.Error(1)
}

func (m *MockRedisClient) TxPipeline() redis.Pipeline {
	ret := m.Called()
	return ret.Get(0).(redis.Pipeline)
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	ret := m.Called(ctx)
	return ret.Error(0)
}

func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	ret := m.Called()
	return ret.Get(0).(*redis.PoolStats)
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	ret := m.Called(ctx, key)
	return ret.Bool(0), ret.Error(1)
}

func (m *MockRedisClient) Pool() redis.Pool {
	ret := m.Called()
	return ret.Get(0).(redis.Pool)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	ret := m.Called(ctx, channel, message)
	return ret.Error(0)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := []interface{}{ctx}
	args = append(args, channels)
	ret := m.Called(args...)
	return ret.Get(0).(redis.PubSub)
}
