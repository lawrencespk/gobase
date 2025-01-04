package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/auth/jwt/events"
	eventtypes "gobase/pkg/auth/jwt/events/types"
	"gobase/pkg/cache"
	"gobase/pkg/client/redis"
	loggerTypes "gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
)

// mockRedisClient 实现
type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, script, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

func (m *mockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *mockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *mockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisClient) TxPipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

func (m *mockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

func (m *mockRedisClient) Pool() redis.Pool {
	args := m.Called()
	return args.Get(0).(redis.Pool)
}

func (m *mockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(redis.PubSub)
}

// mockLogger 实现
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...loggerTypes.Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...loggerTypes.Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...loggerTypes.Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Error(ctx context.Context, msg string, fields ...loggerTypes.Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...loggerTypes.Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	m.Called(ctx, format, args)
}

func (m *mockLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	m.Called(ctx, format, args)
}

func (m *mockLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	m.Called(ctx, format, args)
}

func (m *mockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	m.Called(ctx, format, args)
}

func (m *mockLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	m.Called(ctx, format, args)
}

func (m *mockLogger) WithContext(ctx context.Context) loggerTypes.Logger {
	args := m.Called(ctx)
	return args.Get(0).(loggerTypes.Logger)
}

func (m *mockLogger) WithFields(fields ...loggerTypes.Field) loggerTypes.Logger {
	args := m.Called(fields)
	return args.Get(0).(loggerTypes.Logger)
}

func (m *mockLogger) WithError(err error) loggerTypes.Logger {
	args := m.Called(err)
	return args.Get(0).(loggerTypes.Logger)
}

func (m *mockLogger) WithTime(t time.Time) loggerTypes.Logger {
	args := m.Called(t)
	return args.Get(0).(loggerTypes.Logger)
}

func (m *mockLogger) WithCaller(skip int) loggerTypes.Logger {
	args := m.Called(skip)
	return args.Get(0).(loggerTypes.Logger)
}

func (m *mockLogger) SetLevel(level loggerTypes.Level) {
	m.Called(level)
}

func (m *mockLogger) GetLevel() loggerTypes.Level {
	args := m.Called()
	return args.Get(0).(loggerTypes.Level)
}

func (m *mockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// mockCache 实现
type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockCache) GetLevel() cache.Level {
	args := m.Called()
	return args.Get(0).(cache.Level)
}

// 测试用例
func TestPublisher(t *testing.T) {
	t.Run("Publish success", func(t *testing.T) {
		// 准备
		client := new(mockRedisClient)
		logger := new(mockLogger)
		cacheImpl := new(mockCache)
		metrics := metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Subsystem: "events",
			Name:      "publisher_events_total",
			Help:      "Total number of publisher events",
		})
		metrics.WithLabels("event_type", "status")

		t.Log("Creating new publisher...")
		pub, err := events.NewPublisher(client, logger,
			events.WithCache(cacheImpl),
			events.WithMetrics(metrics),
		)
		assert.NoError(t, err)

		// 设置预期
		ctx := context.Background()
		eventType := eventtypes.EventTypeTokenRevoked
		payload := map[string]interface{}{
			"token":   "test_token",
			"user_id": "test_user",
		}

		// 设置 Redis 预期
		t.Log("Setting up Redis expectations...")
		client.On("Publish", mock.Anything, "jwt:events", mock.Anything).Run(func(args mock.Arguments) {
			t.Logf("Redis Publish called with channel: %s", args.Get(1))
			payload := args.Get(2).(string)
			t.Logf("Redis Publish payload: %s", payload)
		}).Return(nil)

		// 设置 Logger 预期
		t.Log("Setting up Logger expectations...")
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			t.Logf("Logger Info called with message: %v", args.Get(1))
			if fields, ok := args.Get(2).([]loggerTypes.Field); ok {
				for _, field := range fields {
					t.Logf("Logger Info field: %s = %v", field.Key, field.Value)
				}
			}
		}).Return()
		logger.On("WithCaller", mock.Anything).Run(func(args mock.Arguments) {
			t.Logf("Logger WithCaller called with skip: %v", args.Get(0))
		}).Return(logger)
		logger.On("WithContext", mock.Anything).Run(func(args mock.Arguments) {
			t.Log("Logger WithContext called")
		}).Return(logger)

		// 设置 Cache 预期
		t.Log("Setting up Cache expectations...")
		cacheImpl.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			t.Logf("Cache Set called with key: %s", args.Get(1))
			t.Logf("Cache Set expiration: %v", args.Get(3))
		}).Return(nil)
		cacheImpl.On("GetLevel").Run(func(args mock.Arguments) {
			t.Log("Cache GetLevel called")
		}).Return(cache.Level(2))

		// 执行
		t.Log("Executing Publish...")
		err = pub.Publish(ctx, events.EventType(eventType), payload)

		// 验证
		t.Log("Verifying results...")
		assert.NoError(t, err)

		t.Log("Verifying Redis expectations...")
		client.AssertExpectations(t)

		t.Log("Verifying Logger expectations...")
		logger.AssertExpectations(t)

		t.Log("Verifying Cache expectations...")
		cacheImpl.AssertExpectations(t)

		t.Log("Test completed")
	})

	t.Run("Publish failure", func(t *testing.T) {
		// 准备
		client := new(mockRedisClient)
		logger := new(mockLogger)
		cacheImpl := new(mockCache)
		metrics := metric.NewCounter(metric.CounterOpts{
			Namespace: "jwt",
			Subsystem: "events",
			Name:      "publisher_events_total",
			Help:      "Total number of publisher events",
		})
		metrics.WithLabels("event_type", "status")

		pub, err := events.NewPublisher(client, logger,
			events.WithCache(cacheImpl),
			events.WithMetrics(metrics),
		)
		assert.NoError(t, err)

		// 设置预期
		ctx := context.Background()
		eventType := eventtypes.EventTypeTokenRevoked
		payload := map[string]interface{}{
			"token":   "test_token",
			"user_id": "test_user",
		}

		// 设置 Redis 预期失败
		client.On("Publish", mock.Anything, "jwt:events", mock.Anything).Return(errors.New("redis error"))

		// 设置 Logger 预期
		logger.On("WithCaller", mock.Anything).Return(logger)
		logger.On("WithContext", mock.Anything).Return(logger)
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

		// 设置 Cache 预期
		cacheImpl.On("GetLevel").Return(cache.Level(2))
		cacheImpl.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// 执行
		err = pub.Publish(ctx, events.EventType(eventType), payload)

		// 验证
		assert.Error(t, err)
		client.AssertExpectations(t)
		logger.AssertExpectations(t)
		cacheImpl.AssertExpectations(t)
	})
}
