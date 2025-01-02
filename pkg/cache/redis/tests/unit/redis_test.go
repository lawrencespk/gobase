package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"

	"github.com/stretchr/testify/assert"
)

func TestRedisClient(t *testing.T) {
	ctx := context.Background()

	t.Run("basic operations", func(t *testing.T) {
		client := newMockRedisClient()

		// Test Set
		err := client.Set(ctx, "test_key", "test_value", time.Minute)
		assert.NoError(t, err)

		// Test Get
		value, err := client.Get(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)

		// Test Get non-existent key
		_, err = client.Get(ctx, "non_existent_key")
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))

		// Test Del
		count, err := client.Del(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Test Exists
		exists, err := client.Exists(ctx, "test_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("error cases", func(t *testing.T) {
		client := newMockRedisClient()
		client.getErr = errors.NewRedisCommandError("get error", nil)
		client.setErr = errors.NewRedisCommandError("set error", nil)
		client.delErr = errors.NewRedisCommandError("del error", nil)

		// Test Get error
		_, err := client.Get(ctx, "test_key")
		assert.Error(t, err)

		// Test Set error
		err = client.Set(ctx, "test_key", "test_value", time.Minute)
		assert.Error(t, err)

		// Test Del error
		_, err = client.Del(ctx, "test_key")
		assert.Error(t, err)
	})
}

func TestLogger(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	t.Run("basic logging", func(t *testing.T) {
		// Test all logging levels
		logger.Debug(ctx, "debug message")
		logger.Info(ctx, "info message")
		logger.Warn(ctx, "warn message")
		logger.Error(ctx, "error message")

		// Test with fields
		logger.WithFields(types.Field{Key: "test", Value: "value"}).Info(ctx, "info with fields")

		// Test with context
		logger.WithContext(ctx).Info(ctx, "info with context")

		// Test with error
		testErr := errors.NewSystemError("test error", nil)
		logger.WithError(testErr).Error(ctx, "error with error")
	})
}
