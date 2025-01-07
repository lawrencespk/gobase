package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	jwtContext "gobase/pkg/middleware/jwt/context"
)

// MockClaims 用于测试的Claims实现
type MockClaims struct {
	*jwt.StandardClaims
}

func TestContextOperations(t *testing.T) {
	// 准备测试数据
	testData := struct {
		claims      jwt.Claims
		token       string
		tokenType   jwt.TokenType
		userID      string
		userName    string
		roles       []string
		permissions []string
		deviceID    string
		ipAddress   string
	}{
		claims:      &MockClaims{jwt.NewStandardClaims()},
		token:       "test-token",
		tokenType:   jwt.AccessToken,
		userID:      "test-user",
		userName:    "Test User",
		roles:       []string{"admin", "user"},
		permissions: []string{"read", "write"},
		deviceID:    "test-device",
		ipAddress:   "127.0.0.1",
	}

	tests := []struct {
		name string
		fn   func(t *testing.T, ctx context.Context)
	}{
		{
			name: "Claims操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithClaims(ctx, testData.claims)
				claims, err := jwtContext.GetClaims(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.claims, claims)

				// 测试错误情况
				claims, err = jwtContext.GetClaims(context.Background())
				assert.Error(t, err)
				assert.Nil(t, claims)
			},
		},
		{
			name: "Token操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithToken(ctx, testData.token)
				token, err := jwtContext.GetToken(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.token, token)

				// 测试错误情况
				token, err = jwtContext.GetToken(context.Background())
				assert.Error(t, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "TokenType操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithTokenType(ctx, testData.tokenType)
				tokenType, err := jwtContext.GetTokenType(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.tokenType, tokenType)

				// 测试错误情况
				tokenType, err = jwtContext.GetTokenType(context.Background())
				assert.Error(t, err)
				assert.Empty(t, tokenType)
			},
		},
		{
			name: "UserID操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithUserID(ctx, testData.userID)
				userID, err := jwtContext.GetUserID(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.userID, userID)

				// 测试错误情况
				userID, err = jwtContext.GetUserID(context.Background())
				assert.Error(t, err)
				assert.Empty(t, userID)
			},
		},
		{
			name: "UserName操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithUserName(ctx, testData.userName)
				userName, err := jwtContext.GetUserName(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.userName, userName)

				// 测试错误情况
				userName, err = jwtContext.GetUserName(context.Background())
				assert.Error(t, err)
				assert.Empty(t, userName)
			},
		},
		{
			name: "Roles操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithRoles(ctx, testData.roles)
				roles, err := jwtContext.GetRoles(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.roles, roles)

				// 测试错误情况
				roles, err = jwtContext.GetRoles(context.Background())
				assert.Error(t, err)
				assert.Nil(t, roles)
			},
		},
		{
			name: "Permissions操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithPermissions(ctx, testData.permissions)
				permissions, err := jwtContext.GetPermissions(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.permissions, permissions)

				// 测试错误情况
				permissions, err = jwtContext.GetPermissions(context.Background())
				assert.Error(t, err)
				assert.Nil(t, permissions)
			},
		},
		{
			name: "DeviceID操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithDeviceID(ctx, testData.deviceID)
				deviceID, err := jwtContext.GetDeviceID(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.deviceID, deviceID)

				// 测试错误情况
				deviceID, err = jwtContext.GetDeviceID(context.Background())
				assert.Error(t, err)
				assert.Empty(t, deviceID)
			},
		},
		{
			name: "IPAddress操作",
			fn: func(t *testing.T, ctx context.Context) {
				ctx = jwtContext.WithIPAddress(ctx, testData.ipAddress)
				ipAddress, err := jwtContext.GetIPAddress(ctx)
				require.NoError(t, err)
				assert.Equal(t, testData.ipAddress, ipAddress)

				// 测试错误情况
				ipAddress, err = jwtContext.GetIPAddress(context.Background())
				assert.Error(t, err)
				assert.Empty(t, ipAddress)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, context.Background())
		})
	}
}

func TestWithJWTContext(t *testing.T) {
	// 准备测试数据
	claims := &MockClaims{
		StandardClaims: jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithUserName("Test User"),
			jwt.WithRoles([]string{"admin", "user"}),
			jwt.WithPermissions([]string{"read", "write"}),
			jwt.WithDeviceID("test-device"),
			jwt.WithIPAddress("127.0.0.1"),
			jwt.WithTokenType(jwt.AccessToken),
		),
	}
	token := "test-token"

	// 测试WithJWTContext
	ctx := context.Background()
	ctx = jwtContext.WithJWTContext(ctx, claims, token)

	// 验证所有值是否正确设置
	tests := []struct {
		name     string
		getValue func() (interface{}, error)
		want     interface{}
	}{
		{
			name: "验证Claims",
			getValue: func() (interface{}, error) {
				return jwtContext.GetClaims(ctx)
			},
			want: claims,
		},
		{
			name: "验证Token",
			getValue: func() (interface{}, error) {
				return jwtContext.GetToken(ctx)
			},
			want: token,
		},
		{
			name: "验证TokenType",
			getValue: func() (interface{}, error) {
				return jwtContext.GetTokenType(ctx)
			},
			want: jwt.AccessToken,
		},
		{
			name: "验证UserID",
			getValue: func() (interface{}, error) {
				return jwtContext.GetUserID(ctx)
			},
			want: "test-user",
		},
		{
			name: "验证UserName",
			getValue: func() (interface{}, error) {
				return jwtContext.GetUserName(ctx)
			},
			want: "Test User",
		},
		{
			name: "验证Roles",
			getValue: func() (interface{}, error) {
				return jwtContext.GetRoles(ctx)
			},
			want: []string{"admin", "user"},
		},
		{
			name: "验证Permissions",
			getValue: func() (interface{}, error) {
				return jwtContext.GetPermissions(ctx)
			},
			want: []string{"read", "write"},
		},
		{
			name: "验证DeviceID",
			getValue: func() (interface{}, error) {
				return jwtContext.GetDeviceID(ctx)
			},
			want: "test-device",
		},
		{
			name: "验证IPAddress",
			getValue: func() (interface{}, error) {
				return jwtContext.GetIPAddress(ctx)
			},
			want: "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.getValue()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
