package testutils

import (
	"context"
	redisTestUtils "gobase/pkg/client/redis/tests/testutils"
)

// StartRedisContainer 复用已有的Redis容器启动逻辑
func StartRedisContainer(ctx context.Context) (string, error) {
	return redisTestUtils.StartRedisSingleContainer()
}

// StopRedisContainer 复用已有的Redis容器停止逻辑
func StopRedisContainer() error {
	return redisTestUtils.CleanupRedisContainers()
}
