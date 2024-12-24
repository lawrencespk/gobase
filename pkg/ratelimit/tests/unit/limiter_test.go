package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/cache/redis/client"
	"gobase/pkg/ratelimit/redis"
)

// MockPipeline 模拟Pipeline接口
type MockPipeline struct {
	mock.Mock
}

// Exec 实现 Pipeline 接口的 Exec 方法
func (m *MockPipeline) Exec(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Discard 实现 Pipeline 接口的 Discard 方法
func (m *MockPipeline) Discard() error {
	args := m.Called()
	return args.Error(0)
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

// Incr 实现 Incr 方法
func (m *MockRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

// IncrBy 实现 IncrBy 方法
func (m *MockRedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

// HGet 实现 HGet 方法
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

// HSet 实现 HSet 方法
func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, values...)...)
	return args.Error(0)
}

// LPush 实现 LPush 方法
func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, values...)...)
	return args.Error(0)
}

// LPop 实现 LPop 方法
func (m *MockRedisClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// SAdd 实现 SAdd 方法
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Error(0)
}

// SRem 实现 SRem 方法
func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Error(0)
}

// ZAdd 实现 ZAdd 方法
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Error(0)
}

// ZRem 实现 ZRem 方法
func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(append([]interface{}{ctx, key}, members...)...)
	return args.Error(0)
}

// Eval 实现 Eval 方法
func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := []interface{}{ctx, script, keys}
	callArgs = append(callArgs, args...)
	result := m.Called(callArgs...)
	return result.Get(0), result.Error(1)
}

// Del 实现 Del 方法
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

// Close 实现 Close 方法
func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TxPipeline 实现 TxPipeline 方法
func (m *MockRedisClient) TxPipeline() client.Pipeline {
	args := m.Called()
	if p := args.Get(0); p != nil {
		return p.(client.Pipeline)
	}
	return new(MockPipeline)
}

func TestSlidingWindowLimiter_Allow(t *testing.T) {
	// 创建模拟的Redis客户端
	mockClient := new(MockRedisClient)

	// 创建限流器
	limiter := redis.NewSlidingWindowLimiter(mockClient)

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

	// 创建限流器
	limiter := redis.NewSlidingWindowLimiter(mockClient)

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
				mockClient.On("Del", mock.Anything, []string{"test_key"}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "should handle redis error",
			key:  "test_key",
			mockSetup: func() {
				mockClient.On("Del", mock.Anything, []string{"test_key"}).
					Return(assert.AnError)
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
