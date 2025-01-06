package mock

import (
	"context"
	"time"

	"gobase/pkg/client/redis"
)

type MockClient struct {
	err   error
	store map[string]string
}

func NewMockClient() redis.Client {
	return &MockClient{
		store: make(map[string]string),
	}
}

func (m *MockClient) SetError(err error) {
	m.err = err
}

// 基础操作
func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if value, exists := m.store[key]; exists {
		return value, nil
	}
	return "", redis.ErrNil
}

func (m *MockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.store[key] = value.(string)
	return nil
}

func (m *MockClient) Del(ctx context.Context, keys ...string) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	var deleted int64
	for _, key := range keys {
		if _, exists := m.store[key]; exists {
			delete(m.store, key)
			deleted++
		}
	}
	return deleted, nil
}

// Hash操作
func (m *MockClient) HGet(ctx context.Context, key, field string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "", nil
}

func (m *MockClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *MockClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

// List操作
func (m *MockClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *MockClient) LPop(ctx context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "", nil
}

// Set操作
func (m *MockClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *MockClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

// ZSet操作
func (m *MockClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

func (m *MockClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 1, nil
}

// 事务操作
func (m *MockClient) TxPipeline() redis.Pipeline {
	return &MockPipeline{err: m.err}
}

// Lua脚本
func (m *MockClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

// 连接管理
func (m *MockClient) Close() error {
	return m.err
}

func (m *MockClient) Ping(ctx context.Context) error {
	return m.err
}

// 监控相关
func (m *MockClient) PoolStats() *redis.PoolStats {
	return &redis.PoolStats{}
}

// 缓存管理
func (m *MockClient) Exists(ctx context.Context, key string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return true, nil
}

// 连接池管理
func (m *MockClient) Pool() redis.Pool {
	return &MockPool{err: m.err}
}

// Publish/Subscribe 操作
func (m *MockClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return m.err
}

func (m *MockClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	return &MockPubSub{err: m.err}
}

// MockPool 实现 Pool 接口
type MockPool struct {
	err error
}

// 实现 Pool 接口的方法
func (p *MockPool) Close() error {
	return p.err
}

// 实现 Pool 接口的 Stats 方法
func (p *MockPool) Stats() *redis.PoolStats {
	return &redis.PoolStats{}
}

// MockPubSub 实现 PubSub 接口
type MockPubSub struct {
	err error
}

func (ps *MockPubSub) ReceiveMessage(ctx context.Context) (*redis.Message, error) {
	if ps.err != nil {
		return nil, ps.err
	}
	return &redis.Message{}, nil
}

func (ps *MockPubSub) Close() error {
	return ps.err
}
