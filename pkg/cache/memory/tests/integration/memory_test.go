package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/memory"
	"gobase/pkg/errors"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func TestCache_Integration(t *testing.T) {
	// 创建文件管理器
	fm := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/cache_test.log",
	})

	// 创建日志选项
	options := &logrus.Options{
		Level:        types.DebugLevel,
		ReportCaller: true,
		AsyncConfig: logrus.AsyncConfig{
			Enable: true,
		},
		CompressConfig: logrus.CompressConfig{
			Enable: false,
		},
	}

	// 创建日志记录器
	logger, err := logrus.NewLogger(fm, logrus.QueueConfig{}, options)
	require.NoError(t, err)
	defer logger.Close()

	t.Run("Basic Operations", func(t *testing.T) {
		config := memory.DefaultConfig()
		config.MaxEntries = 100
		config.CleanupInterval = time.Second

		cache, err := memory.NewCache(config, logger)
		require.NoError(t, err)
		defer cache.Stop()

		ctx := context.Background()

		// Test Set and Get
		err = cache.Set(ctx, "key1", "value1", time.Minute)
		require.NoError(t, err)

		value, err := cache.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)

		// Test Delete
		err = cache.Delete(ctx, "key1")
		require.NoError(t, err)

		_, err = cache.Get(ctx, "key1")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.NewCacheNotFoundError("cache miss", nil))
	})

	t.Run("Expiration", func(t *testing.T) {
		config := memory.DefaultConfig()
		config.MaxEntries = 100
		config.CleanupInterval = 100 * time.Millisecond

		cache, err := memory.NewCache(config, logger)
		require.NoError(t, err)
		defer cache.Stop()

		ctx := context.Background()

		// Add entry with short expiration
		err = cache.Set(ctx, "key1", "value1", 50*time.Millisecond)
		require.NoError(t, err)

		// Wait for expiration but not cleanup
		time.Sleep(60 * time.Millisecond)

		// Verify entry has expired
		_, err = cache.Get(ctx, "key1")
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.NewCacheExpiredError("cache expired", nil))

		// Wait for cleanup
		time.Sleep(150 * time.Millisecond)

		// After cleanup, should get cache not found error
		_, err = cache.Get(ctx, "key1")
		require.Error(t, err)
		assert.ErrorIs(t, err, errors.NewCacheNotFoundError("cache miss", nil))
	})

	t.Run("MaxEntries", func(t *testing.T) {
		config := memory.DefaultConfig()
		config.MaxEntries = 2
		config.CleanupInterval = time.Second

		cache, err := memory.NewCache(config, logger)
		require.NoError(t, err)
		defer cache.Stop()

		ctx := context.Background()

		// Add entries up to limit
		err = cache.Set(ctx, "key1", "value1", time.Minute)
		require.NoError(t, err)
		err = cache.Set(ctx, "key2", "value2", time.Minute)
		require.NoError(t, err)

		// Try to add one more entry
		err = cache.Set(ctx, "key3", "value3", time.Minute)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.NewCacheCapacityError("memory cache max entries exceeded", nil))
	})

	t.Run("Clear", func(t *testing.T) {
		config := memory.DefaultConfig()
		config.MaxEntries = 100
		config.CleanupInterval = time.Second

		cache, err := memory.NewCache(config, logger)
		require.NoError(t, err)
		defer cache.Stop()

		ctx := context.Background()

		// Add some entries
		err = cache.Set(ctx, "key1", "value1", time.Minute)
		require.NoError(t, err)
		err = cache.Set(ctx, "key2", "value2", time.Minute)
		require.NoError(t, err)

		// Clear cache
		err = cache.Clear(ctx)
		require.NoError(t, err)

		// Verify entries are cleared
		_, err = cache.Get(ctx, "key1")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.NewCacheNotFoundError("cache miss", nil))
	})
}
