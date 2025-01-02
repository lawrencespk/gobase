package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache"
	"gobase/pkg/cache/multilevel"
	"gobase/pkg/cache/multilevel/tests/mock"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

func TestManager_Operations(t *testing.T) {
	// 使用已有的 mock 实现
	mockRedis := mock.NewMockRedisClient()
	mockLogger := mock.NewMockLogger()

	// 创建配置
	config := &multilevel.Config{
		L1Config: &multilevel.L1Config{
			MaxEntries:      1000,
			CleanupInterval: time.Minute,
		},
		L2Config: &multilevel.L2Config{
			RedisAddr:     "localhost:6379",
			RedisPassword: "",
			RedisDB:       0,
		},
		L1TTL:            time.Hour,
		EnableAutoWarmup: true,
		WarmupInterval:   time.Hour,
	}

	// 创建缓存管理器
	manager, err := multilevel.NewManager(config, mockRedis, mockLogger)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// 测试 Get - 缓存未命中
	ctx := context.Background()
	value, err := manager.Get(ctx, "test_key")
	assert.Error(t, err)
	assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
	assert.Nil(t, value)

	// 测试 Set
	err = manager.Set(ctx, "test_key", "test_value", time.Hour)
	assert.NoError(t, err)

	// 测试 Get - 缓存命中
	value, err = manager.Get(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// 测试 Delete
	err = manager.Delete(ctx, "test_key")
	assert.NoError(t, err)

	// 验证删除后无法获取
	value, err = manager.Get(ctx, "test_key")
	assert.Error(t, err)
	assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
	assert.Nil(t, value)
}

func TestManager_Warmup(t *testing.T) {
	// 使用已有的 mock 实现
	mockRedis := mock.NewMockRedisClient()
	mockLogger := mock.NewMockLogger()

	// 创建配置
	config := &multilevel.Config{
		L1Config: &multilevel.L1Config{
			MaxEntries:      1000,
			CleanupInterval: time.Minute,
		},
		L2Config: &multilevel.L2Config{
			RedisAddr:     "localhost:6379",
			RedisPassword: "",
			RedisDB:       0,
		},
		L1TTL:            time.Hour,
		EnableAutoWarmup: true,
		WarmupInterval:   time.Hour,
	}

	// 创建缓存管理器
	manager, err := multilevel.NewManager(config, mockRedis, mockLogger)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// 测试预热
	ctx := context.Background()
	keys := []string{"key1", "key2", "key3"}

	// 先设置一些数据到 L2
	for _, key := range keys {
		err := manager.SetToLevel(ctx, key, "value_"+key, time.Hour, cache.L2Cache)
		require.NoError(t, err)
	}

	// 执行预热
	err = manager.Warmup(ctx, keys)
	assert.NoError(t, err)

	// 验证数据已预热到 L1
	for _, key := range keys {
		value, err := manager.GetFromLevel(ctx, key, cache.L1Cache)
		assert.NoError(t, err)
		assert.Equal(t, "value_"+key, value)
	}
}
