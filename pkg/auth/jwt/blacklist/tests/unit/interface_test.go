package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/blacklist/tests/mock"
	"gobase/pkg/errors/codes"
)

func TestStoreInterface(t *testing.T) {
	store := mock.NewMockStore()
	ctx := context.Background()
	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Hour * 24

	t.Run("Add and Get", func(t *testing.T) {
		// 重置错误状态
		store.SetError(false)

		// 测试正常添加和获取
		err := store.Add(ctx, tokenID, reason, expiration)
		require.NoError(t, err)

		gotReason, err := store.Get(ctx, tokenID)
		require.NoError(t, err)
		assert.Equal(t, reason, gotReason)

		// 测试错误情况
		store.SetError(true)
		_, err = store.Get(ctx, tokenID)
		require.Error(t, err)
		assert.Equal(t, codes.CacheError, err.(interface{ Code() string }).Code())
	})

	t.Run("Remove", func(t *testing.T) {
		// 重置错误状态
		store.SetError(false)

		// 先添加一条记录
		err := store.Add(ctx, tokenID, reason, expiration)
		require.NoError(t, err)

		// 测试删除
		err = store.Remove(ctx, tokenID)
		require.NoError(t, err)

		// 验证已被删除
		_, err = store.Get(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())

		// 测试错误情况
		store.SetError(true)
		err = store.Remove(ctx, tokenID)
		require.Error(t, err)
		assert.Equal(t, codes.CacheError, err.(interface{ Code() string }).Code())
	})

	t.Run("Expiration", func(t *testing.T) {
		// 重置错误状态
		store.SetError(false)

		// 测试短期过期
		shortExpiration := time.Millisecond * 100
		err := store.Add(ctx, tokenID, reason, shortExpiration)
		require.NoError(t, err)

		// 等待过期
		time.Sleep(shortExpiration * 2)

		// 验证已过期
		_, err = store.Get(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())
	})
}
