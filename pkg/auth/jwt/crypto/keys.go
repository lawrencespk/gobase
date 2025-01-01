package crypto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"sync"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
)

// KeyManager 密钥管理器
type KeyManager struct {
	method    jwt.SigningMethod
	algorithm Algorithm
	keyPair   *jwt.KeyPair
	mutex     sync.RWMutex
	logger    types.Logger
}

// NewKeyManager 创建密钥管理器
func NewKeyManager(method jwt.SigningMethod, logger types.Logger) (*KeyManager, error) {
	algorithm, err := createAlgorithm(method)
	if err != nil {
		return nil, err
	}

	return &KeyManager{
		method:    method,
		algorithm: algorithm,
		logger:    logger,
	}, nil
}

// InitializeKeys 初始化密钥
func (km *KeyManager) InitializeKeys(ctx context.Context, config *jwt.KeyConfig) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	switch km.method {
	case jwt.HS256, jwt.HS384, jwt.HS512:
		if config.SecretKey == "" {
			return errors.NewKeyInvalidError("secret key is required for HMAC", nil)
		}
		km.keyPair = &jwt.KeyPair{
			PrivateKey: []byte(config.SecretKey),
			PublicKey:  []byte(config.SecretKey),
		}

	case jwt.RS256, jwt.RS384, jwt.RS512:
		if config.PrivateKey == "" || config.PublicKey == "" {
			// 生成新的RSA密钥对
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return errors.NewKeyInvalidError("failed to generate RSA key pair", err)
			}

			km.keyPair = &jwt.KeyPair{
				PrivateKey: privateKey,
				PublicKey:  &privateKey.PublicKey,
			}

			// 记录日志
			km.logger.Info(ctx, "generated new RSA key pair")
		} else {
			// 解析已有的密钥对
			privateKey, publicKey, err := parseRSAKeyPair(config.PrivateKey, config.PublicKey)
			if err != nil {
				return err
			}

			km.keyPair = &jwt.KeyPair{
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			}
		}
	}

	return nil
}

// GetSigningKey 获取签名密钥
func (km *KeyManager) GetSigningKey() (interface{}, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	if km.keyPair == nil {
		return nil, errors.NewKeyInvalidError("key pair not initialized", nil)
	}

	return km.keyPair.PrivateKey, nil
}

// GetVerificationKey 获取验证密钥
func (km *KeyManager) GetVerificationKey() (interface{}, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	if km.keyPair == nil {
		return nil, errors.NewKeyInvalidError("key pair not initialized", nil)
	}

	return km.keyPair.PublicKey, nil
}

// RotateKeys 轮换密钥
func (km *KeyManager) RotateKeys(ctx context.Context) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	switch km.method {
	case jwt.RS256, jwt.RS384, jwt.RS512:
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return errors.NewKeyInvalidError("failed to generate RSA key pair", err)
		}

		km.keyPair = &jwt.KeyPair{
			PrivateKey: privateKey,
			PublicKey:  &privateKey.PublicKey,
		}

		km.logger.Info(ctx, "rotated RSA key pair")

	default:
		return errors.NewAlgorithmMismatchError("key rotation not supported for this method", nil)
	}

	return nil
}

// 辅助函数

func createAlgorithm(method jwt.SigningMethod) (Algorithm, error) {
	switch method {
	case jwt.HS256, jwt.HS384, jwt.HS512:
		return NewHMAC(method)
	case jwt.RS256, jwt.RS384, jwt.RS512:
		return NewRSA(method)
	default:
		return nil, errors.NewAlgorithmMismatchError("unsupported signing method", nil)
	}
}

func parseRSAKeyPair(privateKeyPEM, publicKeyPEM string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 解析私钥
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, nil, errors.NewKeyInvalidError("failed to parse private key PEM", nil)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, errors.NewKeyInvalidError("failed to parse private key", err)
	}

	// 解析公钥
	block, _ = pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, nil, errors.NewKeyInvalidError("failed to parse public key PEM", nil)
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, nil, errors.NewKeyInvalidError("failed to parse public key", err)
	}

	return privateKey, publicKey, nil
}
