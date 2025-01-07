package context

import (
	"context"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
)

// GetClaims 从上下文中获取JWT Claims
func GetClaims(ctx context.Context) (jwt.Claims, error) {
	if claims, ok := ctx.Value(ClaimsKey).(jwt.Claims); ok {
		return claims, nil
	}
	return nil, errors.NewClaimsMissingError("claims not found in context", nil)
}

// GetToken 从上下文中获取JWT Token
func GetToken(ctx context.Context) (string, error) {
	if token, ok := ctx.Value(TokenKey).(string); ok {
		return token, nil
	}
	return "", errors.NewTokenNotFoundError("token not found in context", nil)
}

// GetTokenType 从上下文中获取Token类型
func GetTokenType(ctx context.Context) (jwt.TokenType, error) {
	if tokenType, ok := ctx.Value(TokenTypeKey).(jwt.TokenType); ok {
		return tokenType, nil
	}
	return "", errors.NewTokenTypeMismatchError("token type not found in context", nil)
}

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) (string, error) {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID, nil
	}
	return "", errors.NewClaimsMissingError("user id not found in context", nil)
}

// GetUserName 从上下文中获取用户名
func GetUserName(ctx context.Context) (string, error) {
	if userName, ok := ctx.Value(UserNameKey).(string); ok {
		return userName, nil
	}
	return "", errors.NewClaimsMissingError("user name not found in context", nil)
}

// GetRoles 从上下文中获取角色列表
func GetRoles(ctx context.Context) ([]string, error) {
	if roles, ok := ctx.Value(RolesKey).([]string); ok {
		return roles, nil
	}
	return nil, errors.NewClaimsMissingError("roles not found in context", nil)
}

// GetPermissions 从上下文中获取权限列表
func GetPermissions(ctx context.Context) ([]string, error) {
	if permissions, ok := ctx.Value(PermissionsKey).([]string); ok {
		return permissions, nil
	}
	return nil, errors.NewClaimsMissingError("permissions not found in context", nil)
}

// GetDeviceID 从上下文中获取设备ID
func GetDeviceID(ctx context.Context) (string, error) {
	if deviceID, ok := ctx.Value(DeviceIDKey).(string); ok {
		return deviceID, nil
	}
	return "", errors.NewClaimsMissingError("device id not found in context", nil)
}

// GetIPAddress 从上下文中获取IP地址
func GetIPAddress(ctx context.Context) (string, error) {
	if ipAddress, ok := ctx.Value(IPAddressKey).(string); ok {
		return ipAddress, nil
	}
	return "", errors.NewClaimsMissingError("ip address not found in context", nil)
}

// WithClaims 向上下文中添加JWT Claims
func WithClaims(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// WithToken 向上下文中添加JWT Token
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenKey, token)
}

// WithTokenType 向上下文中添加Token类型
func WithTokenType(ctx context.Context, tokenType jwt.TokenType) context.Context {
	return context.WithValue(ctx, TokenTypeKey, tokenType)
}

// WithUserID 向上下文中添加用户ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithUserName 向上下文中添加用户名
func WithUserName(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, UserNameKey, userName)
}

// WithRoles 向上下文中添加角色列表
func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesKey, roles)
}

// WithPermissions 向上下文中添加权限列表
func WithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, PermissionsKey, permissions)
}

// WithDeviceID 向上下文中添加设备ID
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, DeviceIDKey, deviceID)
}

// WithIPAddress 向上下文中添加IP地址
func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ipAddress)
}

// WithJWTContext 向上下文中添加所有JWT相关信息
func WithJWTContext(ctx context.Context, claims jwt.Claims, token string) context.Context {
	ctx = WithClaims(ctx, claims)
	ctx = WithToken(ctx, token)
	ctx = WithTokenType(ctx, claims.GetTokenType())
	ctx = WithUserID(ctx, claims.GetUserID())
	ctx = WithUserName(ctx, claims.GetUserName())
	ctx = WithRoles(ctx, claims.GetRoles())
	ctx = WithPermissions(ctx, claims.GetPermissions())
	ctx = WithDeviceID(ctx, claims.GetDeviceID())
	ctx = WithIPAddress(ctx, claims.GetIPAddress())
	return ctx
}
