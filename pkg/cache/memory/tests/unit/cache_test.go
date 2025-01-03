package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cachepkg "gobase/pkg/cache"
	"gobase/pkg/cache/memory"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
)

func TestCache_Basic(t *testing.T) {
	ctx := context.Background()
	log, err := logger.NewLogger()
	require.NoError(t, err)

	cache, err := memory.NewCache(&memory.Config{
		MaxEntries:      1000,
		CleanupInterval: time.Second,
		DefaultTTL:      time.Hour,
	}, log)
	require.NoError(t, err)

	// 使用公开的 Stop 方法
	defer cache.Stop()

	t.Run("Set and Get", func(t *testing.T) {
		// Arrange
		key := "test_key"
		value := "test_value"
		ttl := time.Minute

		// Act
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// Assert
		got, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	t.Run("Delete", func(t *testing.T) {
		// Arrange
		key := "test_delete"
		value := "test_value"
		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		// Act
		err = cache.Delete(ctx, key)
		require.NoError(t, err)

		// Assert
		_, err = cache.Get(ctx, key)
		assert.ErrorIs(t, err, errors.NewCacheNotFoundError("cache miss", nil))
	})

	t.Run("Clear", func(t *testing.T) {
		// Arrange
		keys := []string{"key1", "key2", "key3"}
		for _, key := range keys {
			err := cache.Set(ctx, key, "value", time.Minute)
			require.NoError(t, err)
		}

		// Act
		err := cache.Clear(ctx)
		require.NoError(t, err)

		// Assert
		for _, key := range keys {
			_, err := cache.Get(ctx, key)
			assert.ErrorIs(t, err, errors.NewCacheNotFoundError("cache miss", nil))
		}
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		// Arrange
		key := "test_ttl"
		value := "test_value"
		ttl := time.Millisecond * 100

		// Act
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// Assert
		time.Sleep(ttl * 2)
		_, err = cache.Get(ctx, key)
		assert.ErrorIs(t, err, errors.NewCacheExpiredError("cache expired", nil))
	})

	t.Run("GetLevel", func(t *testing.T) {
		// Act
		level := cache.GetLevel()

		// Assert
		assert.Equal(t, cachepkg.L1Cache, level)
	})
}

func TestCache_InvalidConfig(t *testing.T) {
	log, err := logger.NewLogger()
	require.NoError(t, err)

	tests := []struct {
		name   string
		config *memory.Config
	}{
		{
			name: "negative max entries",
			config: &memory.Config{
				MaxEntries:      -1,
				CleanupInterval: time.Second,
			},
		},
		{
			name: "zero cleanup interval",
			config: &memory.Config{
				MaxEntries:      1000,
				CleanupInterval: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCache, err := memory.NewCache(tt.config, log)
			assert.Error(t, err)
			assert.Nil(t, testCache)
		})
	}
}

func TestCache_Concurrent(t *testing.T) {
	ctx := context.Background()
	log, err := logger.NewLogger()
	require.NoError(t, err)

	cache, err := memory.NewCache(&memory.Config{
		MaxEntries:      1000,
		CleanupInterval: time.Second,
		DefaultTTL:      time.Hour,
	}, log)
	require.NoError(t, err)
	defer cache.Stop()

	t.Run("Concurrent Set and Get", func(t *testing.T) {
		const goroutines = 10
		const operationsPerGoroutine = 100

		done := make(chan bool)
		for i := 0; i < goroutines; i++ {
			go func(id int) {
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j)
					value := fmt.Sprintf("value_%d_%d", id, j)

					err := cache.Set(ctx, key, value, time.Minute)
					assert.NoError(t, err)

					got, err := cache.Get(ctx, key)
					assert.NoError(t, err)
					assert.Equal(t, value, got)
				}
				done <- true
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			<-done
		}
	})
}
