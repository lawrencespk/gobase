package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt"
)

func TestStandardClaims_NewClaims(t *testing.T) {
	tests := []struct {
		name    string
		opts    []jwt.ClaimsOption
		want    *jwt.StandardClaims
		wantErr bool
	}{
		{
			name: "创建基本claims",
			opts: []jwt.ClaimsOption{
				jwt.WithUserID("test-user"),
				jwt.WithUserName("Test User"),
				jwt.WithTokenType(jwt.AccessToken),
			},
			wantErr: false,
		},
		{
			name: "创建完整claims",
			opts: []jwt.ClaimsOption{
				jwt.WithUserID("test-user"),
				jwt.WithUserName("Test User"),
				jwt.WithRoles([]string{"admin"}),
				jwt.WithPermissions([]string{"read", "write"}),
				jwt.WithDeviceID("device-123"),
				jwt.WithIPAddress("127.0.0.1"),
				jwt.WithTokenType(jwt.AccessToken),
				jwt.WithTokenID("token-123"),
				jwt.WithExpiresAt(time.Now().Add(time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "缺少必填字段",
			opts: []jwt.ClaimsOption{
				jwt.WithUserName("Test User"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := jwt.NewStandardClaims(tt.opts...)
			err := claims.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, claims.GetUserID())
			if claims.GetTokenType() == jwt.AccessToken {
				assert.Equal(t, jwt.AccessToken, claims.GetTokenType())
			}
		})
	}
}

func TestStandardClaims_GettersAndSetters(t *testing.T) {
	claims := jwt.NewStandardClaims()

	// 测试基本字段
	claims.UserID = "test-user"
	assert.Equal(t, "test-user", claims.GetUserID())

	claims.UserName = "Test User"
	assert.Equal(t, "Test User", claims.GetUserName())

	// 测试角色和权限
	roles := []string{"admin", "user"}
	claims.Roles = roles
	assert.Equal(t, roles, claims.GetRoles())

	permissions := []string{"read", "write"}
	claims.Permissions = permissions
	assert.Equal(t, permissions, claims.GetPermissions())

	// 测试设备和IP
	claims.DeviceID = "device-123"
	assert.Equal(t, "device-123", claims.GetDeviceID())

	claims.IPAddress = "127.0.0.1"
	assert.Equal(t, "127.0.0.1", claims.GetIPAddress())

	// 测试Token相关
	claims.TokenType = jwt.AccessToken
	assert.Equal(t, jwt.AccessToken, claims.GetTokenType())

	claims.TokenID = "token-123"
	assert.Equal(t, "token-123", claims.GetTokenID())
}
