package unit

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	// 启动 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupRedisContainers()

	var client redis.Client
	client, err = redis.NewClient(
		redis.WithAddress(addr),
	)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	t.Run("string interface", func(t *testing.T) {
		err := client.Set(ctx, "test_str", "value", time.Minute)
		assert.NoError(t, err)

		val, err := client.Get(ctx, "test_str")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("hash interface", func(t *testing.T) {
		// 先清理可能存在的旧数据
		_, _ = client.HDel(ctx, "test_hash", "field")

		// 测试 HSet
		n, err := client.HSet(ctx, "test_hash", map[string]interface{}{
			"field": "value",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		// 测试 HGet
		val, err := client.HGet(ctx, "test_hash", "field")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)

		// 清理测试数据
		_, _ = client.HDel(ctx, "test_hash", "field")
	})

	t.Run("list interface", func(t *testing.T) {
		// 测试列表接口方法
		n, err := client.LPush(ctx, "test_list", "value")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		val, err := client.LPop(ctx, "test_list")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("set interface", func(t *testing.T) {
		// 测试集合接口方法
		n, err := client.SAdd(ctx, "test_set", "member")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		n, err = client.SRem(ctx, "test_set", "member")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("sorted set interface", func(t *testing.T) {
		// 测试有序集合接口方法
		n, err := client.ZAdd(ctx, "test_zset", &redis.Z{
			Score:  1.0,
			Member: "member",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)

		n, err = client.ZRem(ctx, "test_zset", "member")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})
}
