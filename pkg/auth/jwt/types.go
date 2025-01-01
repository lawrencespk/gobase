package jwt

import (
	"time"
)

// TokenType 定义token类型
type TokenType string

const (
	// AccessToken 访问令牌
	AccessToken TokenType = "access"
	// RefreshToken 刷新令牌
	RefreshToken TokenType = "refresh"
)

// TokenInfo 存储Token的基本信息
type TokenInfo struct {
	// Token的原始字符串
	Raw string
	// Token类型
	Type TokenType
	// Token的Claims
	Claims Claims
	// 过期时间
	ExpiresAt time.Time
	// 是否已被吊销
	IsRevoked bool
}

// TokenPair 包含访问令牌和刷新令牌
type TokenPair struct {
	// 访问令牌
	AccessToken string
	// 刷新令牌
	RefreshToken string
	// 访问令牌过期时间
	AccessExpiresAt time.Time
	// 刷新令牌过期时间
	RefreshExpiresAt time.Time
}

// SigningMethod 定义签名方法类型
type SigningMethod string

const (
	// HS256 HMAC SHA256
	HS256 SigningMethod = "HS256"
	// HS384 HMAC SHA384
	HS384 SigningMethod = "HS384"
	// HS512 HMAC SHA512
	HS512 SigningMethod = "HS512"
	// RS256 RSA SHA256
	RS256 SigningMethod = "RS256"
	// RS384 RSA SHA384
	RS384 SigningMethod = "RS384"
	// RS512 RSA SHA512
	RS512 SigningMethod = "RS512"
)

// KeyPair 存储密钥对
type KeyPair struct {
	// 私钥
	PrivateKey interface{}
	// 公钥
	PublicKey interface{}
}

// TokenStatus 定义Token状态
type TokenStatus string

const (
	// TokenStatusValid Token有效
	TokenStatusValid TokenStatus = "valid"
	// TokenStatusExpired Token过期
	TokenStatusExpired TokenStatus = "expired"
	// TokenStatusRevoked Token已吊销
	TokenStatusRevoked TokenStatus = "revoked"
	// TokenStatusInvalid Token无效
	TokenStatusInvalid TokenStatus = "invalid"
)

// KeyConfig JWT密钥配置
type KeyConfig struct {
	SecretKey  string // HMAC密钥
	PrivateKey string // RSA私钥
	PublicKey  string // RSA公钥
}
