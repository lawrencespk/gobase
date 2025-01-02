package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gobase/pkg/cache"
	"gobase/pkg/cache/multilevel"
	"gobase/pkg/cache/multilevel/tests/testutils"
	"gobase/pkg/client/redis"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

// go test -timeout 5m -v ./pkg/cache/multilevel/tests/integration/...

type MultiLevelCacheTestSuite struct {
	suite.Suite
	redis   redis.Client
	manager *multilevel.Manager
	ctx     context.Context
}

func TestMultiLevelCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 移除环境变量设置和并行执行
	// 直接使用更长的测试超时
	if deadline, ok := t.Deadline(); ok {
		if time.Until(deadline) < 2*time.Minute {
			t.Skip("Skipping test due to insufficient time")
		}
	}

	suite.Run(t, new(MultiLevelCacheTestSuite))
}

func (s *MultiLevelCacheTestSuite) SetupSuite() {
	var err error

	// 启动Redis容器
	redisClient, redisAddr, err := testutils.StartRedisContainer(context.Background())
	require.NoError(s.T(), err)
	s.redis = redisClient

	// 创建文件管理器
	fileManager := logrus.NewFileManager(logrus.FileOptions{
		DefaultPath:   "logs/test.log",
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
	})

	// 配置日志
	logger, err := logrus.NewLogger(
		fileManager,
		logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     100,
			Workers:       1,
			FlushInterval: time.Second,
		},
		&logrus.Options{
			Level:        types.InfoLevel,
			Development:  true,
			ReportCaller: true,
			TimeFormat:   time.RFC3339,
			OutputPaths:  []string{"stdout", "logs/test.log"},
		},
	)
	require.NoError(s.T(), err)

	// 创建多级缓存管理器
	config := &multilevel.Config{
		L1Config: &multilevel.L1Config{
			MaxEntries:      1000,
			CleanupInterval: time.Minute,
		},
		L2Config: &multilevel.L2Config{
			RedisAddr:     redisAddr,
			RedisPassword: "",
			RedisDB:       0,
		},
		L1TTL:             time.Hour,
		EnableAutoWarmup:  true,
		WarmupInterval:    time.Minute,
		WarmupConcurrency: 10,
	}

	s.manager, err = multilevel.NewManager(config, s.redis, logger)
	require.NoError(s.T(), err)

	s.ctx = context.Background()
}

func (s *MultiLevelCacheTestSuite) TearDownSuite() {
	if s.redis != nil {
		testutils.StopRedisContainer(s.redis)
	}
}

func (s *MultiLevelCacheTestSuite) TestBasicOperations() {
	t := s.T()

	// 测试Set和Get
	err := s.manager.Set(s.ctx, "test_key", "test_value", time.Hour)
	require.NoError(t, err)

	value, err := s.manager.Get(s.ctx, "test_key")
	require.NoError(t, err)
	require.Equal(t, "test_value", value)

	// 测试Delete
	err = s.manager.Delete(s.ctx, "test_key")
	require.NoError(t, err)

	_, err = s.manager.Get(s.ctx, "test_key")
	require.Error(t, err)
}

func (s *MultiLevelCacheTestSuite) TestCacheLevels() {
	t := s.T()

	// 测试L1缓存
	err := s.manager.SetToLevel(s.ctx, "l1_key", "l1_value", time.Hour, cache.L1Cache)
	require.NoError(t, err)

	value, err := s.manager.GetFromLevel(s.ctx, "l1_key", cache.L1Cache)
	require.NoError(t, err)
	require.Equal(t, "l1_value", value)

	// 测试L2缓存
	err = s.manager.SetToLevel(s.ctx, "l2_key", "l2_value", time.Hour, cache.L2Cache)
	require.NoError(t, err)

	value, err = s.manager.GetFromLevel(s.ctx, "l2_key", cache.L2Cache)
	require.NoError(t, err)
	require.Equal(t, "l2_value", value)
}

func (s *MultiLevelCacheTestSuite) TestCacheWarmup() {
	t := s.T()

	// 在L2中设置一些数据
	keys := []string{"warmup_1", "warmup_2", "warmup_3"}
	for _, key := range keys {
		err := s.manager.SetToLevel(s.ctx, key, "value_"+key, time.Hour, cache.L2Cache)
		require.NoError(t, err)
	}

	// 执行预热
	err := s.manager.Warmup(s.ctx, keys)
	require.NoError(t, err)

	// 验证数据是否已经预热到L1
	for _, key := range keys {
		value, err := s.manager.GetFromLevel(s.ctx, key, cache.L1Cache)
		require.NoError(t, err)
		require.Equal(t, "value_"+key, value)
	}
}

func (s *MultiLevelCacheTestSuite) TestConcurrentAccess() {
	t := s.T()

	const goroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 写入数据
				err := s.manager.Set(s.ctx, key, value, time.Hour)
				require.NoError(t, err)

				// 读取数据
				got, err := s.manager.Get(s.ctx, key)
				require.NoError(t, err)
				require.Equal(t, value, got)
			}
		}(i)
	}

	wg.Wait()
}
