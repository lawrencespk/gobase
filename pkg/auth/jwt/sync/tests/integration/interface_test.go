package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/config"
	"gobase/pkg/config/types"
	"gobase/pkg/logger/logrus"
	logtypes "gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/trace/jaeger"
)

type IntegrationTestSuite struct {
	client  redis.Client
	ctx     context.Context
	metrics *collector.RedisCollector
	tracer  *jaeger.Provider
	logger  logtypes.Logger
}

func setupTestConfig() error {
	// 创建测试配置
	cfg := &types.Config{
		Jaeger: types.JaegerConfig{
			Enable:      true,
			ServiceName: "jwt-test",
			Agent: types.JaegerAgentConfig{
				Host: "localhost",
				Port: "6831",
			},
			Collector: types.JaegerCollectorConfig{
				Endpoint: "http://localhost:14268/api/traces",
				Timeout:  5,
			},
			Sampler: types.JaegerSamplerConfig{
				Type:  "const",
				Param: 1,
			},
			Buffer: types.JaegerBufferConfig{
				Enable:        true,
				Size:          1000,
				FlushInterval: 1,
			},
		},
	}

	// 将 types.Config 转换为内部 Config
	internalCfg := &config.Config{
		Jaeger: cfg.Jaeger,
		// 设置其他必需的配置项以通过验证
		ELK: config.ELKConfig{
			Addresses: []string{"http://localhost:9200"},
			Username:  "elastic",
			Password:  "elastic",
			Index:     "test-index",
			Timeout:   30,
			Bulk: config.BulkConfig{
				BatchSize:  1000,
				FlushBytes: 5 * 1024 * 1024,
				Interval:   "5s",
			},
		},
		Logger: config.LoggerConfig{
			Level:  "debug",
			Output: "console",
		},
	}

	// 设置全局配置
	config.SetConfig(internalCfg)
	return nil
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// 初始化测试配置
	if err := setupTestConfig(); err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}

	// 创建日志记录器
	logger, err := logrus.NewLogger(
		nil, // FileManager
		logrus.QueueConfig{
			MaxSize:         1000,
			BatchSize:       100,
			Workers:         1,
			FlushInterval:   time.Second,
			RetryCount:      3,
			RetryInterval:   time.Second,
			MaxBatchWait:    time.Second * 5,
			ShutdownTimeout: time.Second * 10,
		},
		&logrus.Options{
			Level:       logtypes.DebugLevel,
			TimeFormat:  time.RFC3339,
			OutputPaths: []string{"stdout"},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 创建 metrics collector
	metrics := collector.NewRedisCollector("redis_test")

	// 创建 Redis 客户端
	client, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithLogger(logger),
		redis.WithEnableMetrics(true),
		redis.WithMetricsNamespace("redis_test"),
		redis.WithPoolSize(10),
		redis.WithMinIdleConns(2),
		redis.WithIdleTimeout(time.Minute),
		redis.WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	// 创建 jaeger provider
	provider, err := jaeger.NewProvider()
	if err != nil {
		t.Fatalf("Failed to create jaeger provider: %v", err)
	}

	return &IntegrationTestSuite{
		client:  client,
		ctx:     context.Background(),
		metrics: metrics,
		tracer:  provider,
		logger:  logger,
	}
}

func teardownIntegrationTest(suite *IntegrationTestSuite) {
	if suite.client != nil {
		suite.client.Close()
	}
	if suite.tracer != nil {
		suite.tracer.Close()
	}
	if suite.logger != nil {
		suite.logger.Sync()
	}
}

func TestRedisIntegration(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer teardownIntegrationTest(suite)

	t.Run("Basic Operations", func(t *testing.T) {
		// 测试基本的 Set/Get 操作
		err := suite.client.Set(suite.ctx, "test_key", "test_value", time.Minute)
		if err != nil {
			t.Fatalf("Set operation failed: %v", err)
		}

		val, err := suite.client.Get(suite.ctx, "test_key")
		if err != nil {
			t.Fatalf("Get operation failed: %v", err)
		}

		if val != "test_value" {
			t.Errorf("Expected 'test_value', got '%v'", val)
		}

		// 验证 metrics
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Connection Pool", func(t *testing.T) {
		// 测试并发操作以验证连接池
		const concurrency = 5
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				key := fmt.Sprintf("test_key_%d", id)
				err := suite.client.Set(suite.ctx, key, "value", time.Minute)
				if err != nil {
					t.Errorf("Concurrent Set failed: %v", err)
				}
				done <- true
			}(i)
		}

		// 等待所有操作完成
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}
