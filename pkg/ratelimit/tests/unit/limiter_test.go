package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/cache/redis/ratelimit"
	"gobase/pkg/client/redis"
	redislimiter "gobase/pkg/ratelimit/redis"
)

// MockPipeline 模拟Pipeline接口
type MockPipeline struct {
	mock.Mock
}

// Exec 实现 Pipeline 接口的 Exec 方法
func (m *MockPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	return nil, args.Error(0)
}

// Discard 实现 Pipeline 接口的 Discard 方法
func (m *MockPipeline) Discard() error {
	args := m.Called()
	return args.Error(0)
}

// Close 实现 Pipeline 接口的 Close 方法
func (m *MockPipeline) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Del 实现 Del 方法
func (m *MockPipeline) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// Get 实现 Get 方法
func (m *MockPipeline) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// Set 实现 Set 方法
func (m *MockPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// HSet 实现 HSet 方法
func (m *MockPipeline) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, values...)...)
	return args.Get(0).(int64), args.Error(1)
}

// HGet 实现 HGet 方法
func (m *MockPipeline) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

// HDel 实现 HDel 方法
func (m *MockPipeline) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, fields)...)
	return args.Get(0).(int64), args.Error(1)
}

// SAdd 实现 SAdd 方法
func (m *MockPipeline) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// SRem 实现 SRem 方法
func (m *MockPipeline) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// ZAdd 实现 ZAdd 方法
func (m *MockPipeline) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members)...)
	return args.Get(0).(int64), args.Error(1)
}

// ZRem 实现 ZRem 方法
func (m *MockPipeline) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// MockRedisClient 模拟Redis客户端
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
	args := m.Called(append([]interface{}{ctx, key}, values...)...)
	return args.Get(0).(int64), args.Error(1)
}

// HDel 实现 HDel 方法
func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, fields)...)
	return args.Get(0).(int64), args.Error(1)
}

// LPush 实现 LPush 方法
func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, values...)...)
	return args.Get(0).(int64), args.Error(1)
}

// LPop 实现 LPop 方法
func (m *MockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// SAdd 实现 SAdd 方法
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// SRem 实现 SRem 方法
func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// ZAdd 实现 ZAdd 方法
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members)...)
	return args.Get(0).(int64), args.Error(1)
}

// ZRem 实现 ZRem 方法
func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Get(0).(int64), args.Error(1)
}

// TxPipeline 实现 TxPipeline 方法
func (m *MockRedisClient) TxPipeline() redis.Pipeline {
	args := m.Called()
	if p := args.Get(0); p != nil {
		return p.(redis.Pipeline)
	}
	return new(MockPipeline)
}

// Eval 实现 Eval 方法
func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	mockArgs := m.Called(append([]interface{}{ctx, script, keys}, args...)...)
	return mockArgs.Get(0), mockArgs.Error(1)
}

// Close 实现 Close 方法
func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	if stats := args.Get(0); stats != nil {
		return stats.(*redis.PoolStats)
	}
	return nil
}

func (m *MockRedisClient) Pool() redis.Pool {
	args := m.Called()
	if pool := args.Get(0); pool != nil {
		return pool.(redis.Pool)
	}
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Publish 实现 Publish 方法
func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

// Subscribe 实现 Subscribe 方法
func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) redis.PubSub {
	args := m.Called(append([]interface{}{ctx}, channels)...)
	if ps := args.Get(0); ps != nil {
		return ps.(redis.PubSub)
	}
	return nil
}

func TestSlidingWindowLimiter_Allow(t *testing.T) {
	// 创建模拟的Redis客户端
	mockClient := new(MockRedisClient)

	// 使用 Store 包装 mockClient
	store := ratelimit.NewStore(mockClient)

	// 创建限流器
	limiter := redislimiter.NewSlidingWindowLimiter(store)

	tests := []struct {
		name      string
		key       string
		limit     int64
		window    time.Duration
		mockSetup func()
		want      bool
		wantErr   bool
	}{
		{
			name:   "should allow when under limit",
			key:    "test_key",
			limit:  10,
			window: time.Second,
			mockSetup: func() {
				mockClient.On("Eval",
					mock.Anything,                            // ctx
					mock.AnythingOfType("string"),            // script
					[]string{"test_key", "test_key:counter"}, // keys
					mock.AnythingOfType("int64"),             // now
					mock.AnythingOfType("int64"),             // window in nanoseconds
					int64(10),                                // limit
					int64(1),                                 // n=1 for Allow()
				).Return(interface{}(int64(1)), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "should reject when over limit",
			key:    "test_key",
			limit:  10,
			window: time.Second,
			mockSetup: func() {
				mockClient.On("Eval",
					mock.Anything,                            // ctx
					mock.AnythingOfType("string"),            // script
					[]string{"test_key", "test_key:counter"}, // keys
					mock.AnythingOfType("int64"),             // now
					mock.AnythingOfType("int64"),             // window in nanoseconds
					int64(10),                                // limit
					int64(1),                                 // n=1 for Allow()
				).Return(interface{}(int64(0)), nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "should handle redis error",
			key:    "test_key",
			limit:  10,
			window: time.Second,
			mockSetup: func() {
				mockClient.On("Eval",
					mock.Anything,                            // ctx
					mock.AnythingOfType("string"),            // script
					[]string{"test_key", "test_key:counter"}, // keys
					mock.AnythingOfType("int64"),             // now
					mock.AnythingOfType("int64"),             // window in nanoseconds
					int64(10),                                // limit
					int64(1),                                 // n=1 for Allow()
				).Return(interface{}(nil), assert.AnError)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 mock
			mockClient.ExpectedCalls = nil
			mockClient.Calls = nil

			// 设置mock行为
			tt.mockSetup()

			// 执行测试
			got, err := limiter.Allow(context.Background(), tt.key, tt.limit, tt.window)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// 验证mock调用
			mockClient.AssertExpectations(t)
		})
	}
}

func TestSlidingWindowLimiter_Reset(t *testing.T) {
	// 创建模拟的Redis客户端
	mockClient := new(MockRedisClient)

	// 使用 Store 包装 mockClient
	store := ratelimit.NewStore(mockClient)

	// 创建限流器
	limiter := redislimiter.NewSlidingWindowLimiter(store)

	tests := []struct {
		name      string
		key       string
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "should reset successfully",
			key:  "test_key",
			mockSetup: func() {
				mockClient.On("Del", mock.Anything, []string{"test_key", "test_key:counter"}).
					Return(int64(2), nil)
			},
			wantErr: false,
		},
		{
			name: "should handle redis error",
			key:  "test_key",
			mockSetup: func() {
				mockClient.On("Del", mock.Anything, []string{"test_key", "test_key:counter"}).
					Return(int64(0), assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 mock
			mockClient.ExpectedCalls = nil
			mockClient.Calls = nil

			// 设置mock行为
			tt.mockSetup()

			// 执行测试
			err := limiter.Reset(context.Background(), tt.key)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证mock调用
			mockClient.AssertExpectations(t)
		})
	}
}
