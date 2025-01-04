package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/events"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
)

func BenchmarkPublisher(b *testing.B) {
	// 初始化Redis客户端
	redisAddr, err := testutils.StartRedisContainer(context.Background())
	require.NoError(b, err)
	defer testutils.StopRedisContainer()

	client, err := redis.NewClient(
		redis.WithAddress(redisAddr),
	)
	require.NoError(b, err)
	defer client.Close()

	// 使用生产环境配置的日志记录器
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,   // 32KB buffer
		FlushInterval: time.Second, // 1s flush interval
		MaxOpenFiles:  100,         // max open files
		DefaultPath:   "logs/benchmark.log",
	}
	fileManager := logrus.NewFileManager(fileOpts)

	// 创建 logger options
	opts := logrus.DefaultOptions()
	opts.Level = types.ErrorLevel // 直接设置日志级别

	// 创建 logger
	queueConfig := logrus.QueueConfig{
		MaxSize:       1000,
		BatchSize:     100,
		Workers:       1,
		FlushInterval: time.Second,
	}
	logger, err := logrus.NewLogger(fileManager, queueConfig, opts)
	require.NoError(b, err)

	// 创建指标收集器
	metrics := metric.NewCounter(metric.CounterOpts{
		Namespace: "jwt",
		Subsystem: "events",
		Name:      "publisher_events_total",
		Help:      "JWT events benchmark counter",
	})
	metrics.WithLabels("event_type", "status") // 添加 status 标签

	// 创建发布者
	publisher, err := events.NewPublisher(client, logger,
		events.WithMetrics(metrics),
	)
	require.NoError(b, err)

	// 准备测试数据
	ctx := context.Background()
	testEvent := map[string]interface{}{
		"token_id":  "test-token",
		"user_id":   "test-user",
		"device_id": "test-device",
		"ip":        "127.0.0.1",
		"timestamp": time.Now().Unix(),
	}

	b.ResetTimer()

	// 运行基准测试
	b.Run("Publish", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := publisher.Publish(ctx, events.TokenRevoked, testEvent)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSubscriber(b *testing.B) {
	// 初始化Redis客户端
	redisAddr, err := testutils.StartRedisContainer(context.Background())
	require.NoError(b, err)
	defer testutils.StopRedisContainer()

	client, err := redis.NewClient(
		redis.WithAddress(redisAddr),
	)
	require.NoError(b, err)
	defer client.Close()

	// 使用生产环境配置的日志记录器
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,   // 32KB buffer
		FlushInterval: time.Second, // 1s flush interval
		MaxOpenFiles:  100,         // max open files
		DefaultPath:   "logs/benchmark.log",
	}
	fileManager := logrus.NewFileManager(fileOpts)

	// 创建 logger options
	opts := logrus.DefaultOptions()
	opts.Level = types.ErrorLevel // 直接设置日志级别

	// 创建 logger
	queueConfig := logrus.QueueConfig{
		MaxSize:       1000,
		BatchSize:     100,
		Workers:       1,
		FlushInterval: time.Second,
	}
	logger, err := logrus.NewLogger(fileManager, queueConfig, opts)
	require.NoError(b, err)

	// 创建指标收集器
	metrics := metric.NewCounter(metric.CounterOpts{
		Namespace: "jwt",
		Subsystem: "events",
		Name:      "subscriber_events_total",
		Help:      "JWT events benchmark counter",
	})
	metrics.WithLabels("event_type") // 添加标签支持

	// 创建订阅者
	subscriber := events.NewSubscriber(client,
		events.WithSubscriberLogger(logger),
		events.WithSubscriberMetrics(metrics),
	)

	// 注册处理器
	processed := 0
	subscriber.RegisterHandler(events.TokenRevoked, func(event *events.Event) error {
		processed++
		return nil
	})

	b.ResetTimer()

	// 运行基准测试
	b.Run("HandleMessage", func(b *testing.B) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := subscriber.Subscribe(ctx)
		if err != nil && err != context.DeadlineExceeded {
			b.Fatal(err)
		}

		b.ReportMetric(float64(processed), "events_processed")
	})
}
