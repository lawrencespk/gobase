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

// TestStandardClaims_ExpiresAt 测试过期时间相关功能
func TestStandardClaims_ExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *jwt.StandardClaims
		wantErr   bool
		checkFunc func(*testing.T, *jwt.StandardClaims)
	}{
		{
			name: "设置未来过期时间",
			setup: func() *jwt.StandardClaims {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)
				return claims
			},
			wantErr: false,
			checkFunc: func(t *testing.T, claims *jwt.StandardClaims) {
				assert.NotNil(t, claims.ExpiresAt)
				assert.True(t, time.Now().Before(claims.ExpiresAt.Time))
			},
		},
		{
			name: "设置过去过期时间",
			setup: func() *jwt.StandardClaims {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(-time.Hour)),
				)
				return claims
			},
			wantErr: true,
		},
		{
			name: "动态设置过期时间",
			setup: func() *jwt.StandardClaims {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
				)
				claims.SetExpiresAt(time.Now().Add(time.Hour))
				return claims
			},
			wantErr: false,
			checkFunc: func(t *testing.T, claims *jwt.StandardClaims) {
				assert.NotNil(t, claims.ExpiresAt)
				assert.True(t, time.Now().Before(claims.ExpiresAt.Time))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := tt.setup()
			err := claims.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, claims)
			}
		})
	}
}

// TestStandardClaims_Validate 测试Claims验证功能
func TestStandardClaims_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *jwt.StandardClaims
		wantErr bool
		errType error
	}{
		{
			name: "有效的Claims",
			setup: func() *jwt.StandardClaims {
				return jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)
			},
			wantErr: false,
		},
		{
			name: "缺少UserID",
			setup: func() *jwt.StandardClaims {
				return jwt.NewStandardClaims(
					jwt.WithTokenType(jwt.AccessToken),
				)
			},
			wantErr: true,
			errType: jwt.ErrClaimsMissing,
		},
		{
			name: "缺少TokenType",
			setup: func() *jwt.StandardClaims {
				return jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
				)
			},
			wantErr: true,
			errType: jwt.ErrClaimsInvalid,
		},
		{
			name: "已过期",
			setup: func() *jwt.StandardClaims {
				return jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(-time.Hour)),
				)
			},
			wantErr: true,
			errType: jwt.ErrClaimsExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := tt.setup()
			err := claims.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}
