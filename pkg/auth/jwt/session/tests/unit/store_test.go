package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/auth/jwt/session/tests/mock"
	"gobase/pkg/errors"

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestSessionStore(t *testing.T) {
	ctx := context.Background()
	logger := mock.NewMockLogger()

	opts := &session.Options{
		Redis: &session.RedisOptions{
			Addr: "localhost:6379",
		},
		KeyPrefix:     "test:",
		EnableMetrics: true,
		Log:           logger,
	}

	t.Run("Save and Load Session", func(t *testing.T) {
		// 为每个测试用例创建新的 mock 客户端和 store
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		now := time.Now()
		expiresAt := now.Add(time.Hour)
		sess := &session.Session{
			UserID:    "user-1",
			TokenID:   "token-1",
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
			Metadata: map[string]interface{}{
				"key": "value",
			},
		}

		pipeline := mock.NewMockPipeline()
		client.On("TxPipeline").Return(pipeline)

		pipeline.On("Set", tmock.AnythingOfType("*context.valueCtx"), "test:session:token-1", tmock.Anything, tmock.Anything).Return(nil)
		pipeline.On("SAdd", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:user-1", "token-1").Return(int64(1), nil)
		pipeline.On("ExpireAt", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:user-1", tmock.Anything).Return(true, nil)
		pipeline.On("Exec", tmock.AnythingOfType("*context.valueCtx")).Return(nil, nil)
		pipeline.On("Close").Return(nil)

		err := store.Save(ctx, sess)
		assert.NoError(t, err)

		expectedJSON := `{"user_id":"user-1","token_id":"token-1","expires_at":"` + expiresAt.Format(time.RFC3339) + `","created_at":"` + now.Format(time.RFC3339) + `","updated_at":"` + now.Format(time.RFC3339) + `","metadata":{"key":"value"}}`
		client.On("Get", tmock.AnythingOfType("*context.valueCtx"), "test:session:token-1").Return(expectedJSON, nil)

		loaded, err := store.Get(ctx, sess.TokenID)
		assert.NoError(t, err)
		assert.Equal(t, sess.UserID, loaded.UserID)
		assert.Equal(t, sess.TokenID, loaded.TokenID)
		assert.Equal(t, sess.Metadata["key"], loaded.Metadata["key"])
	})

	t.Run("Session Expiration", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		sess := &session.Session{
			TokenID:   "expired-token",
			ExpiresAt: time.Now().Add(-time.Hour),
		}

		err := store.Save(ctx, sess)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.NewSessionExpiredError("session already expired", nil)))
	})

	t.Run("Delete Session", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		sess := &session.Session{
			UserID:    "user-2",
			TokenID:   "token-to-delete",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		pipeline := mock.NewMockPipeline()
		client.On("TxPipeline").Return(pipeline)

		pipeline.On("Set", tmock.AnythingOfType("*context.valueCtx"), "test:session:token-to-delete", tmock.Anything, tmock.Anything).Return(nil)
		pipeline.On("SAdd", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:user-2", "token-to-delete").Return(int64(1), nil)
		pipeline.On("ExpireAt", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:user-2", tmock.Anything).Return(true, nil)
		pipeline.On("Exec", tmock.AnythingOfType("*context.valueCtx")).Return(nil, nil)
		pipeline.On("Close").Return(nil)

		err := store.Save(ctx, sess)
		assert.NoError(t, err)

		client.On("Del", tmock.AnythingOfType("*context.valueCtx"), []string{"test:session:token-to-delete"}).Return(int64(1), nil)
		err = store.Delete(ctx, sess.TokenID)
		assert.NoError(t, err)

		client.On("Get", tmock.AnythingOfType("*context.valueCtx"), "test:session:token-to-delete").Return("", errors.NewRedisKeyNotFoundError("key not found", nil))
		_, err = store.Get(ctx, sess.TokenID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.NewCacheError("failed to get value", errors.NewRedisKeyNotFoundError("key not found", nil))))
	})
}
