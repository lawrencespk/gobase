package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/auth/jwt/session/tests/testutils"
	"gobase/pkg/client/redis"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

var (
	testStore  session.Store
	testLogger types.Logger
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 初始化测试日志器
	var err error
	testLogger, err = logrus.NewLogger(
		nil,                  // FileManager
		logrus.QueueConfig{}, // 默认队列配置
		&logrus.Options{
			Level:       types.DebugLevel,
			Development: true,
		},
	)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	// 使用已有的Redis容器基础设施
	redisAddr, err := testutils.StartRedisContainer(ctx)
	if err != nil {
		testLogger.WithError(err).Fatal(ctx, "failed to start redis container")
	}

	// 初始化Redis客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress(redisAddr),
		redis.WithLogger(testLogger),
	)
	if err != nil {
		testLogger.WithError(err).Fatal(ctx, "failed to create redis client")
	}

	// 初始化测试配置
	opts := &session.Options{
		Redis: &session.RedisOptions{
			Addr: redisAddr,
		},
		KeyPrefix:     "test:",
		EnableMetrics: true,
		Log:           testLogger,
	}

	// 创建Redis存储
	testStore = session.NewRedisStore(redisClient, opts)

	// 等待Redis就绪
	if err := waitForRedis(ctx); err != nil {
		panic("failed to wait for redis: " + err.Error())
	}

	// 运行测试
	code := m.Run()

	// 清理资源
	if err := testutils.StopRedisContainer(); err != nil {
		panic("failed to stop redis container: " + err.Error())
	}

	os.Exit(code)
}

// waitForRedis 等待Redis就绪
func waitForRedis(ctx context.Context) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return context.DeadlineExceeded
		case <-ticker.C:
			if err := testStore.(*session.RedisStore).Ping(ctx); err == nil {
				return nil
			}
		}
	}
}
