package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/errors/codes"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := blacklist.NewMemoryStore()
	defer store.Close()

	tokenID := "test-token"
	reason := "test blacklist"
	expiration := time.Hour * 24

	t.Run("Basic Operations", func(t *testing.T) {
		// 测试添加
		err := store.Add(ctx, tokenID, reason, expiration)
		require.NoError(t, err)

		// 测试获取
		gotReason, err := store.Get(ctx, tokenID)
		require.NoError(t, err)
		assert.Equal(t, reason, gotReason)

		// 测试删除
		err = store.Remove(ctx, tokenID)
		require.NoError(t, err)

		// 验证删除成功
		_, err = store.Get(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		// 测试空TokenID
		err := store.Add(ctx, "", reason, expiration)
		assert.Equal(t, codes.InvalidParams, err.(interface{ Code() string }).Code())

		// 测试负过期时间
		err = store.Add(ctx, tokenID, reason, -time.Hour)
		assert.Equal(t, codes.InvalidParams, err.(interface{ Code() string }).Code())
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				tid := fmt.Sprintf("%s-%d", tokenID, id)
				_ = store.Add(ctx, tid, reason, expiration)
				_, _ = store.Get(ctx, tid)
				_ = store.Remove(ctx, tid)
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("Expiration", func(t *testing.T) {
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

	t.Run("Cleanup", func(t *testing.T) {
		// 添加多个过期项
		shortExpiration := time.Millisecond * 100
		for i := 0; i < 10; i++ {
			tid := fmt.Sprintf("%s-%d", tokenID, i)
			_ = store.Add(ctx, tid, reason, shortExpiration)
		}

		// 等待过期
		time.Sleep(shortExpiration * 2)

		// 触发清理
		store.(*blacklist.MemoryStore).Cleanup()

		// 验证所有项都已被清理
		for i := 0; i < 10; i++ {
			tid := fmt.Sprintf("%s-%d", tokenID, i)
			_, err := store.Get(ctx, tid)
			assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())
		}
	})
}
