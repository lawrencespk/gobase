package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

var (
	// 定义 Redis 错误常量
	errClosed = errors.NewError(codes.RedisConnError, "redis connection closed", nil)
)

// mockRedisClient 是redis.Client的mock实现
type mockRedisClient struct {
	mock.Mock
}

// ZRangeBy 定义了 ZRangeByScore 的范围选项
type ZRangeBy struct {
	Min    string // 最小分数
	Max    string // 最大分数
	Offset int64  // 跳过的元素数量
	Count  int64  // 返回的最大元素数量
}

// Set 实现 redis.Client 接口
func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// Get 实现 redis.Client 接口
func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// Del 实现 redis.Client 接口
func (m *mockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys[0]) // 在测试中我们只使用一个key
	return args.Get(0).(int64), args.Error(1)
}

// Exists 实现 redis.Client 接口
func (m *mockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Eval 实现 redis.Client 接口
func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, script, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

// Close 实现 redis.Client 接口
func (m *mockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Pipeline 实现 redis.Client 接口
func (m *mockRedisClient) Pipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

// HGet 实现 redis.Client 接口
func (m *mockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

// HSet 实现 redis.Client 接口
func (m *mockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

// HDel 实现 redis.Client 接口
func (m *mockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

// HExists 实现 redis.Client 接口
func (m *mockRedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	args := m.Called(ctx, key, field)
	return args.Bool(0), args.Error(1)
}

// ZAdd 实现 redis.Client 接口
func (m *mockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZRem 实现 redis.Client 接口
func (m *mockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZScore 实现 redis.Client 接口
func (m *mockRedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	args := m.Called(ctx, key, member)
	return args.Get(0).(float64), args.Error(1)
}

// ZRangeByScore 实现 redis.Client 接口
func (m *mockRedisClient) ZRangeByScore(ctx context.Context, key string, opt *ZRangeBy) ([]string, error) {
	args := m.Called(ctx, key, opt)
	return args.Get(0).([]string), args.Error(1)
}

// LPush 实现 redis.Client 接口
func (m *mockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

// RPush 实现 redis.Client 接口
func (m *mockRedisClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

// LPop 实现 redis.Client 接口
func (m *mockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// RPop 实现 redis.Client 接口
func (m *mockRedisClient) RPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// LLen 实现 redis.Client 接口
func (m *mockRedisClient) LLen(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

// LRange 实现 redis.Client 接口
func (m *mockRedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

// SAdd 实现 redis.Client 接口
func (m *mockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// SRem 实现 redis.Client 接口
func (m *mockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// SMembers 实现 redis.Client 接口
func (m *mockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

// SIsMember 实现 redis.Client 接口
func (m *mockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	args := m.Called(ctx, key, member)
	return args.Bool(0), args.Error(1)
}

// Ping 实现 redis.Client 接口
func (m *mockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Pool 实现 redis.Client 接口
func (m *mockRedisClient) Pool() redis.Pool {
	args := m.Called()
	return args.Get(0).(redis.Pool)
}

// PoolStats 实现 redis.Client 接口
func (m *mockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// TxPipeline 实现 redis.Client 接口
func (m *mockRedisClient) TxPipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

// Publish 实现 redis.Client 接口
func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

// Subscribe 实现 redis.Client 接口
func (m *mockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(redis.PubSub)
}

func TestNewRedisStore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := new(mockRedisClient)
		client.On("Close").Return(nil)
		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(client, opts)
		assert.NoError(t, err)
		assert.NotNil(t, store)
	})

	t.Run("nil client", func(t *testing.T) {
		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(nil, opts)
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}

func TestRedisStore_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := new(mockRedisClient)
		client.On("Set", mock.Anything, "blacklist:token1", "test", time.Hour).Return(nil)
		client.On("Close").Return(nil)

		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(client, opts)
		assert.NoError(t, err)

		err = store.Add(context.Background(), "token1", "test", time.Hour)
		assert.NoError(t, err)

		store.Close()
		client.AssertExpectations(t)
	})
}

func TestRedisStore_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := new(mockRedisClient)
		client.On("Get", mock.Anything, "blacklist:token1").Return("test", nil)
		client.On("Close").Return(nil)

		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(client, opts)
		assert.NoError(t, err)
		defer store.Close()

		reason, err := store.Get(context.Background(), "token1")
		assert.NoError(t, err)
		assert.Equal(t, "test", reason)

		store.Close()
		client.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		client := new(mockRedisClient)
		client.On("Get", mock.Anything, "blacklist:token1").Return("", redis.ErrNil)
		client.On("Close").Return(nil)

		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(client, opts)
		assert.NoError(t, err)

		reason, err := store.Get(context.Background(), "token1")
		assert.Error(t, err)
		assert.Equal(t, "", reason)
		assert.Equal(t, codes.StoreErrNotFound, errors.GetErrorCode(err))

		store.Close()
		client.AssertExpectations(t)
	})

	t.Run("redis error", func(t *testing.T) {
		client := new(mockRedisClient)
		client.On("Get", mock.Anything, "blacklist:token1").Return("", errClosed)
		client.On("Close").Return(nil)

		opts := blacklist.DefaultOptions()
		store, err := blacklist.NewRedisStore(client, opts)
		assert.NoError(t, err)

		reason, err := store.Get(context.Background(), "token1")
		assert.Error(t, err)
		assert.Equal(t, "", reason)
		assert.Equal(t, codes.StoreErrGet, errors.GetErrorCode(err))

		store.Close()
		client.AssertExpectations(t)
	})
}
