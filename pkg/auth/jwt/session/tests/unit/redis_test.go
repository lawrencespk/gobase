package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/auth/jwt/session/tests/mock"

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestRedisStore(t *testing.T) {
	ctx := context.Background()

	opts := &session.Options{
		Redis: &session.RedisOptions{
			Addr: "localhost:6379",
		},
		KeyPrefix:     "test:",
		EnableMetrics: true,
		Log:           &mockLogger{},
	}

	t.Run("NewRedisStore", func(t *testing.T) {
		store := session.NewRedisStore(mock.NewMockClient(), opts)
		assert.NotNil(t, store)
	})

	t.Run("Basic Operations", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		sess := &session.Session{
			TokenID:   "key",
			UserID:    "test-user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// 创建 pipeline mock
		pipeline := mock.NewMockPipeline()
		client.On("TxPipeline").Return(pipeline)

		// 设置 pipeline 操作的期望
		pipeline.On("Set", tmock.AnythingOfType("*context.valueCtx"), "test:session:key", tmock.AnythingOfType("string"), tmock.AnythingOfType("time.Duration")).Return(nil)
		pipeline.On("SAdd", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:test-user", "key").Return(int64(1), nil)
		pipeline.On("ExpireAt", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:test-user", tmock.AnythingOfType("time.Time")).Return(true, nil)
		pipeline.On("Exec", tmock.AnythingOfType("*context.valueCtx")).Return(nil, nil)

		// Test Save
		err := store.Save(ctx, sess)
		assert.NoError(t, err)

		// Test Get
		sessionData, err := json.Marshal(sess)
		assert.NoError(t, err)
		client.On("Get", tmock.AnythingOfType("*context.valueCtx"), "test:session:key").Return(string(sessionData), nil)

		got, err := store.Get(ctx, "key")
		assert.NoError(t, err)
		assert.NotNil(t, got)

		// Test Delete
		client.On("Del", tmock.AnythingOfType("*context.valueCtx"), []string{"test:session:key"}).Return(int64(1), nil)
		err = store.Delete(ctx, "key")
		assert.NoError(t, err)

		client.AssertExpectations(t)
		pipeline.AssertExpectations(t)
	})

	t.Run("Close", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		client.On("Close").Return(nil)
		err := store.Close(ctx)
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Ping", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		client.On("Ping", tmock.AnythingOfType("*context.valueCtx")).Return(nil)
		err := store.Ping(ctx)
		assert.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Refresh", func(t *testing.T) {
		client := mock.NewMockClient()
		store := session.NewRedisStore(client, opts)

		newExpiration := time.Now().Add(2 * time.Hour)
		sess := &session.Session{
			TokenID:   "key",
			UserID:    "test-user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// 序列化会话数据
		data, err := json.Marshal(sess)
		assert.NoError(t, err)

		// Mock Get
		client.On("Get", tmock.AnythingOfType("*context.valueCtx"), "test:session:key").Return(string(data), nil)

		// Mock Save 操作
		pipeline := mock.NewMockPipeline()
		client.On("TxPipeline").Return(pipeline)

		// 设置 pipeline 操作的期望
		pipeline.On("Set", tmock.AnythingOfType("*context.valueCtx"), "test:session:key", tmock.AnythingOfType("string"), tmock.AnythingOfType("time.Duration")).Return(nil)
		pipeline.On("SAdd", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:test-user", "key").Return(int64(1), nil)
		pipeline.On("ExpireAt", tmock.AnythingOfType("*context.valueCtx"), "test:session:user:test-user", tmock.AnythingOfType("time.Time")).Return(true, nil)
		pipeline.On("Exec", tmock.AnythingOfType("*context.valueCtx")).Return(nil, nil)

		err = store.Refresh(ctx, "key", newExpiration)
		assert.NoError(t, err)

		client.AssertExpectations(t)
		pipeline.AssertExpectations(t)
	})
}
