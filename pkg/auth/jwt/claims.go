package jwt

import (
	"time"

	"gobase/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
)

// 错误变量定义
var (
	ErrClaimsMissing = errors.NewClaimsMissingError("claims missing", nil)
	ErrClaimsExpired = errors.NewClaimsExpiredError("claims expired", nil)
	ErrClaimsInvalid = errors.NewClaimsInvalidError("claims invalid", nil)
)

// Claims 定义JWT的Claims接口
type Claims interface {
	// 继承标准JWT Claims接口
	jwt.Claims
	// 获取用户ID
	GetUserID() string
	// 获取用户名
	GetUserName() string
	// 获取角色列表
	GetRoles() []string
	// 获取权限列表
	GetPermissions() []string
	// 获取设备ID
	GetDeviceID() string
	// 获取IP地址
	GetIPAddress() string
	// 获取Token类型
	GetTokenType() TokenType
	// 获取Token ID
	GetTokenID() string
	// 验证Claims
	Validate() error
	// 设置过期时间
	SetExpiresAt(time.Time)
	// 获取过期时间
	GetExpiresAt() time.Time
}

// StandardClaims 标准Claims实现
type StandardClaims struct {
	jwt.RegisteredClaims
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	DeviceID    string    `json:"device_id"`
	IPAddress   string    `json:"ip_address"`
	TokenType   TokenType `json:"token_type"`
	TokenID     string    `json:"token_id"`
}

// NewStandardClaims 创建标准Claims
func NewStandardClaims(options ...ClaimsOption) *StandardClaims {
	claims := &StandardClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gobase",
			Subject:   "token",
			Audience:  []string{"gobase"},
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 应用选项
	for _, opt := range options {
		opt(claims)
	}

	return claims
}

// GetUserID 获取用户ID
func (c *StandardClaims) GetUserID() string {
	return c.UserID
}

// GetUserName 获取用户名
func (c *StandardClaims) GetUserName() string {
	return c.UserName
}

// GetRoles 获取角色列表
func (c *StandardClaims) GetRoles() []string {
	return c.Roles
}

// GetPermissions 获取权限列表
func (c *StandardClaims) GetPermissions() []string {
	return c.Permissions
}

// GetDeviceID 获取设备ID
func (c *StandardClaims) GetDeviceID() string {
	return c.DeviceID
}

// GetIPAddress 获取IP地址
func (c *StandardClaims) GetIPAddress() string {
	return c.IPAddress
}

// GetTokenType 获取Token类型
func (c *StandardClaims) GetTokenType() TokenType {
	return c.TokenType
}

// GetTokenID 获取Token ID
func (c *StandardClaims) GetTokenID() string {
	return c.TokenID
}

// Validate 验证Claims
func (c *StandardClaims) Validate() error {
	// 验证必填字段
	if c.UserID == "" {
		return errors.NewClaimsMissingError("user_id is required", nil)
	}

	// 验证过期时间
	if c.ExpiresAt != nil && time.Now().After(c.ExpiresAt.Time) {
		return ErrClaimsExpired
	}

	// 验证Token类型
	if c.TokenType == "" {
		return errors.NewClaimsInvalidError("token_type is required", nil)
	}

	return nil
}

// SetExpiresAt 实现Claims接口的SetExpiresAt方法
func (c *StandardClaims) SetExpiresAt(t time.Time) {
	c.ExpiresAt = jwt.NewNumericDate(t)
}

// GetExpiresAt 获取过期时间
func (c *StandardClaims) GetExpiresAt() time.Time {
	if c.ExpiresAt == nil {
		return time.Time{}
	}
	return c.ExpiresAt.Time
}

// ClaimsOption Claims配置选项
type ClaimsOption func(*StandardClaims)

// WithUserID 设置用户ID
func WithUserID(userID string) ClaimsOption {
	return func(c *StandardClaims) {
		c.UserID = userID
	}
}

// WithUserName 设置用户名
func WithUserName(userName string) ClaimsOption {
	return func(c *StandardClaims) {
		c.UserName = userName
	}
}

// WithRoles 设置角色列表
func WithRoles(roles []string) ClaimsOption {
	return func(c *StandardClaims) {
		c.Roles = roles
	}
}

// WithPermissions 设置权限列表
func WithPermissions(permissions []string) ClaimsOption {
	return func(c *StandardClaims) {
		c.Permissions = permissions
	}
}

// WithDeviceID 设置设备ID
func WithDeviceID(deviceID string) ClaimsOption {
	return func(c *StandardClaims) {
		c.DeviceID = deviceID
	}
}

// WithIPAddress 设置IP地址
func WithIPAddress(ip string) ClaimsOption {
	return func(c *StandardClaims) {
		c.IPAddress = ip
	}
}

// WithTokenType 设置Token类型
func WithTokenType(tokenType TokenType) ClaimsOption {
	return func(c *StandardClaims) {
		c.TokenType = tokenType
	}
}

// WithTokenID 设置Token ID
func WithTokenID(tokenID string) ClaimsOption {
	return func(c *StandardClaims) {
		c.TokenID = tokenID
	}
}

// WithExpiresAt 设置过期时间
func WithExpiresAt(expiresAt time.Time) ClaimsOption {
	return func(c *StandardClaims) {
		c.ExpiresAt = jwt.NewNumericDate(expiresAt)
	}
}

// WithNotBefore 设置生效时间
func WithNotBefore(notBefore time.Time) ClaimsOption {
	return func(c *StandardClaims) {
		c.NotBefore = jwt.NewNumericDate(notBefore)
	}
}

// WithIssuer 设置签发者
func WithIssuer(issuer string) ClaimsOption {
	return func(c *StandardClaims) {
		c.Issuer = issuer
	}
}

// WithSubject 设置主题
func WithSubject(subject string) ClaimsOption {
	return func(c *StandardClaims) {
		c.Subject = subject
	}
}

// WithAudience 设置受众
func WithAudience(audience []string) ClaimsOption {
	return func(c *StandardClaims) {
		c.Audience = audience
	}
}
