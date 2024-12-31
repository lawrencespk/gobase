package integration

import (
	"context"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedisIntegrationTestSuite struct {
	suite.Suite
	client redis.Client
	ctx    context.Context
}

func TestRedisIntegration(t *testing.T) {
	suite.Run(t, new(RedisIntegrationTestSuite))
}

func (s *RedisIntegrationTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	// 启动Redis容器
	addr, err := testutils.StartRedisContainer(s.ctx)
	s.Require().NoError(err, "Failed to start Redis container")

	// 创建Redis客户端
	s.client, err = redis.NewClient(
		redis.WithAddress(addr),
		redis.WithMaxRetries(3),
		redis.WithRetryBackoff(time.Millisecond*100),
		redis.WithEnableMetrics(true),
		redis.WithEnableTracing(true),
	)
	s.Require().NoError(err, "Failed to create Redis client")
}

func (s *RedisIntegrationTestSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	s.Require().NoError(testutils.StopRedisContainer(), "Failed to stop Redis container")
}

func (s *RedisIntegrationTestSuite) TestBasicOperations() {
	// 测试基本的 SET/GET 操作
	s.Run("SET and GET", func() {
		err := s.client.Set(s.ctx, "test_key", "test_value", time.Minute)
		s.Require().NoError(err)

		val, err := s.client.Get(s.ctx, "test_key")
		s.Require().NoError(err)
		s.Equal("test_value", val)
	})

	// 测试 DEL 操作
	s.Run("DEL", func() {
		n, err := s.client.Del(s.ctx, "test_key")
		s.Require().NoError(err)
		s.Equal(int64(1), n)

		_, err = s.client.Get(s.ctx, "test_key")
		s.Error(err) // 应该返回key不存在错误
	})
}

func (s *RedisIntegrationTestSuite) TestHashOperations() {
	// 测试Hash操作
	s.Run("HSET and HGET", func() {
		n, err := s.client.HSet(s.ctx, "test_hash", "field1", "value1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)

		val, err := s.client.HGet(s.ctx, "test_hash", "field1")
		s.Require().NoError(err)
		s.Equal("value1", val)
	})

	s.Run("HDEL", func() {
		n, err := s.client.HDel(s.ctx, "test_hash", "field1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)
	})
}

func (s *RedisIntegrationTestSuite) TestListOperations() {
	// 测试List操作
	s.Run("LPUSH and LPOP", func() {
		n, err := s.client.LPush(s.ctx, "test_list", "value1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)

		val, err := s.client.LPop(s.ctx, "test_list")
		s.Require().NoError(err)
		s.Equal("value1", val)
	})
}

func (s *RedisIntegrationTestSuite) TestSetOperations() {
	// 测试Set操作
	s.Run("SADD and SREM", func() {
		n, err := s.client.SAdd(s.ctx, "test_set", "member1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)

		n, err = s.client.SRem(s.ctx, "test_set", "member1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)
	})
}

func (s *RedisIntegrationTestSuite) TestSortedSetOperations() {
	// 测试Sorted Set操作
	s.Run("ZADD and ZREM", func() {
		n, err := s.client.ZAdd(s.ctx, "test_zset", &redis.Z{
			Score:  1.0,
			Member: "member1",
		})
		s.Require().NoError(err)
		s.Equal(int64(1), n)

		n, err = s.client.ZRem(s.ctx, "test_zset", "member1")
		s.Require().NoError(err)
		s.Equal(int64(1), n)
	})
}

func (s *RedisIntegrationTestSuite) TestPipeline() {
	// 测试Pipeline操作
	s.Run("Pipeline operations", func() {
		pipe := s.client.TxPipeline()
		s.Require().NotNil(pipe)

		err := pipe.Set(s.ctx, "pipe_key1", "value1", time.Minute)
		s.Require().NoError(err)
		err = pipe.Set(s.ctx, "pipe_key2", "value2", time.Minute)
		s.Require().NoError(err)

		cmds, err := pipe.Exec(s.ctx)
		s.Require().NoError(err)
		s.Len(cmds, 2)

		// 验证Pipeline结果
		val1, err := s.client.Get(s.ctx, "pipe_key1")
		s.Require().NoError(err)
		s.Equal("value1", val1)

		val2, err := s.client.Get(s.ctx, "pipe_key2")
		s.Require().NoError(err)
		s.Equal("value2", val2)
	})
}

func (s *RedisIntegrationTestSuite) TestPoolManagement() {
	// 测试连接池管理
	s.Run("Pool stats", func() {
		stats := s.client.Pool().Stats()
		s.Require().NotNil(stats)
		s.GreaterOrEqual(stats.TotalConns, uint32(1))
	})
}

func (s *RedisIntegrationTestSuite) TestErrorHandling() {
	// 测试错误处理
	s.Run("Invalid operation", func() {
		_, err := s.client.Get(s.ctx, "non_existent_key")
		s.Error(err)
	})
}

func (s *RedisIntegrationTestSuite) TestMetrics() {
	// 测试指标收集
	s.Run("Metrics collection", func() {
		err := s.client.Set(s.ctx, "metrics_test", "value", time.Minute)
		s.Require().NoError(err)

		// 这里可以添加对Prometheus指标的验证
		// 需要从registry中获取指标并验证
	})
}

func (s *RedisIntegrationTestSuite) TestTracing() {
	// 测试分布式追踪
	s.Run("Trace generation", func() {
		err := s.client.Set(s.ctx, "trace_test", "value", time.Minute)
		s.Require().NoError(err)

		// 这里可以添加对OpenTracing span的验证
		// 需要从tracer中获取span并验证
	})
}
