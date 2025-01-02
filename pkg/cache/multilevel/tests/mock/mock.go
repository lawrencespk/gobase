package mock

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
)

// MockRedisClient 实现 redis.Client 接口的模拟客户端
type MockRedisClient struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

// Get 实现获取数据，如果key不存在返回 RedisKeyNotFoundError
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", errors.NewRedisKeyNotFoundError("key not found", nil)
}

// Set 实现存储数据
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	strVal, ok := value.(string)
	if !ok {
		return errors.NewInvalidParamsError("value must be string", nil)
	}
	m.data[key] = strVal
	return nil
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Del 实现删除数据
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			count++
		}
	}
	return count, nil
}

// 连接管理
func (m *MockRedisClient) Close() error {
	m.data = nil
	return nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	return nil
}

// 其他必需的方法
func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.data[key]
	return ok, nil
}

func (m *MockRedisClient) Pipeline() redis.Pipeline {
	return nil
}

func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	return &redis.PoolStats{}
}

func (m *MockRedisClient) Pool() redis.Pool {
	return nil
}

// 在 MockRedisClient 结构体中添加 Eval 方法
func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	// 由于这是测试用的 mock，我们可以返回一个简单的实现
	return nil, nil
}

// 我还注意到缺少了 Publish 和 Subscribe 方法，也一并添加
func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return nil
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	return nil
}

// 在 MockRedisClient 结构体中添加 Hash 操作相关方法
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return "", nil
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return 0, nil
}

// 在 MockRedisClient 结构体中添加 List 操作相关方法
func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	return "", nil
}

// 在 MockRedisClient 结构体中添加 Set 操作相关方法
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

// 在 MockRedisClient 结构体中添加 TxPipeline 方法
func (m *MockRedisClient) TxPipeline() redis.Pipeline {
	return nil
}

// 在 MockRedisClient 结构体中添加 ZSet 操作相关方法
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return 0, nil
}

// MockLogger 实现 types.Logger 接口的模拟日志器
type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// 实现 Logger 接口的所有方法
func (l *MockLogger) Debug(ctx context.Context, msg string, fields ...types.Field)   {}
func (l *MockLogger) Info(ctx context.Context, msg string, fields ...types.Field)    {}
func (l *MockLogger) Warn(ctx context.Context, msg string, fields ...types.Field)    {}
func (l *MockLogger) Error(ctx context.Context, msg string, fields ...types.Field)   {}
func (l *MockLogger) Fatal(ctx context.Context, msg string, fields ...types.Field)   {}
func (l *MockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}
func (l *MockLogger) Infof(ctx context.Context, format string, args ...interface{})  {}
func (l *MockLogger) Warnf(ctx context.Context, format string, args ...interface{})  {}
func (l *MockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (l *MockLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}

func (l *MockLogger) WithFields(fields ...types.Field) types.Logger { return l }
func (l *MockLogger) WithError(err error) types.Logger              { return l }
func (l *MockLogger) WithContext(ctx context.Context) types.Logger  { return l }
func (l *MockLogger) WithTime(t time.Time) types.Logger             { return l }
func (l *MockLogger) WithCaller(skip int) types.Logger              { return l }

func (l *MockLogger) SetLevel(level types.Level) {}
func (l *MockLogger) GetLevel() types.Level      { return types.InfoLevel }
func (l *MockLogger) Sync() error                { return nil }
