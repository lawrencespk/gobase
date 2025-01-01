package crypto

import (
	"gobase/pkg/auth/jwt"
)

// KeyProvider 密钥提供者接口
type KeyProvider interface {
	// GetSigningKey 获取签名密钥
	GetSigningKey() (interface{}, error)
	// GetVerificationKey 获取验证密钥
	GetVerificationKey() (interface{}, error)
	// RotateKeys 轮换密钥
	RotateKeys() error
}

// Algorithm 加密算法接口
type Algorithm interface {
	// Sign 签名
	Sign(data []byte, key interface{}) ([]byte, error)
	// Verify 验证签名
	Verify(data []byte, signature []byte, key interface{}) error
	// Name 获取算法名称
	Name() jwt.SigningMethod
}
