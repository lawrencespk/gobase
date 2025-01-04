package unit

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
)

func TestHMAC_Sign(t *testing.T) {
	tests := []struct {
		name    string
		method  jwt.SigningMethod
		data    []byte
		key     interface{}
		wantErr bool
	}{
		{
			name:    "HS256 Valid",
			method:  jwt.HS256,
			data:    []byte("test data"),
			key:     []byte("secret"),
			wantErr: false,
		},
		{
			name:    "HS384 Valid",
			method:  jwt.HS384,
			data:    []byte("test data"),
			key:     []byte("secret"),
			wantErr: false,
		},
		{
			name:    "HS512 Valid",
			method:  jwt.HS512,
			data:    []byte("test data"),
			key:     []byte("secret"),
			wantErr: false,
		},
		{
			name:    "Invalid Key Type",
			method:  jwt.HS256,
			data:    []byte("test data"),
			key:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, err := crypto.NewHMAC(tt.method)
			require.NoError(t, err)

			sig, err := alg.Sign(tt.data, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, sig)

			// Verify signature
			err = alg.Verify(tt.data, sig, tt.key)
			assert.NoError(t, err)
		})
	}
}

func TestRSA_Sign(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name    string
		method  jwt.SigningMethod
		data    []byte
		key     interface{}
		wantErr bool
	}{
		{
			name:    "RS256 Valid",
			method:  jwt.RS256,
			data:    []byte("test data"),
			key:     privateKey,
			wantErr: false,
		},
		{
			name:    "RS384 Valid",
			method:  jwt.RS384,
			data:    []byte("test data"),
			key:     privateKey,
			wantErr: false,
		},
		{
			name:    "RS512 Valid",
			method:  jwt.RS512,
			data:    []byte("test data"),
			key:     privateKey,
			wantErr: false,
		},
		{
			name:    "Invalid Key Type",
			method:  jwt.RS256,
			data:    []byte("test data"),
			key:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, err := crypto.NewRSA(tt.method)
			require.NoError(t, err)

			sig, err := alg.Sign(tt.data, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, sig)

			// Verify signature
			err = alg.Verify(tt.data, sig, &privateKey.PublicKey)
			assert.NoError(t, err)
		})
	}
}
