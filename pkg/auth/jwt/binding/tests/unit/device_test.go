package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/auth/jwt/binding/tests/mock"
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

func TestDeviceValidator_ValidateDevice(t *testing.T) {
	// 初始化
	store := mock.NewMockStore()
	validator, err := binding.NewDeviceValidator(store)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func() (jwt.Claims, *binding.DeviceInfo)
		wantError bool
		errorCode string
	}{
		{
			name: "验证成功 - 设备匹配",
			setup: func() (jwt.Claims, *binding.DeviceInfo) {
				claims := jwt.NewStandardClaims(
					jwt.WithTokenID("test-token"),
					jwt.WithDeviceID("device-123"),
					jwt.WithUserID("user-123"),
				)
				device := &binding.DeviceInfo{
					ID:          "device-123",
					Type:        "mobile",
					Name:        "iPhone",
					OS:          "iOS",
					Browser:     "Safari",
					Fingerprint: "fp-123",
				}
				err := store.SaveDeviceBinding(ctx, claims.GetUserID(), claims.GetDeviceID(), device)
				require.NoError(t, err)
				return claims, device
			},
			wantError: false,
		},
		{
			name: "验证失败 - 设备不匹配",
			setup: func() (jwt.Claims, *binding.DeviceInfo) {
				claims := jwt.NewStandardClaims(
					jwt.WithTokenID("test-token-2"),
					jwt.WithDeviceID("device-456"),
					jwt.WithUserID("user-456"),
				)
				device := &binding.DeviceInfo{
					ID:          "device-456",
					Fingerprint: "fp-456",
				}
				err := store.SaveDeviceBinding(ctx, claims.GetUserID(), claims.GetDeviceID(), device)
				require.NoError(t, err)
				return claims, &binding.DeviceInfo{
					ID:          "device-123",
					Fingerprint: "fp-123",
				}
			},
			wantError: true,
			errorCode: codes.BindingMismatch,
		},
		{
			name: "验证失败 - 绑定不存在",
			setup: func() (jwt.Claims, *binding.DeviceInfo) {
				claims := jwt.NewStandardClaims(
					jwt.WithTokenID("test-token-3"),
					jwt.WithDeviceID("device-789"),
					jwt.WithUserID("user-789"),
				)
				return claims, &binding.DeviceInfo{
					ID:          "device-123",
					Fingerprint: "fp-123",
				}
			},
			wantError: true,
			errorCode: codes.StoreErrNotFound,
		},
		{
			name: "验证失败 - 设备信息无效",
			setup: func() (jwt.Claims, *binding.DeviceInfo) {
				claims := jwt.NewStandardClaims(
					jwt.WithTokenID("test-token-4"),
					jwt.WithDeviceID("device-000"),
					jwt.WithUserID("user-000"),
				)
				return claims, &binding.DeviceInfo{
					ID: "", // 无效的设备ID
				}
			},
			wantError: true,
			errorCode: codes.BindingInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, device := tt.setup()
			err := validator.ValidateDevice(ctx, claims, device)

			if tt.wantError {
				assert.Error(t, err)
				var customErr types.Error
				assert.ErrorAs(t, err, &customErr)
				assert.Equal(t, tt.errorCode, customErr.Code())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
