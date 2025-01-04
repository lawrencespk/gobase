package stress

import (
	"compress/gzip"
	"context"
	"errors"
	"sync"
	"sync/atomic"
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

const (
	numPublishers  = 5   // 减少发布者数量
	numSubscribers = 3   // 减少订阅者数量
	eventsPerPub   = 100 // 减少每个发布者的消息数量
	testTimeout    = 60  // 增加测试超时时间
)

func TestEventStress(t *testing.T) {
	t.Logf("Starting test with configuration:")
	t.Logf("- Publishers: %d", numPublishers)
	t.Logf("- Subscribers: %d", numSubscribers)
	t.Logf("- Events per publisher: %d", eventsPerPub)
	t.Logf("- Test timeout: %d seconds", testTimeout)

	// 启动 Redis 容器
	redisAddr, err := testutils.StartRedisContainer(context.Background())
	require.NoError(t, err)
	defer testutils.CleanupRedisContainers()

	// 给容器一些时间完全启动
	time.Sleep(2 * time.Second)

	// 初始化 Redis 客户端，使用我们自己的 redis 包
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{redisAddr}),
		redis.WithPassword(""),
		redis.WithDB(0),
		redis.WithMaxRetries(3),
		redis.WithReadTimeout(time.Second*3),
		redis.WithWriteTimeout(time.Second*3),
	)
	require.NoError(t, err)
	defer redisClient.Close()

	// 在启动订阅者之前先检查 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = redisClient.Ping(ctx)
	cancel()
	require.NoError(t, err, "Redis connection check failed")

	// 用于跟踪接收到的消息数量
	var receivedCount int32
	expectedTotal := numPublishers * eventsPerPub

	// 创建一个完成通道
	done := make(chan struct{})

	// 初始化指标收集器，使用我们自己的 metric 包
	pubMetrics := metric.NewCounter(metric.CounterOpts{
		Name: "jwt_events_published_total",
		Help: "Total number of JWT events published",
	}).WithLabels("type", "status")
	require.NoError(t, pubMetrics.Register())

	subMetrics := metric.NewCounter(metric.CounterOpts{
		Name: "jwt_events_received_total",
		Help: "Total number of JWT events received",
	}).WithLabels("type", "status")
	require.NoError(t, subMetrics.Register())

	// 初始化日志
	fileManager := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/stress_test.log",
	})
	defer fileManager.Close()

	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,
		BatchSize:       100,
		FlushInterval:   time.Second,
		Workers:         4,
		RetryCount:      3,
		RetryInterval:   time.Millisecond * 100,
		MaxBatchWait:    time.Second,
		ShutdownTimeout: time.Second * 5,
	}

	loggerOpts := &logrus.Options{
		Level:        types.DebugLevel,
		ReportCaller: true,
		OutputPaths:  []string{"stdout", "logs/stress_test.log"},
		AsyncConfig: logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    8192,
			FlushInterval: time.Second,
		},
		CompressConfig: logrus.CompressConfig{
			Enable:       true,
			Algorithm:    "gzip",
			Level:        gzip.BestCompression,
			DeleteSource: true,
			Interval:     time.Hour,
			LogPaths:     []string{"logs/stress_test.log"},
		},
	}

	t.Logf("Logger options:")
	t.Logf("- AsyncConfig: %+v", loggerOpts.AsyncConfig)
	t.Logf("- CompressConfig: %+v", loggerOpts.CompressConfig)
	t.Logf("- FileManager config: %+v", fileManager)
	t.Logf("- QueueConfig: %+v", queueConfig)

	logger, err := logrus.NewLogger(fileManager, queueConfig, loggerOpts)
	require.NoError(t, err)
	defer logger.Close()

	// 创建发布者
	pub, err := events.NewPublisher(redisClient, logger, events.WithMetrics(pubMetrics))
	require.NoError(t, err)

	// 创建处理器
	handler := func(event *events.Event) error {
		count := atomic.AddInt32(&receivedCount, 1)

		// 添加日志以便调试
		logger.Debug(ctx, "Received event",
			types.Field{Key: "event_id", Value: event.ID},
			types.Field{Key: "event_type", Value: string(event.Type)},
			types.Field{Key: "count", Value: count},
		)

		if count >= int32(expectedTotal) {
			select {
			case <-done:
			default:
				close(done)
			}
		}
		return nil
	}

	// 创建带超时的父上下文
	parentCtx, parentCancel := context.WithTimeout(context.Background(), time.Duration(testTimeout)*time.Second)
	defer parentCancel()

	// 创建发布者上下文
	pubCtx, pubCancel := context.WithCancel(parentCtx)
	defer pubCancel()

	// 创建订阅者上下文
	subCtx, subCancel := context.WithCancel(parentCtx)
	defer subCancel()

	// 创建一个通道来表示订阅者已准备就绪
	subscribersReady := make(chan struct{})
	var subscribersCount int32

	// 启动订阅者
	var wg sync.WaitGroup
	subscribers := make([]*events.Subscriber, numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		sub := events.NewSubscriber(
			redisClient,
			events.WithSubscriberLogger(logger),
			events.WithSubscriberMetrics(subMetrics),
		)
		subscribers[i] = sub
		sub.RegisterHandler(events.TokenRevoked, handler)

		wg.Add(1)
		go func(s *events.Subscriber, index int) {
			defer wg.Done()

			// 订阅并等待连接建立
			if err := s.Subscribe(subCtx); err != nil {
				if !errors.Is(err, context.Canceled) {
					logger.Error(subCtx, "Subscriber error",
						types.Field{Key: "error", Value: err},
						types.Field{Key: "subscriber_index", Value: index},
					)
				}
				return
			}

			// 标记订阅者已准备就绪
			if atomic.AddInt32(&subscribersCount, 1) == int32(numSubscribers) {
				close(subscribersReady)
			}

			// 等待上下文取消
			<-subCtx.Done()
		}(sub, i)
	}

	// 等待所有订阅者准备就绪或超时
	select {
	case <-subscribersReady:
		logger.Info(parentCtx, "All subscribers are ready")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for subscribers to be ready")
	}

	// 启动发布者
	var pubWg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < numPublishers; i++ {
		pubWg.Add(1)
		go func(id int) {
			defer pubWg.Done()
			for j := 0; j < eventsPerPub; j++ {
				select {
				case <-pubCtx.Done():
					return
				default:
					data := map[string]interface{}{
						"publisher_id": id,
						"event_num":    j,
						"timestamp":    time.Now().UnixNano(),
					}

					if err := pub.Publish(pubCtx, events.TokenRevoked, data); err != nil {
						logger.Error(pubCtx, "Failed to publish event",
							types.Field{Key: "error", Value: err},
							types.Field{Key: "publisher_id", Value: id},
							types.Field{Key: "event_num", Value: j},
						)
						continue
					}

					// 添加短暂延迟避免消息堆积
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// 等待发布完成
	pubWg.Wait()

	// 等待所有消息被处理或超时
	select {
	case <-done:
		t.Logf("All messages processed successfully")
	case <-parentCtx.Done():
		t.Errorf("Test timed out. Received %d/%d messages", atomic.LoadInt32(&receivedCount), expectedTotal)
	}

	// 取消订阅者
	subCancel()

	// 等待所有订阅者退出
	wg.Wait()

	// 输出统计信息
	duration := time.Since(startTime)
	messagesPerSecond := float64(atomic.LoadInt32(&receivedCount)) / duration.Seconds()

	t.Logf("Test completed in %v", duration)
	t.Logf("Messages processed: %d/%d", atomic.LoadInt32(&receivedCount), expectedTotal)
	t.Logf("Throughput: %.2f messages/second", messagesPerSecond)
}

// BenchmarkEventPublish 基准测试事件发布性能
func BenchmarkEventPublish(b *testing.B) {
	// 初始化 Redis 客户端，使用 Option 模式
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{"localhost:6379"}),
		redis.WithPassword(""),
		redis.WithDB(0),
	)
	require.NoError(b, err)
	defer redisClient.Close()

	// 初始化日志
	fileManager := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/bench_test.log",
	})
	defer fileManager.Close()

	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,
		BatchSize:       100,
		FlushInterval:   time.Second,
		Workers:         4,
		RetryCount:      3,
		RetryInterval:   time.Millisecond * 100,
		MaxBatchWait:    time.Second,
		ShutdownTimeout: time.Second * 5,
	}

	loggerOpts := &logrus.Options{
		Level:        types.ErrorLevel,
		ReportCaller: true,
		OutputPaths:  []string{"stdout", "logs/bench_test.log"},
		AsyncConfig: logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    8192,
			FlushInterval: time.Second,
		},
		CompressConfig: logrus.CompressConfig{
			Enable:    true,
			Algorithm: "gzip",
		},
	}

	logger, err := logrus.NewLogger(fileManager, queueConfig, loggerOpts)
	require.NoError(b, err)
	defer logger.Close()

	publisher, err := events.NewPublisher(redisClient, logger)
	require.NoError(b, err)

	ctx := context.Background()
	payload := map[string]interface{}{
		"test_id":   "benchmark",
		"timestamp": time.Now().UnixNano(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := publisher.Publish(ctx, events.TokenRevoked, payload)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
