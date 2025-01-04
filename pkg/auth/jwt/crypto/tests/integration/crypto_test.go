package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
	"gobase/pkg/auth/jwt/crypto/tests/mock"
)

func TestCryptoIntegration(t *testing.T) {
	logger := mock.NewMockLogger()
	ctx := context.Background()
	testData := []byte("test data")

	tests := []struct {
		name   string
		method jwt.SigningMethod
		config *jwt.KeyConfig
	}{
		{
			name:   "HMAC End-to-End",
			method: jwt.HS256,
			config: &jwt.KeyConfig{
				SecretKey: "test-secret",
			},
		},
		{
			name:   "RSA End-to-End",
			method: jwt.RS256,
			config: &jwt.KeyConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create key manager
			km, err := crypto.NewKeyManager(tt.method, logger)
			require.NoError(t, err)

			// Initialize keys
			err = km.InitializeKeys(ctx, tt.config)
			require.NoError(t, err)

			// Create algorithm
			alg, err := crypto.CreateAlgorithm(tt.method)
			require.NoError(t, err)

			// Get signing key
			signingKey, err := km.GetSigningKey()
			require.NoError(t, err)

			// Sign data
			signature, err := alg.Sign(testData, signingKey)
			require.NoError(t, err)

			// Get verification key
			verificationKey, err := km.GetVerificationKey()
			require.NoError(t, err)

			// Verify signature
			err = alg.Verify(testData, signature, verificationKey)
			assert.NoError(t, err)

			// For RSA, test key rotation
			if tt.method == jwt.RS256 {
				err = km.RotateKeys(ctx)
				require.NoError(t, err)

				// Old signature should fail verification with new keys
				newVerificationKey, err := km.GetVerificationKey()
				require.NoError(t, err)
				err = alg.Verify(testData, signature, newVerificationKey)
				assert.Error(t, err)
			}
		})
	}
}
