package unit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/auth/jwt/binding/tests/mock"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func TestIPValidator_ValidateIP(t *testing.T) {
	// 初始化
	log, err := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	require.NoError(t, err)

	store := mock.NewMockStore()

	tests := []struct {
		name      string
		claims    jwt.Claims
		currentIP string
		boundIP   string
		storeErr  bool
		wantErr   bool
		errorCode string
	}{
		{
			name: "成功匹配",
			claims: &jwt.StandardClaims{
				TokenID: "test-token",
			},
			currentIP: "127.0.0.1",
			boundIP:   "127.0.0.1",
			storeErr:  false,
			wantErr:   false,
		},
		{
			name: "IP不匹配",
			claims: &jwt.StandardClaims{
				TokenID: "test-token",
			},
			currentIP: "127.0.0.1",
			boundIP:   "192.168.1.1",
			storeErr:  false,
			wantErr:   true,
			errorCode: codes.BindingMismatch,
		},
		{
			name: "存储错误",
			claims: &jwt.StandardClaims{
				TokenID: "test-token",
			},
			currentIP: "127.0.0.1",
			storeErr:  true,
			wantErr:   true,
			errorCode: codes.CacheError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := binding.NewValidator(store, binding.WithLogger(log))

			store.SetError(tt.storeErr)

			if !tt.storeErr && tt.boundIP != "" {
				err := store.SaveIPBinding(context.Background(), "test-user", tt.claims.GetTokenID(), tt.boundIP)
				require.NoError(t, err)
			}

			err := validator.ValidateIP(context.Background(), tt.claims, tt.currentIP)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorCode != "" {
					var customErr interface {
						Code() string
					}
					if assert.ErrorAs(t, err, &customErr) {
						assert.Equal(t, tt.errorCode, customErr.Code())
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
