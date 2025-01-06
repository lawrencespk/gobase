package jwt

import (
	"encoding/json"
	"gobase/pkg/errors"
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
	Raw string `json:"raw"`
	// Token类型
	Type TokenType `json:"type"`
	// Token的Claims
	Claims Claims `json:"-"` // 不直接序列化接口
	// 过期时间
	ExpiresAt time.Time `json:"expires_at"`
	// 是否已被吊销
	IsRevoked bool `json:"is_revoked"`
}

// MarshalJSON 自定义JSON序列化方法
func (t *TokenInfo) MarshalJSON() ([]byte, error) {
	type Alias TokenInfo
	aux := &struct {
		*Alias
		Claims map[string]interface{} `json:"claims"`
	}{
		Alias: (*Alias)(t),
	}

	// 将 Claims 序列化为 map
	if t.Claims != nil {
		claimsData, err := json.Marshal(t.Claims)
		if err != nil {
			return nil, errors.NewSerializationError("failed to marshal claims", err)
		}

		var claimsMap map[string]interface{}
		if err := json.Unmarshal(claimsData, &claimsMap); err != nil {
			return nil, errors.NewSerializationError("failed to unmarshal claims to map", err)
		}
		aux.Claims = claimsMap
	}

	return json.Marshal(aux)
}

// UnmarshalJSON 自定义JSON反序列化方法
func (t *TokenInfo) UnmarshalJSON(data []byte) error {
	type Alias TokenInfo
	aux := &struct {
		*Alias
		Claims map[string]interface{} `json:"claims"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return errors.NewSerializationError("failed to unmarshal token info", err)
	}

	// 将 map 转换回 Claims
	if aux.Claims != nil {
		claimsData, err := json.Marshal(aux.Claims)
		if err != nil {
			return errors.NewSerializationError("failed to marshal claims map", err)
		}

		// 先尝试解析为 StandardClaims
		var stdClaims StandardClaims
		if err := json.Unmarshal(claimsData, &stdClaims); err == nil {
			t.Claims = &stdClaims
			return nil
		}

		// 如果不是 StandardClaims，保持原始数据
		// 具体类型的解析由调用者处理
		t.Claims = nil
	}

	return nil
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
