package unit_test

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/security"
	"gobase/pkg/cache"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/monitor/prometheus/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestPolicy(t *testing.T) {
	ctx := context.Background()
	mockCache := new(mockCache)
	logger := logrus.NewNopLogger()
	metrics := metrics.NewJWTMetrics()

	testKey := "test:key"
	testValue := []byte("test value")
	testExpiration := time.Hour

	mockCache.On("GetLevel").Return(cache.Level(1))
	mockCache.On("Set", mock.Anything, testKey, testValue, testExpiration).Return(nil)
	mockCache.On("Get", mock.Anything, testKey).Return(testValue, nil)
	mockCache.On("Delete", mock.Anything, testKey).Return(nil)
	mockCache.On("Clear", mock.Anything).Return(nil)

	policy := &security.Policy{
		Cache:              mockCache,
		Logger:             logger,
		Metrics:            metrics,
		MaxTokenAge:        time.Hour,
		TokenReuseInterval: time.Minute,
	}

	t.Run("UpdateConfig", func(t *testing.T) {
		maxAge := 2 * time.Hour
		reuseInterval := 5 * time.Minute

		policy.UpdateConfig(maxAge, reuseInterval)

		assert.Equal(t, maxAge, policy.MaxTokenAge)
		assert.Equal(t, reuseInterval, policy.TokenReuseInterval)
	})

	t.Run("Cache Operations", func(t *testing.T) {
		err := policy.Cache.Set(ctx, testKey, testValue, testExpiration)
		assert.NoError(t, err)

		value, err := policy.Cache.Get(ctx, testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, value)

		err = policy.Cache.Delete(ctx, testKey)
		assert.NoError(t, err)

		err = policy.Cache.Clear(ctx)
		assert.NoError(t, err)

		level := policy.Cache.GetLevel()
		assert.Equal(t, cache.Level(1), level)
	})

	mockCache.AssertExpectations(t)
}
