package validator

import (
	"github.com/gin-gonic/gin"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
)

// ClaimsValidator Claims验证器
type ClaimsValidator struct {
	// 是否验证过期时间
	validateExpiry bool
	// 是否验证Token类型
	validateTokenType bool
	// 允许的Token类型
	allowedTokenTypes []jwt.TokenType
}

// NewClaimsValidator 创建新的Claims验证器
func NewClaimsValidator(opts ...ClaimsValidatorOption) *ClaimsValidator {
	v := &ClaimsValidator{
		validateExpiry:    true,
		validateTokenType: true,
		allowedTokenTypes: []jwt.TokenType{jwt.AccessToken},
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// ClaimsValidatorOption 定义Claims验证器选项
type ClaimsValidatorOption func(*ClaimsValidator)

// WithExpiryValidation 设置是否验证过期时间
func WithExpiryValidation(validate bool) ClaimsValidatorOption {
	return func(v *ClaimsValidator) {
		v.validateExpiry = validate
	}
}

// WithTokenTypeValidation 设置是否验证Token类型
func WithTokenTypeValidation(validate bool) ClaimsValidatorOption {
	return func(v *ClaimsValidator) {
		v.validateTokenType = validate
	}
}

// WithAllowedTokenTypes 设置允许的Token类型
func WithAllowedTokenTypes(types ...jwt.TokenType) ClaimsValidatorOption {
	return func(v *ClaimsValidator) {
		v.allowedTokenTypes = types
	}
}

// Validate 实现TokenValidator接口
func (v *ClaimsValidator) Validate(c *gin.Context, claims jwt.Claims) error {
	// 验证基本Claims
	if err := claims.Validate(); err != nil {
		return errors.NewClaimsInvalidError("invalid claims", err)
	}

	// 验证Token类型
	if v.validateTokenType {
		tokenType := claims.GetTokenType()
		valid := false
		for _, allowed := range v.allowedTokenTypes {
			if tokenType == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return errors.NewTokenTypeMismatchError("invalid token type", nil)
		}
	}

	return nil
}

// ValidateExpiry 返回是否验证过期时间
func (v *ClaimsValidator) ValidateExpiry() bool {
	return v.validateExpiry
}

// ValidateTokenType 返回是否验证Token类型
func (v *ClaimsValidator) ValidateTokenType() bool {
	return v.validateTokenType
}

// GetAllowedTokenTypes 返回允许的Token类型列表
func (v *ClaimsValidator) GetAllowedTokenTypes() []jwt.TokenType {
	return v.allowedTokenTypes
}
