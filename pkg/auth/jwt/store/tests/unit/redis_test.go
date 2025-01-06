package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/store"
	"gobase/pkg/client/redis"
	redismock "gobase/pkg/client/redis/tests/mock"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
	"gobase/pkg/logger/logrus"

	"github.com/stretchr/testify/assert"
)

type redisStoreTestSuite struct {
	store  *store.RedisTokenStore
	client redis.Client
	ctx    context.Context
}

func setupRedisTest() (*redisStoreTestSuite, error) {
	// 创建Redis客户端
	client := redismock.NewMockClient()

	// 创建Store实例
	store := store.NewRedisTokenStore(client, &store.Options{
		KeyPrefix: "test:",
	}, logrus.NewNopLogger())

	return &redisStoreTestSuite{
		store:  store,
		client: client,
		ctx:    context.Background(),
	}, nil
}

func TestRedisStore_Set(t *testing.T) {
	suite, err := setupRedisTest()
	assert.NoError(t, err)

	t.Run("successful set", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}

		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.NoError(t, err)
	})

	t.Run("redis error", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}

		// 模拟Redis错误
		mockErr := errors.NewRedisCommandError("mock error", nil)
		suite.client.(*redismock.MockClient).SetError(mockErr)

		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.Error(t, err)
		if e, ok := err.(types.Error); ok {
			assert.Equal(t, codes.RedisCommandError, e.Code())
		}
	})
}

func TestRedisStore_Get(t *testing.T) {
	suite, err := setupRedisTest()
	assert.NoError(t, err)

	t.Run("get non-existent token", func(t *testing.T) {
		_, err := suite.store.Get(suite.ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, redis.ErrNil, err)
	})

	t.Run("get existing token", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}

		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.NoError(t, err)

		retrieved, err := suite.store.Get(suite.ctx, info.Raw)
		assert.NoError(t, err)
		assert.Equal(t, info.Raw, retrieved.Raw)
	})
}
