package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
	"gobase/pkg/auth/jwt/crypto/tests/mock"
)

func TestNewKeyManager(t *testing.T) {
	tests := []struct {
		name        string
		method      jwt.SigningMethod
		expectError bool
	}{
		{
			name:        "Valid HMAC method",
			method:      jwt.HS256,
			expectError: false,
		},
		{
			name:        "Valid RSA method",
			method:      jwt.RS256,
			expectError: false,
		},
		{
			name:        "Invalid method",
			method:      "invalid",
			expectError: true,
		},
	}

	logger := mock.NewMockLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := crypto.NewKeyManager(tt.method, logger)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, km)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, km)
			}
		})
	}
}

func TestKeyManager_InitializeKeys(t *testing.T) {
	ctx := context.Background()
	logger := mock.NewMockLogger()

	tests := []struct {
		name        string
		method      jwt.SigningMethod
		config      *jwt.KeyConfig
		expectError bool
	}{
		{
			name:   "Initialize HMAC with valid secret",
			method: jwt.HS256,
			config: &jwt.KeyConfig{
				SecretKey: "test-secret-key",
			},
			expectError: false,
		},
		{
			name:        "Initialize HMAC without secret",
			method:      jwt.HS256,
			config:      nil,
			expectError: true,
		},
		{
			name:   "Initialize RSA with provided key",
			method: jwt.RS256,
			config: &jwt.KeyConfig{
				PrivateKey: mock.TestRSAPrivateKeyPEM,
			},
			expectError: false,
		},
		{
			name:        "Initialize RSA with auto-generation",
			method:      jwt.RS256,
			config:      nil,
			expectError: false,
		},
		{
			name:   "Initialize RSA with invalid key",
			method: jwt.RS256,
			config: &jwt.KeyConfig{
				PrivateKey: "invalid-key",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := crypto.NewKeyManager(tt.method, logger)
			require.NoError(t, err)

			err = km.InitializeKeys(ctx, tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证密钥是否正确初始化
				signingKey, err := km.GetSigningKey()
				assert.NoError(t, err)
				assert.NotNil(t, signingKey)

				verificationKey, err := km.GetVerificationKey()
				assert.NoError(t, err)
				assert.NotNil(t, verificationKey)
			}
		})
	}
}

func TestKeyManager_GetKeys(t *testing.T) {
	ctx := context.Background()
	logger := mock.NewMockLogger()

	t.Run("Get keys before initialization", func(t *testing.T) {
		km, err := crypto.NewKeyManager(jwt.HS256, logger)
		require.NoError(t, err)

		_, err = km.GetSigningKey()
		assert.Error(t, err)

		_, err = km.GetVerificationKey()
		assert.Error(t, err)
	})

	t.Run("Get keys after initialization", func(t *testing.T) {
		km, err := crypto.NewKeyManager(jwt.HS256, logger)
		require.NoError(t, err)

		err = km.InitializeKeys(ctx, &jwt.KeyConfig{
			SecretKey: "test-secret",
		})
		require.NoError(t, err)

		signingKey, err := km.GetSigningKey()
		assert.NoError(t, err)
		assert.NotNil(t, signingKey)

		verificationKey, err := km.GetVerificationKey()
		assert.NoError(t, err)
		assert.NotNil(t, verificationKey)
	})
}

func TestKeyManager_RotateKeys(t *testing.T) {
	ctx := context.Background()
	logger := mock.NewMockLogger()

	tests := []struct {
		name        string
		method      jwt.SigningMethod
		config      *jwt.KeyConfig
		expectError bool
	}{
		{
			name:   "Rotate HMAC keys",
			method: jwt.HS256,
			config: &jwt.KeyConfig{
				SecretKey: "test-secret",
			},
			expectError: true, // HMAC不支持轮转
		},
		{
			name:   "Rotate RSA keys",
			method: jwt.RS256,
			config: &jwt.KeyConfig{
				PrivateKey: mock.TestRSAPrivateKeyPEM,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := crypto.NewKeyManager(tt.method, logger)
			require.NoError(t, err)

			err = km.InitializeKeys(ctx, tt.config)
			require.NoError(t, err)

			// 记录原始密钥
			originalSigningKey, err := km.GetSigningKey()
			require.NoError(t, err)

			// 执行轮转
			err = km.RotateKeys(ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// 验证新密钥是否不同于原始密钥
				newSigningKey, err := km.GetSigningKey()
				assert.NoError(t, err)
				assert.NotEqual(t, originalSigningKey, newSigningKey)
			}
		})
	}
}

func TestKeyManager_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	logger := mock.NewMockLogger()
	km, err := crypto.NewKeyManager(jwt.RS256, logger)
	require.NoError(t, err)

	err = km.InitializeKeys(ctx, nil)
	require.NoError(t, err)

	// 并发访问测试
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, err := km.GetSigningKey()
				assert.NoError(t, err)
				_, err = km.GetVerificationKey()
				assert.NoError(t, err)
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
