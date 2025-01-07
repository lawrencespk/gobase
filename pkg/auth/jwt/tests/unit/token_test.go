package jwt_test

import (
	"context"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors/codes"
	errortypes "gobase/pkg/errors/types"
)

func TestTokenManager_GenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*jwt.TokenManager, jwt.Claims)
		wantErr bool
	}{
		{
			name: "使用HMAC生成有效token",
			setup: func(t *testing.T) (*jwt.TokenManager, jwt.Claims) {
				tm, err := jwt.NewTokenManager("test-secret-key", jwt.WithoutTracing())
				require.NoError(t, err, "创建TokenManager失败")

				claims := &jwt.StandardClaims{}
				jwt.WithUserID("test-user")(claims)
				jwt.WithUserName("Test User")(claims)
				jwt.WithTokenType(jwt.AccessToken)(claims)
				jwt.WithExpiresAt(time.Now().Add(time.Hour))(claims)
				return tm, claims
			},
			wantErr: false,
		},
		{
			name: "claims缺少必要字段",
			setup: func(t *testing.T) (*jwt.TokenManager, jwt.Claims) {
				tm, err := jwt.NewTokenManager("test-secret-key", jwt.WithoutTracing())
				require.NoError(t, err, "创建TokenManager失败")

				claims := &jwt.StandardClaims{}
				// 故意不设置必要字段
				return tm, claims
			},
			wantErr: true,
		},
		{
			name: "token已过期",
			setup: func(t *testing.T) (*jwt.TokenManager, jwt.Claims) {
				tm, err := jwt.NewTokenManager("test-secret-key", jwt.WithoutTracing())
				require.NoError(t, err, "创建TokenManager失败")

				claims := &jwt.StandardClaims{}
				jwt.WithUserID("test-user")(claims)
				jwt.WithTokenType(jwt.AccessToken)(claims)
				jwt.WithExpiresAt(time.Now().Add(-time.Hour))(claims)
				return tm, claims
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, claims := tt.setup(t)
			ctx := context.Background()

			token, err := tm.GenerateToken(ctx, claims)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestTokenManager_ValidateToken(t *testing.T) {
	tm, err := jwt.NewTokenManager("test-secret-key", jwt.WithoutTracing())
	require.NoError(t, err)

	tests := []struct {
		name       string
		setupToken func() string
		wantErr    bool
	}{
		{
			name: "验证有效token",
			setupToken: func() string {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)
				token, _ := tm.GenerateToken(context.Background(), claims)
				return token
			},
			wantErr: false,
		},
		{
			name: "验证无效token",
			setupToken: func() string {
				return "invalid.token.string"
			},
			wantErr: true,
		},
		{
			name: "验证过期token",
			setupToken: func() string {
				claims := &jwt.StandardClaims{}
				jwt.WithUserID("test-user")(claims)
				jwt.WithTokenType(jwt.AccessToken)(claims)
				jwt.WithExpiresAt(time.Now().Add(-time.Hour))(claims)
				token, _ := tm.GenerateToken(context.Background(), claims)
				return token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			token := tt.setupToken()

			parsedToken, err := tm.ValidateToken(ctx, token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, parsedToken)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, parsedToken)

			// 验证解析出的claims
			claims, ok := parsedToken.Claims.(*jwt.StandardClaims)
			assert.True(t, ok, "claims should be of type *jwt.StandardClaims")
			if ok {
				assert.Equal(t, "test-user", claims.GetUserID())
				assert.Equal(t, jwt.AccessToken, claims.GetTokenType())
			}
		})
	}
}

func TestTokenManager_HandleValidationError(t *testing.T) {
	tm, err := jwt.NewTokenManager("test-secret", jwt.WithoutTracing())
	require.NoError(t, err)

	tests := []struct {
		name       string
		inputError error
		wantCode   string
	}{
		{
			name:       "处理token过期错误",
			inputError: jwtlib.ErrTokenExpired,
			wantCode:   codes.TokenExpired,
		},
		{
			name:       "处理签名无效错误",
			inputError: jwtlib.ErrSignatureInvalid,
			wantCode:   codes.SignatureInvalid,
		},
		{
			name:       "处理其他验证错误",
			inputError: jwtlib.ErrTokenUnverifiable,
			wantCode:   codes.TokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.HandleValidationError(tt.inputError)
			assert.NotNil(t, err)

			// 验证错误码
			customErr, ok := err.(errortypes.Error)
			if ok {
				assert.Equal(t, tt.wantCode, customErr.Code())
			} else {
				t.Errorf("expected error to implement errors.types.Error interface")
			}
		})
	}
}

func TestNewToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*jwt.Token, error)
		wantErr bool
		errCode string
	}{
		{
			name: "成功创建token",
			setup: func() (*jwt.Token, error) {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)
				return jwt.NewToken(
					jwt.WithClaims(claims),
					jwt.WithSecret("test-secret"),
					jwt.WithExpiration(time.Hour),
				)
			},
			wantErr: false,
		},
		{
			name: "缺少claims",
			setup: func() (*jwt.Token, error) {
				return jwt.NewToken(
					jwt.WithSecret("test-secret"),
					jwt.WithExpiration(time.Hour),
				)
			},
			wantErr: true,
			errCode: codes.TokenGenerationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.setup()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					customErr, ok := err.(errortypes.Error)
					assert.True(t, ok, "error should implement Error interface")
					assert.Equal(t, tt.errCode, customErr.Code())
				}
				assert.Nil(t, token)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, token)
		})
	}
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (string, error)
		wantErr   bool
		errCode   string
		checkFunc func(*testing.T, *jwt.Token)
	}{
		{
			name: "成功解析token",
			setup: func() (string, error) {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
				)
				token, err := jwt.NewToken(
					jwt.WithClaims(claims),
					jwt.WithSecret("test-secret"),
				)
				if err != nil {
					return "", err
				}
				return token.SignedString()
			},
			wantErr: false,
			checkFunc: func(t *testing.T, token *jwt.Token) {
				claims := token.Claims()
				assert.NotNil(t, claims)
				assert.Equal(t, "test-user", claims.GetUserID())
				assert.Equal(t, jwt.AccessToken, claims.GetTokenType())
			},
		},
		{
			name: "无效的token字符串",
			setup: func() (string, error) {
				return "invalid.token.string", nil
			},
			wantErr: true,
			errCode: codes.TokenInvalid,
		},
		{
			name: "缺少secret",
			setup: func() (string, error) {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
				)
				token, err := jwt.NewToken(
					jwt.WithClaims(claims),
					jwt.WithSecret("temp-secret"),
				)
				if err != nil {
					return "", err
				}
				return token.SignedString()
			},
			wantErr: true,
			errCode: codes.TokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := tt.setup()
			require.NoError(t, err)

			token, err := jwt.ParseToken(tokenString, jwt.WithSecret("test-secret"))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					customErr, ok := err.(errortypes.Error)
					assert.True(t, ok, "error should implement Error interface")
					assert.Equal(t, tt.errCode, customErr.Code())
				}
				assert.Nil(t, token)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, token)
			if tt.checkFunc != nil {
				tt.checkFunc(t, token)
			}
		})
	}
}

func TestToken_SignedString(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *jwt.Token
		wantErr bool
		errCode string
	}{
		{
			name: "成功签名token",
			setup: func() *jwt.Token {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
				)
				token, _ := jwt.NewToken(
					jwt.WithClaims(claims),
					jwt.WithSecret("test-secret"),
				)
				return token
			},
			wantErr: false,
		},
		{
			name: "缺少secret",
			setup: func() *jwt.Token {
				claims := jwt.NewStandardClaims(
					jwt.WithUserID("test-user"),
					jwt.WithTokenType(jwt.AccessToken),
				)
				token, _ := jwt.NewToken(jwt.WithClaims(claims))
				return token
			},
			wantErr: true,
			errCode: codes.TokenSignFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			signed, err := token.SignedString()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					customErr, ok := err.(errortypes.Error)
					assert.True(t, ok, "error should implement Error interface")
					assert.Equal(t, tt.errCode, customErr.Code())
				}
				assert.Empty(t, signed)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, signed)
		})
	}
}
