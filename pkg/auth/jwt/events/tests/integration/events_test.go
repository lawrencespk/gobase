package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/events"
	"gobase/pkg/client/redis"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/monitor/prometheus/metric"
)

func TestEvents(t *testing.T) {
	ctx := context.Background()

	// 1. Redis 配置
	redisClient, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
		redis.WithDB(0),
	)
	require.NoError(t, err)
	defer redisClient.Close()

	// 2. Logger 配置
	fileManager := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/test.log",
	})

	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,
		BatchSize:       100,
		Workers:         1,
		FlushInterval:   time.Second,
		RetryCount:      3,
		RetryInterval:   time.Second,
		MaxBatchWait:    time.Second * 5,
		ShutdownTimeout: time.Second * 10,
	}

	options := &logrus.Options{
		Development: true,
		Level:       types.InfoLevel,
		OutputPaths: []string{"stdout", "logs/test.log"},
		AsyncConfig: logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    8192,
			FlushInterval: time.Second,
			BlockOnFull:   false,
			DropOnFull:    true,
			FlushOnExit:   true,
		},
	}

	logger, err := logrus.NewLogger(fileManager, queueConfig, options)
	require.NoError(t, err)
	defer logger.Close()

	// 3. Metrics Counter
	t.Log("Creating metrics counter...")
	eventCollector := collector.NewBusinessCollector("jwt_events")
	counter := metric.NewCounter(metric.CounterOpts{
		Namespace: "jwt",
		Subsystem: "events",
		Name:      "test_total",
		Help:      "Counter for JWT events test",
	})
	t.Log("Counter created successfully")

	t.Log("Setting up counter labels...")
	labels := []string{"event_type", "status"}
	counter.WithLabels(labels...)
	t.Logf("Counter labels set: %v", labels)

	err = counter.Register()
	require.NoError(t, err)
	t.Log("Counter registered successfully")

	// 4. 创建 Publisher
	t.Log("Creating publisher...")
	publisher, err := events.NewPublisher(
		redisClient,
		logger,
		events.WithMetrics(counter),
	)
	require.NoError(t, err)
	t.Log("Publisher created successfully")

	// 5. 创建 Subscriber
	t.Log("Creating subscriber...")
	subscriber := events.NewSubscriber(
		redisClient,
		events.WithSubscriberLogger(logger),
		events.WithSubscriberMetrics(counter),
	)
	t.Log("Subscriber created successfully")

	// 注册测试处理器
	t.Log("Registering event handler...")
	subscriber.RegisterHandler(events.TokenRevoked, func(event *events.Event) error {
		t.Log("Event handler called")
		return nil
	})

	// 6. 发布测试事件
	t.Log("Preparing test event data...")
	eventData := map[string]interface{}{
		"user_id":  "test-user",
		"token_id": "test-token",
	}

	t.Log("Publishing event...")
	err = publisher.Publish(ctx, events.TokenRevoked, eventData)
	require.NoError(t, err)
	t.Log("Event published successfully")

	// 记录指标
	t.Log("Recording metrics...")
	t.Log("Observing operation with collector...")
	eventCollector.ObserveOperation("publish_event", 0, nil)

	t.Log("Incrementing counter...")
	counter.WithLabelValues("token_revoked", "success").Inc()
	t.Log("Counter incremented successfully")

	// ... rest of the test ...
}
