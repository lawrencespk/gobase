package unit

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/middleware/jwt/validator"
)

// mockInvalidClaims 用于测试的无效Claims实现
type mockInvalidClaims struct {
	*jwt.StandardClaims
}

func newMockInvalidClaims() *mockInvalidClaims {
	return &mockInvalidClaims{
		StandardClaims: jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(-time.Hour)), // 已过期
		),
	}
}

func TestClaimsValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		validator   *validator.ClaimsValidator
		claims      jwt.Claims
		wantErr     bool
		errContains string
	}{
		{
			name:      "验证成功",
			validator: validator.NewClaimsValidator(),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"127.0.0.1",
			),
			wantErr: false,
		},
		{
			name:      "无效的Claims",
			validator: validator.NewClaimsValidator(),
			claims:    newMockInvalidClaims(),
			wantErr:   true,
		},
		{
			name: "禁用过期验证",
			validator: validator.NewClaimsValidator(
				validator.WithExpiryValidation(false),
			),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"127.0.0.1",
			),
			wantErr: false,
		},
		{
			name: "禁用Token类型验证",
			validator: validator.NewClaimsValidator(
				validator.WithTokenTypeValidation(false),
			),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"127.0.0.1",
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			err := tt.validator.Validate(c, tt.claims)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClaimsValidatorOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []validator.ClaimsValidatorOption
		check   func(*validator.ClaimsValidator) bool
	}{
		{
			name: "禁用过期验证",
			options: []validator.ClaimsValidatorOption{
				validator.WithExpiryValidation(false),
			},
			check: func(v *validator.ClaimsValidator) bool {
				return !v.ValidateExpiry()
			},
		},
		{
			name: "禁用Token类型验证",
			options: []validator.ClaimsValidatorOption{
				validator.WithTokenTypeValidation(false),
			},
			check: func(v *validator.ClaimsValidator) bool {
				return !v.ValidateTokenType()
			},
		},
		{
			name: "设置允许的Token类型",
			options: []validator.ClaimsValidatorOption{
				validator.WithAllowedTokenTypes(jwt.RefreshToken),
			},
			check: func(v *validator.ClaimsValidator) bool {
				types := v.GetAllowedTokenTypes()
				return len(types) == 1 && types[0] == jwt.RefreshToken
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.NewClaimsValidator(tt.options...)
			assert.True(t, tt.check(v))
		})
	}
}
