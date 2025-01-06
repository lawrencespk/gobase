package unit

import (
	"testing"

	"gobase/pkg/auth/jwt/session"

	"github.com/stretchr/testify/assert"
)

func TestSessionInterface(t *testing.T) {
	t.Run("Store Interface", func(t *testing.T) {
		// 测试 Store 接口的方法签名
		// 使用已经在 manager_test.go 中定义的 mockStore
		var _ session.Store = (*mockStore)(nil)
	})

	t.Run("Options", func(t *testing.T) {
		opts := &session.Options{
			Redis: &session.RedisOptions{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
			KeyPrefix:     "test:",
			EnableMetrics: true,
		}
		assert.NotNil(t, opts)
		assert.NotEmpty(t, opts.KeyPrefix)
	})

	t.Run("RedisOptions", func(t *testing.T) {
		redisOpts := &session.RedisOptions{
			Addr:     "localhost:6379",
			Password: "test",
			DB:       1,
		}
		assert.NotNil(t, redisOpts)
		assert.Equal(t, "localhost:6379", redisOpts.Addr)
	})
}
