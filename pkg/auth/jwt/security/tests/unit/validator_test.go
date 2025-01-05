package unit

import (
	"context"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/security"
	"gobase/pkg/errors/types"
)

// mockClaims 实现jwt.Claims接口
type mockClaims struct {
	// 业务字段
	userID      string
	userName    string
	roles       []string
	permissions []string
	deviceID    string
	ipAddress   string
	tokenType   jwt.TokenType
	tokenID     string
	expiresAt   time.Time
	audience    []string
}

// 实现标准 jwt.Claims 接口的方法
func (m *mockClaims) GetExpirationTime() (*jwtv5.NumericDate, error) {
	if m.expiresAt.IsZero() {
		return nil, nil
	}
	return jwtv5.NewNumericDate(m.expiresAt), nil
}

func (m *mockClaims) GetIssuedAt() (*jwtv5.NumericDate, error)  { return nil, nil }
func (m *mockClaims) GetNotBefore() (*jwtv5.NumericDate, error) { return nil, nil }
func (m *mockClaims) GetIssuer() (string, error)                { return "", nil }
func (m *mockClaims) GetSubject() (string, error)               { return "", nil }
func (m *mockClaims) GetAudience() (jwtv5.ClaimStrings, error)  { return m.audience, nil }

// 实现我们自己的Claims接口的方法
func (m *mockClaims) GetUserID() string           { return m.userID }
func (m *mockClaims) GetUserName() string         { return m.userName }
func (m *mockClaims) GetRoles() []string          { return m.roles }
func (m *mockClaims) GetPermissions() []string    { return m.permissions }
func (m *mockClaims) GetDeviceID() string         { return m.deviceID }
func (m *mockClaims) GetIPAddress() string        { return m.ipAddress }
func (m *mockClaims) GetTokenType() jwt.TokenType { return m.tokenType }
func (m *mockClaims) GetTokenID() string          { return m.tokenID }

// 实现Claims接口的Validate方法
func (m *mockClaims) Validate() error {
	if m.userID == "" {
		return jwt.ErrClaimsMissing
	}
	if time.Now().After(m.expiresAt) {
		return jwt.ErrClaimsExpired
	}
	return nil
}

func TestTokenValidator_ValidateToken(t *testing.T) {
	ctx := context.Background()

	policy := &security.Policy{
		MaxTokenAge:         time.Hour,
		EnableDeviceBinding: true,
		EnableIPBinding:     true,
	}

	validator := security.NewTokenValidator(policy)

	t.Run("有效Token", func(t *testing.T) {
		claims := &mockClaims{
			userID:    "user-1",
			deviceID:  "device-1",
			ipAddress: "192.168.1.1",
			tokenType: jwt.AccessToken,
			expiresAt: time.Now().Add(30 * time.Minute),
		}

		tokenInfo := &jwt.TokenInfo{
			Type:      jwt.AccessToken,
			Claims:    claims,
			ExpiresAt: claims.expiresAt,
		}

		err := validator.ValidateToken(ctx, tokenInfo)
		assert.NoError(t, err)
	})

	t.Run("过期Token", func(t *testing.T) {
		claims := &mockClaims{
			userID:    "user-1",
			deviceID:  "device-1",
			ipAddress: "192.168.1.1",
			tokenType: jwt.AccessToken,
			expiresAt: time.Now().Add(-time.Minute),
		}

		tokenInfo := &jwt.TokenInfo{
			Type:      jwt.AccessToken,
			Claims:    claims,
			ExpiresAt: claims.expiresAt,
		}

		err := validator.ValidateToken(ctx, tokenInfo)
		assert.Error(t, err)

		var tokenErr types.Error
		assert.ErrorAs(t, err, &tokenErr)
		assert.Equal(t, "2121", tokenErr.Code())
	})

	t.Run("缺少设备ID", func(t *testing.T) {
		claims := &mockClaims{
			userID:    "user-1",
			ipAddress: "192.168.1.1",
			tokenType: jwt.AccessToken,
			expiresAt: time.Now().Add(30 * time.Minute),
		}

		tokenInfo := &jwt.TokenInfo{
			Type:      jwt.AccessToken,
			Claims:    claims,
			ExpiresAt: claims.expiresAt,
		}

		err := validator.ValidateToken(ctx, tokenInfo)
		assert.Error(t, err)

		var bindingErr types.Error
		assert.ErrorAs(t, err, &bindingErr)
		assert.Equal(t, "2160", bindingErr.Code())
	})

	t.Run("无效IP地址", func(t *testing.T) {
		claims := &mockClaims{
			userID:    "user-1",
			deviceID:  "device-1",
			ipAddress: "invalid-ip",
			tokenType: jwt.AccessToken,
			expiresAt: time.Now().Add(30 * time.Minute),
		}

		tokenInfo := &jwt.TokenInfo{
			Type:      jwt.AccessToken,
			Claims:    claims,
			ExpiresAt: claims.expiresAt,
		}

		err := validator.ValidateToken(ctx, tokenInfo)
		assert.Error(t, err)

		var bindingErr types.Error
		assert.ErrorAs(t, err, &bindingErr)
		assert.Equal(t, "2160", bindingErr.Code())
	})
}
