package unit

import (
	"context"
	"fmt"
	"sync"
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

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *multilevel.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &multilevel.Config{
				L1Config: &multilevel.L1Config{
					MaxEntries:      1000,
					CleanupInterval: time.Minute,
				},
				L2Config: &multilevel.L2Config{
					RedisAddr:     "localhost:6379",
					RedisPassword: "",
					RedisDB:       0,
				},
				L1TTL:             time.Hour,
				EnableAutoWarmup:  true,
				WarmupInterval:    time.Hour,
				WarmupConcurrency: 10,
			},
			wantErr: false,
		},
		{
			name: "missing L1 config",
			config: &multilevel.Config{
				L2Config: &multilevel.L2Config{
					RedisAddr: "localhost:6379",
				},
				L1TTL: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "missing L2 config",
			config: &multilevel.Config{
				L1Config: &multilevel.L1Config{
					MaxEntries:      1000,
					CleanupInterval: time.Minute,
				},
				L1TTL: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "invalid L1 TTL",
			config: &multilevel.Config{
				L1Config: &multilevel.L1Config{
					MaxEntries:      1000,
					CleanupInterval: time.Minute,
				},
				L2Config: &multilevel.L2Config{
					RedisAddr: "localhost:6379",
				},
				L1TTL: 0, // invalid
			},
			wantErr: true,
		},
		{
			name: "invalid warmup interval",
			config: &multilevel.Config{
				L1Config: &multilevel.L1Config{
					MaxEntries:      1000,
					CleanupInterval: time.Minute,
				},
				L2Config: &multilevel.L2Config{
					RedisAddr: "localhost:6379",
				},
				L1TTL:            time.Hour,
				EnableAutoWarmup: true,
				WarmupInterval:   0, // invalid when EnableAutoWarmup is true
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_ErrorHandling(t *testing.T) {
	mockRedis := mock.NewMockRedisClient()
	mockLogger := mock.NewMockLogger()

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

	manager, err := multilevel.NewManager(config, mockRedis, mockLogger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("invalid cache level", func(t *testing.T) {
		_, err := manager.GetFromLevel(ctx, "key", cache.Level(999))
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.CacheNotFoundError))

		err = manager.SetToLevel(ctx, "key", "value", time.Hour, cache.Level(999))
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.CacheNotFoundError))

		err = manager.DeleteFromLevel(ctx, "key", cache.Level(999))
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.CacheNotFoundError))
	})

	t.Run("L1 and L2 interaction", func(t *testing.T) {
		// 测试L1未命中但L2命中的情况
		err := manager.SetToLevel(ctx, "key", "value", time.Hour, cache.L2Cache)
		require.NoError(t, err)

		value, err := manager.Get(ctx, "key")
		assert.NoError(t, err)
		assert.Equal(t, "value", value)

		// 验证数据是否已经写入L1
		l1Value, err := manager.GetFromLevel(ctx, "key", cache.L1Cache)
		assert.NoError(t, err)
		assert.Equal(t, "value", l1Value)
	})
}

func TestManager_ConcurrentAccess(t *testing.T) {
	mockRedis := mock.NewMockRedisClient()
	mockLogger := mock.NewMockLogger()

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

	manager, err := multilevel.NewManager(config, mockRedis, mockLogger)
	require.NoError(t, err)

	ctx := context.Background()

	// 并发写入测试
	t.Run("concurrent write", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key_%d", i)
				err := manager.Set(ctx, key, fmt.Sprintf("value_%d", i), time.Hour)
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()
	})

	// 并发读取测试
	t.Run("concurrent read", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key_%d", i)
				value, err := manager.Get(ctx, key)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("value_%d", i), value)
			}(i)
		}
		wg.Wait()
	})
}

func TestManager_WarmupEdgeCases(t *testing.T) {
	mockRedis := mock.NewMockRedisClient()
	mockLogger := mock.NewMockLogger()

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

	manager, err := multilevel.NewManager(config, mockRedis, mockLogger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("empty keys", func(t *testing.T) {
		err := manager.Warmup(ctx, []string{})
		assert.NoError(t, err)
	})

	t.Run("non-existent keys", func(t *testing.T) {
		err := manager.Warmup(ctx, []string{"non_existent_key"})
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
	})

	t.Run("mixed existing and non-existing keys", func(t *testing.T) {
		// 先设置一个存在的key
		err := manager.SetToLevel(ctx, "existing_key", "value", time.Hour, cache.L2Cache)
		require.NoError(t, err)

		// 预热混合的keys
		err = manager.Warmup(ctx, []string{"existing_key", "non_existent_key"})
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.RedisKeyNotFoundError))

		// 验证存在的key是否已预热到L1
		value, err := manager.GetFromLevel(ctx, "existing_key", cache.L1Cache)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	})
}
