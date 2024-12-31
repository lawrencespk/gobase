package unit

import (
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Run("error codes", func(t *testing.T) {
		assert.True(t, errors.HasErrorCode(redis.ErrInvalidConfig, codes.CacheError))
		assert.True(t, errors.HasErrorCode(redis.ErrConnectionFailed, codes.CacheError))
		assert.True(t, errors.HasErrorCode(redis.ErrOperationFailed, codes.CacheError))
		assert.True(t, errors.HasErrorCode(redis.ErrKeyNotFound, codes.NotFound))
		assert.True(t, errors.HasErrorCode(redis.ErrPoolTimeout, codes.CacheError))
		assert.True(t, errors.HasErrorCode(redis.ErrConnPool, codes.CacheError))
	})

	t.Run("error messages", func(t *testing.T) {
		assert.Contains(t, redis.ErrInvalidConfig.Error(), "invalid redis config")
		assert.Contains(t, redis.ErrConnectionFailed.Error(), "failed to connect")
		assert.Contains(t, redis.ErrOperationFailed.Error(), "operation failed")
		assert.Contains(t, redis.ErrKeyNotFound.Error(), "key not found")
		assert.Contains(t, redis.ErrPoolTimeout.Error(), "pool timeout")
		assert.Contains(t, redis.ErrConnPool.Error(), "connection pool error")
	})
}
