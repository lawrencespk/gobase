package security

import (
	"context"
	"net"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
)

// TokenValidator Token验证器
type TokenValidator struct {
	policy *Policy
}

// NewTokenValidator 创建Token验证器
func NewTokenValidator(policy *Policy) *TokenValidator {
	return &TokenValidator{
		policy: policy,
	}
}

// ValidateToken 验证Token
func (v *TokenValidator) ValidateToken(ctx context.Context, token *jwt.TokenInfo) error {
	// 验证过期时间
	if time.Now().After(token.ExpiresAt) {
		return errors.NewTokenExpiredError("token has expired", nil)
	}

	// 验证Token类型
	if err := v.validateTokenType(token.Type); err != nil {
		return err
	}

	// 验证Claims
	if err := v.validateClaims(token.Claims); err != nil {
		return err
	}

	return nil
}

// validateTokenType 验证Token类型
func (v *TokenValidator) validateTokenType(tokenType jwt.TokenType) error {
	switch tokenType {
	case jwt.AccessToken, jwt.RefreshToken:
		return nil
	default:
		return errors.NewTokenTypeMismatchError("invalid token type", nil)
	}
}

// validateClaims 验证Claims
func (v *TokenValidator) validateClaims(claims jwt.Claims) error {
	// 验证必要字段
	if claims.GetUserID() == "" {
		return errors.NewClaimsMissingError("user ID is required", nil)
	}

	// 验证IP绑定
	if v.policy.EnableIPBinding {
		if claims.GetIPAddress() == "" {
			return errors.NewBindingInvalidError("IP address is required", nil)
		}
		if err := v.validateIPAddress(claims.GetIPAddress()); err != nil {
			return err
		}
	}

	// 验证设备绑定
	if v.policy.EnableDeviceBinding && claims.GetDeviceID() == "" {
		return errors.NewBindingInvalidError("device ID is required", nil)
	}

	return nil
}

// validateIPAddress 验证IP地址
func (v *TokenValidator) validateIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		return errors.NewBindingInvalidError("invalid IP address", nil)
	}
	return nil
}
