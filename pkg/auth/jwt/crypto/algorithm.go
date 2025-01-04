package crypto

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
)

// HMAC 实现HMAC-SHA系列算法
type HMAC struct {
	hash   crypto.Hash
	method jwt.SigningMethod
}

// NewHMAC 创建HMAC算法实例
func NewHMAC(method jwt.SigningMethod) (Algorithm, error) {
	var hash crypto.Hash

	switch method {
	case jwt.HS256:
		hash = crypto.SHA256
	case jwt.HS384:
		hash = crypto.SHA384
	case jwt.HS512:
		hash = crypto.SHA512
	default:
		return nil, errors.NewAlgorithmMismatchError("unsupported HMAC method", nil)
	}

	return &HMAC{
		hash:   hash,
		method: method,
	}, nil
}

// Sign 使用HMAC算法签名
func (h *HMAC) Sign(data []byte, key interface{}) ([]byte, error) {
	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errors.NewKeyInvalidError("HMAC key must be []byte", nil)
	}

	hasher := hmac.New(h.hash.New, keyBytes)
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

// Verify 验证HMAC签名
func (h *HMAC) Verify(data []byte, signature []byte, key interface{}) error {
	expected, err := h.Sign(data, key)
	if err != nil {
		return err
	}

	if !hmac.Equal(signature, expected) {
		return errors.NewSignatureInvalidError("HMAC signature verification failed", nil)
	}

	return nil
}

// Name 获取算法名称
func (h *HMAC) Name() jwt.SigningMethod {
	return h.method
}

// RSA 实现RSA-SHA系列算法
type RSA struct {
	hash   crypto.Hash
	method jwt.SigningMethod
}

// NewRSA 创建RSA算法实例
func NewRSA(method jwt.SigningMethod) (Algorithm, error) {
	var hash crypto.Hash

	switch method {
	case jwt.RS256:
		hash = crypto.SHA256
	case jwt.RS384:
		hash = crypto.SHA384
	case jwt.RS512:
		hash = crypto.SHA512
	default:
		return nil, errors.NewAlgorithmMismatchError("unsupported RSA method", nil)
	}

	return &RSA{
		hash:   hash,
		method: method,
	}, nil
}

// Sign 使用RSA算法签名
func (r *RSA) Sign(data []byte, key interface{}) ([]byte, error) {
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.NewKeyInvalidError("RSA key must be *rsa.PrivateKey", nil)
	}

	hasher := r.hash.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(nil, privateKey, r.hash, hashed)
	if err != nil {
		return nil, errors.NewSignatureInvalidError("RSA signing failed", err)
	}

	return signature, nil
}

// Verify 验证RSA签名
func (r *RSA) Verify(data []byte, signature []byte, key interface{}) error {
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.NewKeyInvalidError("RSA key must be *rsa.PublicKey", nil)
	}

	hasher := r.hash.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	err := rsa.VerifyPKCS1v15(publicKey, r.hash, hashed, signature)
	if err != nil {
		return errors.NewSignatureInvalidError("RSA signature verification failed", err)
	}

	return nil
}

// Name 获取算法名称
func (r *RSA) Name() jwt.SigningMethod {
	return r.method
}

// CreateAlgorithm 根据签名方法创建相应的算法实例
func CreateAlgorithm(method jwt.SigningMethod) (Algorithm, error) {
	switch method {
	case jwt.HS256, jwt.HS384, jwt.HS512:
		return NewHMAC(method)
	case jwt.RS256, jwt.RS384, jwt.RS512:
		return NewRSA(method)
	default:
		return nil, errors.NewAlgorithmMismatchError("unsupported signing method", nil)
	}
}
