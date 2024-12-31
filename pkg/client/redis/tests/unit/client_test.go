package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// 启用测试模式
	redis.DisableConnectionCheck = true
	defer func() {
		redis.DisableConnectionCheck = false
	}()

	tests := []struct {
		name    string
		opts    []redis.Option
		wantErr bool
		errCode string
	}{
		{
			name: "valid configuration",
			opts: []redis.Option{
				redis.WithAddress("localhost:6379"),
				redis.WithPassword(""),
				redis.WithDB(0),
			},
			wantErr: false,
		},
		{
			name:    "missing address",
			opts:    []redis.Option{},
			wantErr: true,
			errCode: codes.CacheError,
		},
		{
			name: "invalid db number",
			opts: []redis.Option{
				redis.WithAddress("localhost:6379"),
				redis.WithDB(-1),
			},
			wantErr: true,
			errCode: codes.InvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := redis.NewClient(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err, "expected an error but got nil")
				assert.True(t, errors.HasErrorCode(err, tt.errCode),
					"expected error code %s but got %v", tt.errCode, err)
				assert.Nil(t, client, "client should be nil when error occurs")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NoError(t, client.Close())
			}
		})
	}
}

func TestClientOperations(t *testing.T) {
	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	client, err := redis.NewClient(
		redis.WithAddress(addr),
	)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("basic operations", func(t *testing.T) {
		// SET 操作
		err := client.Set(ctx, "test_key", "test_value", time.Minute)
		assert.NoError(t, err)

		// GET 操作
		val, err := client.Get(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)

		// DEL 操作
		n, err := client.Del(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("hash operations", func(t *testing.T) {
		// HSET 操作
		n, err := client.HSet(ctx, "test_hash", "field1", "value1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		// HGET 操作
		val, err := client.HGet(ctx, "test_hash", "field1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)

		// HDEL 操作
		n, err = client.HDel(ctx, "test_hash", "field1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("sorted set operations", func(t *testing.T) {
		// ZADD 操作
		n, err := client.ZAdd(ctx, "test_zset", &redis.Z{
			Score:  1.0,
			Member: "member1",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		// ZREM 操作
		n, err = client.ZRem(ctx, "test_zset", "member1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})
}

func TestNewClientFromConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// 启动 Redis 容器
		addr, err := testutils.StartRedisSingleContainer()
		if err != nil {
			t.Fatal(err)
		}
		defer testutils.CleanupRedisContainers()

		cfg := &redis.Config{
			Addresses:     []string{addr}, // 使用容器地址
			Database:      0,
			EnableMetrics: true,
			EnableTracing: true,
			// 添加重试配置，以确保连接成功
			MaxRetries:   3,
			DialTimeout:  time.Second * 5,
			ReadTimeout:  time.Second * 2,
			WriteTimeout: time.Second * 2,
		}

		client, err := redis.NewClientFromConfig(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// 验证客户端是否真的可用
		err = client.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		client, err := redis.NewClientFromConfig(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClientErrors(t *testing.T) {
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	client, err := redis.NewClient(redis.WithAddress(addr))
	assert.NoError(t, err)
	defer client.Close()

	t.Run("key not found", func(t *testing.T) {
		_, err := client.Get(context.Background(), "non_existent_key")
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.CacheError))
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		err := client.Set(ctx, "test_key", "value", time.Minute)
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.TimeoutError))
	})
}
