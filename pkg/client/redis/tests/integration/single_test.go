package integration

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"

	"gobase/pkg/client/redis/tests/testutils"
)

func TestRedisSingleNodeIntegration(t *testing.T) {
	suite.Run(t, new(RedisSingleNodeTestSuite))
}

type RedisSingleNodeTestSuite struct {
	suite.Suite
	client *redis.Client
}

func (s *RedisSingleNodeTestSuite) SetupSuite() {
	// 启动单节点 Redis 容器
	addr, err := testutils.StartRedisSingleContainer()
	s.Require().NoError(err)

	// 创建客户端
	s.client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func (s *RedisSingleNodeTestSuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	testutils.CleanupRedisContainers()
}

func (s *RedisSingleNodeTestSuite) TestBasicOperations() {
	// 基本操作测试
	ctx := context.Background()

	// Set
	err := s.client.Set(ctx, "key", "value", 0).Err()
	s.Require().NoError(err)

	// Get
	val, err := s.client.Get(ctx, "key").Result()
	s.Require().NoError(err)
	s.Equal("value", val)
}

// 其他单节点相关测试...
