package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/auth/jwt/events"
	"gobase/pkg/client/redis"
)

// mockRedisSubscriber 实现
type mockRedisSubscriber struct {
	mock.Mock
}

// 基础操作
func (m *mockRedisSubscriber) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockRedisSubscriber) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockRedisSubscriber) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// Hash操作
func (m *mockRedisSubscriber) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *mockRedisSubscriber) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisSubscriber) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

// List操作
func (m *mockRedisSubscriber) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisSubscriber) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// Set操作
func (m *mockRedisSubscriber) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisSubscriber) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// ZSet操作
func (m *mockRedisSubscriber) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisSubscriber) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

// 事务操作
func (m *mockRedisSubscriber) TxPipeline() redis.Pipeline {
	args := m.Called()
	return args.Get(0).(redis.Pipeline)
}

// Lua脚本
func (m *mockRedisSubscriber) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, script, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

// 连接管理
func (m *mockRedisSubscriber) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRedisSubscriber) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// 监控相关
func (m *mockRedisSubscriber) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// 缓存管理
func (m *mockRedisSubscriber) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// 连接池管理
func (m *mockRedisSubscriber) Pool() redis.Pool {
	args := m.Called()
	return args.Get(0).(redis.Pool)
}

// Publish/Subscribe 操作
func (m *mockRedisSubscriber) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

func (m *mockRedisSubscriber) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(redis.PubSub)
}

// 首先添加 mock PubSub 实现
type mockPubSub struct {
	mock.Mock
}

func (m *mockPubSub) ReceiveMessage(ctx context.Context) (*redis.Message, error) {
	args := m.Called(ctx)
	if msg := args.Get(0); msg != nil {
		return msg.(*redis.Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPubSub) Close() error {
	return m.Called().Error(0)
}

// 测试用例
func TestSubscriber(t *testing.T) {
	t.Run("Subscribe success", func(t *testing.T) {
		// 准备
		client := new(mockRedisSubscriber)
		logger := new(mockLogger)
		mockPubSub := new(mockPubSub)

		// 添加调试计数器
		receiveCount := 0
		messageProcessed := make(chan struct{})

		// 设置 Subscribe 期望
		client.On("Subscribe", mock.Anything, mock.MatchedBy(func(channels []string) bool {
			t.Logf("Subscribe called with channels: %v", channels)
			return len(channels) == 1 && channels[0] == "jwt:events"
		})).Return(mockPubSub)

		// 设置 PubSub 期望 - 只期望一次调用
		mockPubSub.On("ReceiveMessage", mock.Anything).Run(func(args mock.Arguments) {
			receiveCount++
			ctx := args.Get(0).(context.Context)
			t.Logf("ReceiveMessage called %d times, context: %v, done: %v",
				receiveCount, ctx, ctx.Done())
			close(messageProcessed)
		}).Return(&redis.Message{
			Channel: "jwt:events",
			Payload: `{"id":"test-event","type":"token_revoked"}`,
		}, nil).Once()

		mockPubSub.On("Close").Run(func(args mock.Arguments) {
			t.Log("Close called")
		}).Return(nil)

		// 设置 Logger 期望
		logger.On("Warn", mock.Anything, "no handler registered for event type", mock.Anything).Run(func(args mock.Arguments) {
			t.Logf("Warn called with message: %v", args.Get(1))
		}).Return()

		sub := events.NewSubscriber(client, events.WithSubscriberLogger(logger))

		// 执行
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			t.Log("Starting Subscribe")
			err := sub.Subscribe(ctx)
			t.Logf("Subscribe returned with error: %v", err)
			assert.NoError(t, err)
			close(done)
		}()

		// 等待消息处理完成
		t.Log("Waiting for message to be processed")
		<-messageProcessed

		t.Log("Canceling context")
		cancel()

		// 等待订阅完成
		t.Log("Waiting for subscription to complete")
		<-done
		t.Log("Subscription completed")

		// 验证
		t.Logf("Final ReceiveMessage call count: %d", receiveCount)
		client.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}
