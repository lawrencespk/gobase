package unit

import (
	"testing"

	"gobase/pkg/auth/jwt/store"

	"github.com/stretchr/testify/assert"
)

func TestStoreConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &store.Config{
			Type:     store.TypeMemory,
			Host:     "localhost",
			Port:     6379,
			Password: "test",
			DB:       0,
		}
		assert.NotNil(t, cfg)
		assert.Equal(t, store.TypeMemory, cfg.Type)
	})
}

func TestStoreType(t *testing.T) {
	t.Run("supported types", func(t *testing.T) {
		assert.True(t, store.IsValidType(string(store.TypeMemory)))
		assert.True(t, store.IsValidType(string(store.TypeRedis)))
		assert.False(t, store.IsValidType("invalid"))
	})
}
