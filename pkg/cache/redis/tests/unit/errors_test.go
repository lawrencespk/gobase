package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/cache/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/stretchr/testify/assert"
)

func TestCacheErrors(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockRedisClient{
		getErr: errors.NewRedisKeyNotFoundError("key not found", nil),
		setErr: errors.NewRedisCommandError("set failed", nil),
		delErr: errors.NewRedisCommandError("del failed", nil),
	}

	cache, err := redis.NewCache(redis.Options{
		Client: mockClient,
	})
	assert.NoError(t, err)

	t.Run("Get error handling", func(t *testing.T) {
		_, err := cache.Get(ctx, "test_key")
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
	})

	t.Run("Set error handling", func(t *testing.T) {
		err := cache.Set(ctx, "test_key", "value", time.Minute)
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisCommandError))
	})

	t.Run("Delete error handling", func(t *testing.T) {
		err := cache.Delete(ctx, "test_key")
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisCommandError))
	})
}
