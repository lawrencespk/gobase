package validator

import (
	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// BlacklistValidator 黑名单验证器
type BlacklistValidator struct {
	// 黑名单接口
	blacklist blacklist.TokenBlacklist
}

// NewBlacklistValidator 创建新的黑名单验证器
func NewBlacklistValidator(blacklist blacklist.TokenBlacklist) *BlacklistValidator {
	return &BlacklistValidator{
		blacklist: blacklist,
	}
}

// Validate 实现TokenValidator接口
func (v *BlacklistValidator) Validate(c *gin.Context, claims jwt.Claims) error {
	// 从上下文获取token
	token, exists := c.Get("jwt_token")
	if !exists {
		return errors.NewTokenNotFoundError("token not found in context", nil)
	}

	// 使用 gin.Context 而不是 Request.Context()
	isBlacklisted, err := v.blacklist.IsBlacklisted(c, token.(string))
	if err != nil {
		return errors.NewTokenBlacklistError("failed to check token blacklist", err)
	}

	if isBlacklisted {
		return errors.NewTokenRevokedError("token has been revoked", nil)
	}

	return nil
}
