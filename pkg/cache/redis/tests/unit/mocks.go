package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"time"
)

type mockRedisClient struct {
	getErr error
	setErr error
	delErr error
	data   map[string]string
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
	}
}

// 基础操作
func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return "", errors.NewRedisKeyNotFoundError("key not found", nil)
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value.(string)
	return nil
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	if m.delErr != nil {
		return 0, m.delErr
	}
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			count++
		}
	}
	return count, nil
}

// Hash操作
func (m *mockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return "", nil
}

func (m *mockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return 0, nil
}

func (m *mockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return 0, nil
}

// List操作
func (m *mockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return 0, nil
}

func (m *mockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	return "", nil
}

// Set操作
func (m *mockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

func (m *mockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

// ZSet操作
func (m *mockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return 0, nil
}

func (m *mockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

// 事务操作
func (m *mockRedisClient) TxPipeline() redis.Pipeline {
	return nil
}

// Lua脚本
func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

// 连接管理
func (m *mockRedisClient) Close() error {
	return nil
}

func (m *mockRedisClient) Ping(ctx context.Context) error {
	return nil
}

// 监控相关
func (m *mockRedisClient) PoolStats() *redis.PoolStats {
	return nil
}

// 缓存管理
func (m *mockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// 连接池管理
func (m *mockRedisClient) Pool() redis.Pool {
	return nil
}

// Publish/Subscribe 操作
func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return nil
}

func (m *mockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...types.Field)   {}
func (m *mockLogger) Info(ctx context.Context, msg string, fields ...types.Field)    {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...types.Field)    {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...types.Field)   {}
func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...types.Field)   {}
func (m *mockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}
func (m *mockLogger) Infof(ctx context.Context, format string, args ...interface{})  {}
func (m *mockLogger) Warnf(ctx context.Context, format string, args ...interface{})  {}
func (m *mockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (m *mockLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}
func (m *mockLogger) WithContext(ctx context.Context) types.Logger                   { return m }
func (m *mockLogger) WithFields(fields ...types.Field) types.Logger                  { return m }
func (m *mockLogger) WithError(err error) types.Logger                               { return m }
func (m *mockLogger) WithTime(t time.Time) types.Logger                              { return m }
func (m *mockLogger) WithCaller(skip int) types.Logger                               { return m }
func (m *mockLogger) SetLevel(level types.Level)                                     {}
func (m *mockLogger) GetLevel() types.Level                                          { return 0 }
func (m *mockLogger) Sync() error                                                    { return nil }
