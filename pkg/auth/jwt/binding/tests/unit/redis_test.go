package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	errorTypes "gobase/pkg/errors/types"
)

type mockRedisClient struct {
	redis.Client
	setErr  error
	getErr  error
	delErr  error
	getData string
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	return m.getData, nil
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.setErr
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	if m.delErr != nil {
		return 0, m.delErr
	}
	return 1, nil
}

func (m *mockRedisClient) Close() error {
	return nil
}

func TestRedisStore(t *testing.T) {
	ctx := context.Background()

	t.Run("SaveDeviceBinding", func(t *testing.T) {
		mockClient := &mockRedisClient{}
		store, err := binding.NewRedisStore(mockClient)
		require.NoError(t, err)

		device := &binding.DeviceInfo{
			ID:          "test-device",
			Type:        "mobile",
			Name:        "iPhone",
			OS:          "iOS",
			Browser:     "Safari",
			Fingerprint: "test-fp",
		}

		// 测试正常保存
		err = store.SaveDeviceBinding(ctx, "user1", "device1", device)
		require.NoError(t, err)

		// 测试参数验证
		err = store.SaveDeviceBinding(ctx, "", "device1", device)
		assert.Equal(t, codes.InvalidParams, err.(errorTypes.Error).Code())

		// 测试Redis错误
		mockClient.setErr = errors.NewRedisCommandError("connection failed", nil)
		err = store.SaveDeviceBinding(ctx, "user1", "device1", device)
		assert.Equal(t, codes.StoreErrSet, err.(errorTypes.Error).Code())
	})

	t.Run("GetDeviceBinding", func(t *testing.T) {
		mockClient := &mockRedisClient{
			getData: `{"id":"test-device","type":"mobile"}`,
		}
		store, err := binding.NewRedisStore(mockClient)
		require.NoError(t, err)

		// 测试正常获取
		device, err := store.GetDeviceBinding(ctx, "device1")
		require.NoError(t, err)
		assert.Equal(t, "test-device", device.ID)

		// 测试参数验证
		_, err = store.GetDeviceBinding(ctx, "")
		assert.Equal(t, codes.InvalidParams, err.(errorTypes.Error).Code())

		// 测试Redis错误
		mockClient.getErr = errors.NewRedisCommandError("connection failed", nil)
		_, err = store.GetDeviceBinding(ctx, "device1")
		assert.Equal(t, codes.StoreErrGet, err.(errorTypes.Error).Code())
	})

	t.Run("DeleteBinding", func(t *testing.T) {
		mockClient := &mockRedisClient{}
		store, err := binding.NewRedisStore(mockClient)
		require.NoError(t, err)

		// 测试正常删除
		err = store.DeleteBinding(ctx, "device1")
		require.NoError(t, err)

		// 测试参数验证
		err = store.DeleteBinding(ctx, "")
		assert.Equal(t, codes.InvalidParams, err.(errorTypes.Error).Code())

		// 测试Redis错误
		mockClient.delErr = errors.NewRedisCommandError("connection failed", nil)
		err = store.DeleteBinding(ctx, "device1")
		assert.Equal(t, codes.StoreErrDelete, err.(errorTypes.Error).Code())
	})

	t.Run("Close", func(t *testing.T) {
		mockClient := &mockRedisClient{}
		store, err := binding.NewRedisStore(mockClient)
		require.NoError(t, err)
		assert.NoError(t, store.Close())
	})
}
