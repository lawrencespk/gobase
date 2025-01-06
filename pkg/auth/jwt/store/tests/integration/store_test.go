package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/store"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func setupTest(t *testing.T) (*store.RedisTokenStore, *store.MemoryStore) {
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 启动 Redis 容器
	redisAddr, err := testutils.StartRedisSingleContainer()
	require.NoError(t, err)
	defer testutils.CleanupRedisContainers()

	// 强制使用 IPv4 地址
	redisAddr = strings.Replace(redisAddr, "localhost", "127.0.0.1", 1)

	// 创建 Redis 客户端，增加更稳定的配置
	redisClient, err := redis.NewClient(
		redis.WithAddress(redisAddr),
		redis.WithPoolSize(10),
		redis.WithMinIdleConns(5),
		redis.WithMaxRetries(5),
		redis.WithDialTimeout(5*time.Second),
		redis.WithReadTimeout(5*time.Second),
		redis.WithWriteTimeout(5*time.Second),
	)
	require.NoError(t, err)

	// 确保 Redis 连接可用
	require.NoError(t, redisClient.Ping(ctx))

	// 创建文件管理器
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/test.log",
	}
	fileManager := logrus.NewFileManager(fileOpts)

	// 创建队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:       1000,
		BatchSize:     100,
		FlushInterval: time.Second,
		RetryInterval: time.Second,
	}

	// 创建logger选项
	logOpts := logrus.DefaultOptions()
	logOpts.Level = types.InfoLevel
	logOpts.Development = true

	// 创建logger
	logger, err := logrus.NewLogger(fileManager, queueConfig, logOpts)
	require.NoError(t, err)

	// 创建Redis store
	redisStore := store.NewRedisTokenStore(redisClient, &store.Options{
		KeyPrefix: "test:",
	}, logger)

	// 创建Memory store
	memStore, err := store.NewMemoryStore(store.Options{
		CleanupInterval: time.Minute,
	})
	require.NoError(t, err)

	return redisStore, memStore
}

func TestTokenStore(t *testing.T) {
	redisStore, memStore := setupTest(t)
	defer func() {
		require.NoError(t, memStore.Close())
		require.NoError(t, redisStore.Close())
	}()

	// 使用带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expiration := time.Hour

	// 创建测试用的 Claims
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test_user"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(expiration)),
	)

	tokenInfo := &jwt.TokenInfo{
		Raw:       "test_token",
		Type:      jwt.AccessToken,
		Claims:    claims,
		ExpiresAt: time.Now().Add(expiration),
		IsRevoked: false,
	}

	// 测试Redis store
	t.Run("Redis Store", func(t *testing.T) {
		// 每个操作使用独立的子上下文
		setCtx, setCancel := context.WithTimeout(ctx, 5*time.Second)
		defer setCancel()
		err := redisStore.Set(setCtx, tokenInfo.Raw, tokenInfo, expiration)
		require.NoError(t, err)

		getCtx, getCancel := context.WithTimeout(ctx, 5*time.Second)
		defer getCancel()
		got, err := redisStore.Get(getCtx, tokenInfo.Raw)
		require.NoError(t, err)
		require.Equal(t, claims.GetUserID(), got.Claims.GetUserID())

		delCtx, delCancel := context.WithTimeout(ctx, 5*time.Second)
		defer delCancel()
		err = redisStore.Delete(delCtx, tokenInfo.Raw)
		require.NoError(t, err)

		checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
		defer checkCancel()
		_, err = redisStore.Get(checkCtx, tokenInfo.Raw)
		require.Error(t, err)
	})

	// 测试Memory store
	t.Run("Memory Store", func(t *testing.T) {
		err := memStore.Set(ctx, tokenInfo.Raw, tokenInfo, time.Hour)
		require.NoError(t, err)

		got, err := memStore.Get(ctx, tokenInfo.Raw)
		require.NoError(t, err)
		require.Equal(t, claims.GetUserID(), got.Claims.GetUserID())

		err = memStore.Delete(ctx, tokenInfo.Raw)
		require.NoError(t, err)

		_, err = memStore.Get(ctx, tokenInfo.Raw)
		require.Error(t, err)
	})
}
