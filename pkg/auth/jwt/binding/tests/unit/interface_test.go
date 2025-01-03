package unit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/auth/jwt/binding/tests/mock"
	"gobase/pkg/errors/codes"
)

func TestStoreInterface(t *testing.T) {
	store := mock.NewMockStore()

	ctx := context.Background()
	userID := "test-user"
	tokenID := "test-token"

	t.Run("IP binding", func(t *testing.T) {
		// 测试正常保存和获取
		ip := "127.0.0.1"
		err := store.SaveIPBinding(ctx, userID, tokenID, ip)
		require.NoError(t, err)

		gotIP, err := store.GetIPBinding(ctx, tokenID)
		require.NoError(t, err)
		assert.Equal(t, ip, gotIP)

		// 测试错误情况
		store.SetError(true)
		_, err = store.GetIPBinding(ctx, tokenID)
		require.Error(t, err)
		assert.Equal(t, codes.CacheError, err.(interface{ Code() string }).Code())
	})

	t.Run("Device binding", func(t *testing.T) {
		// 重置错误状态
		store.SetError(false)

		// 测试正常保存和获取
		device := &binding.DeviceInfo{
			ID:          "device-id",
			Fingerprint: "device-fingerprint",
		}
		err := store.SaveDeviceBinding(ctx, userID, tokenID, device)
		require.NoError(t, err)

		gotDevice, err := store.GetDeviceBinding(ctx, tokenID)
		require.NoError(t, err)
		assert.Equal(t, device, gotDevice)

		// 测试错误情况
		store.SetError(true)
		_, err = store.GetDeviceBinding(ctx, tokenID)
		require.Error(t, err)
		assert.Equal(t, codes.CacheError, err.(interface{ Code() string }).Code())
	})

	t.Run("Delete binding", func(t *testing.T) {
		// 重置错误状态
		store.SetError(false)

		// 先保存一些数据
		ip := "127.0.0.1"
		device := &binding.DeviceInfo{
			ID:          "device-id",
			Fingerprint: "device-fingerprint",
		}

		err := store.SaveIPBinding(ctx, userID, tokenID, ip)
		require.NoError(t, err)

		err = store.SaveDeviceBinding(ctx, userID, tokenID, device)
		require.NoError(t, err)

		// 测试删除
		err = store.DeleteBinding(ctx, tokenID)
		require.NoError(t, err)

		// 验证数据已被删除
		_, err = store.GetIPBinding(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())

		_, err = store.GetDeviceBinding(ctx, tokenID)
		assert.Equal(t, codes.StoreErrNotFound, err.(interface{ Code() string }).Code())

		// 测试错误情况
		store.SetError(true)
		err = store.DeleteBinding(ctx, tokenID)
		require.Error(t, err)
		assert.Equal(t, codes.CacheError, err.(interface{ Code() string }).Code())
	})
}
