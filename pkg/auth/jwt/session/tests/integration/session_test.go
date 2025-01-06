package integration

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Session Lifecycle", func(t *testing.T) {
		// 创建会话
		sess := &session.Session{
			UserID:    "test-user",
			TokenID:   "test-token",
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"device": "mobile",
				"os":     "ios",
			},
		}

		// 保存会话
		err := testStore.Save(ctx, sess)
		require.NoError(t, err)

		// 获取会话
		loaded, err := testStore.Get(ctx, sess.TokenID)
		require.NoError(t, err)
		assert.Equal(t, sess.UserID, loaded.UserID)
		assert.Equal(t, sess.Metadata["device"], loaded.Metadata["device"])

		// 删除会话
		err = testStore.Delete(ctx, sess.TokenID)
		require.NoError(t, err)

		// 验证会话已删除
		_, err = testStore.Get(ctx, sess.TokenID)
		assert.Error(t, err)
		assert.True(t, session.IsRedisKeyNotFoundError(err))
	})

	t.Run("Session Expiration", func(t *testing.T) {
		// 创建过期会话
		sess := &session.Session{
			UserID:    "expired-user",
			TokenID:   "expired-token",
			ExpiresAt: time.Now().Add(-time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 尝试保存过期会话
		err := testStore.Save(ctx, sess)
		assert.Error(t, err)
		assert.True(t, errors.IsTokenExpiredError(err))
	})

	t.Run("Session Refresh", func(t *testing.T) {
		// 创建会话
		sess := &session.Session{
			UserID:    "refresh-user",
			TokenID:   "refresh-token",
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 保存会话
		err := testStore.Save(ctx, sess)
		require.NoError(t, err)

		// 刷新会话
		newExpiration := time.Now().Add(2 * time.Hour)
		err = testStore.Refresh(ctx, sess.TokenID, newExpiration)
		require.NoError(t, err)

		// 验证新的过期时间
		loaded, err := testStore.Get(ctx, sess.TokenID)
		require.NoError(t, err)
		assert.True(t, loaded.ExpiresAt.Equal(newExpiration))
	})
}
