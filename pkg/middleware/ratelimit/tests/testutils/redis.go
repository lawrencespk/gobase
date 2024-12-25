package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/redis/client"
	"gobase/pkg/cache/redis/config/types"
)

// SetupRedisClient 创建用于测试的Redis客户端
func SetupRedisClient(t *testing.T) client.Client {
	cfg := &types.Config{
		// 基础配置
		Addresses: []string{"localhost:6379"}, // 使用本地Redis
		Database:  0,                          // 使用默认数据库

		// 连接池配置
		PoolSize:     10, // 测试环境使用较小的连接池
		MinIdleConns: 2,
		MaxRetries:   3,

		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// 禁用高级特性
		EnableTLS:     false,
		EnableCluster: false,
		EnableMetrics: false,
		EnableTracing: false,
	}

	// 创建Redis客户端
	redisClient, err := client.NewClient(cfg)
	require.NoError(t, err, "Failed to create Redis client")

	// 测试连接 - 使用Get方法测试一个不存在的键
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Get(ctx, "_test_connection_")
	require.Error(t, err) // 应该返回key不存在的错误,这也说明连接是正常的

	return redisClient
}
