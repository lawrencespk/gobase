package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/sync"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
)

// MockCache 实现 Cache 接口的 mock
type MockCache struct {
	setnx func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	del   func(ctx context.Context, key string) error
}

func NewMockCache() *MockCache {
	return &MockCache{
		setnx: func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
			return true, nil
		},
		del: func(ctx context.Context, key string) error {
			return nil
		},
	}
}

// 实现 Cache 接口
func (m *MockCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return m.setnx(ctx, key, value, expiration)
}

func (m *MockCache) Del(ctx context.Context, key string) error {
	return m.del(ctx, key)
}

// 其他必需的 Cache 接口方法
func (m *MockCache) Get(ctx context.Context, key string) (interface{}, error) { return nil, nil }
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (m *MockCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) { return nil, nil }
func (m *MockCache) MSet(ctx context.Context, pairs ...interface{}) error            { return nil }
func (m *MockCache) Incr(ctx context.Context, key string) (int64, error)             { return 0, nil }
func (m *MockCache) Exists(ctx context.Context, key string) (bool, error)            { return false, nil }
func (m *MockCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return false, nil
}
func (m *MockCache) Close() error { return nil }

type RedisTestSuite struct {
	locker  sync.Locker
	client  redisClient.Cache
	ctx     context.Context
	mock    *MockCache
	logger  types.Logger
	metrics *collector.BusinessCollector
}

func setupTestSuite(t *testing.T) *RedisTestSuite {
	// 创建mock缓存
	mockCache := NewMockCache()

	// 创建测试logger
	logger, err := logrus.NewLogger(
		nil,
		logrus.QueueConfig{},
		&logrus.Options{
			OutputPaths: []string{"stdout"},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 创建metrics collector
	metrics := collector.NewBusinessCollector("jwt_lock")

	return &RedisTestSuite{
		ctx:     context.Background(),
		client:  mockCache,
		mock:    mockCache,
		logger:  logger,
		metrics: metrics,
		locker:  sync.NewLocker(mockCache, "test-key"),
	}
}

func TestRedisLocker(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.logger.Sync()

	t.Run("Lock Operations", func(t *testing.T) {
		// 测试获取锁
		t.Run("Lock", func(t *testing.T) {
			// 设置mock行为
			suite.mock.setnx = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
				return true, nil
			}

			err := suite.locker.Lock(suite.ctx)
			if err != nil {
				t.Errorf("Lock failed: %v", err)
			}
		})

		// 测试尝试获取锁
		t.Run("TryLock", func(t *testing.T) {
			suite.mock.setnx = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
				return true, nil
			}

			success, err := suite.locker.TryLock(suite.ctx)
			if err != nil {
				t.Errorf("TryLock failed: %v", err)
			}
			if !success {
				t.Error("Expected TryLock to succeed")
			}
		})

		// 测试释放锁
		t.Run("Unlock", func(t *testing.T) {
			suite.mock.del = func(ctx context.Context, key string) error {
				return nil
			}

			err := suite.locker.Unlock(suite.ctx)
			if err != nil {
				t.Errorf("Unlock failed: %v", err)
			}
		})

		// 测试锁已被占用
		t.Run("Lock Already Held", func(t *testing.T) {
			suite.mock.setnx = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
				return false, nil
			}

			err := suite.locker.Lock(suite.ctx)
			if !errors.HasErrorCode(err, codes.RedisLockError) {
				t.Errorf("Expected RedisLockError, got: %v", err)
			}
		})

		// 测试Redis错误
		t.Run("Redis Error", func(t *testing.T) {
			redisErr := errors.NewRedisCommandError("redis error", nil)
			suite.mock.setnx = func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
				return false, redisErr
			}

			_, err := suite.locker.TryLock(suite.ctx)
			if !errors.HasErrorCode(err, codes.RedisLockError) {
				t.Errorf("Expected RedisLockError, got: %v", err)
			}
		})
	})
}
