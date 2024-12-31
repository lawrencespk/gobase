package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/ratelimit"
	redis "gobase/pkg/client/redis"
	redistestutils "gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/ratelimit/core"
	redislimiter "gobase/pkg/ratelimit/redis"
)

// SetupRedisClient 创建一个用于测试的Redis客户端
func SetupRedisClient(t *testing.T) redis.Client {
	// 启动 Redis 容器并获取地址
	addr, err := redistestutils.StartRedisSingleContainer()
	require.NoError(t, err, "Failed to start Redis container")

	// 创建Redis客户端
	client, err := redis.NewClient(
		redis.WithAddresses([]string{addr}),
		redis.WithDB(0),
		redis.WithPoolSize(10),
		redis.WithMaxRetries(3),
		redis.WithDialTimeout(5*time.Second),
		redis.WithReadTimeout(3*time.Second),
		redis.WithWriteTimeout(3*time.Second),
	)
	require.NoError(t, err, "Failed to create Redis client")
	require.NotNil(t, client, "Redis client should not be nil")

	return client
}

// NewRedisLimiter 创建一个基于Redis的限流器
func NewRedisLimiter(client redis.Client) core.Limiter {
	// 创建 Redis store
	store := ratelimit.NewStore(client)

	// 创建限流器
	return redislimiter.NewSlidingWindowLimiter(store)
}
