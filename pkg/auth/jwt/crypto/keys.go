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

	switch km.algorithm.(type) {
	case *HMAC:
		if config == nil || config.SecretKey == "" {
			return errors.NewKeyInvalidError("secret key is required for HMAC", nil)
		}
		km.keyPair = &jwt.KeyPair{
			PrivateKey: []byte(config.SecretKey),
			PublicKey:  []byte(config.SecretKey),
		}
		return nil

	case *RSA:
		var privateKey *rsa.PrivateKey
		var publicKey *rsa.PublicKey
		var err error

		if config != nil && config.PrivateKey != "" {
			// 使用提供的密钥
			privateKey, publicKey, err = parseRSAKeyPair(ctx, km.logger, config.PrivateKey)
			if err != nil {
				return err
			}
		} else {
			// 自动生成密钥对
			privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return errors.NewKeyInvalidError("failed to generate RSA key pair", err)
			}
			publicKey = &privateKey.PublicKey
		}

		km.keyPair = &jwt.KeyPair{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		}
		return nil

	default:
		return errors.NewKeyInvalidError("unsupported algorithm", nil)
	}
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

// RotateKeys 轮转密钥
func (km *KeyManager) RotateKeys(ctx context.Context) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	// 先检查方法类型
	switch km.method {
	case jwt.HS256, jwt.HS384, jwt.HS512:
		return errors.NewKeyInvalidError("key rotation not supported for HMAC", nil)
	case jwt.RS256, jwt.RS384, jwt.RS512:
		// 检查密钥对是否已初始化
		if km.keyPair == nil {
			return errors.NewKeyInvalidError("key pair not initialized", nil)
		}
		// 生成新的 RSA 密钥对
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return errors.NewKeyInvalidError("failed to generate RSA key pair", err)
		}
		km.keyPair = &jwt.KeyPair{
			PrivateKey: privateKey,
			PublicKey:  &privateKey.PublicKey,
		}
		return nil
	default:
		return errors.NewAlgorithmMismatchError("unsupported signing method", nil)
	}
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

func parseRSAKeyPair(ctx context.Context, logger types.Logger, privateKeyPEM string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 1. 验证 PEM 内容是否为空
	if len(privateKeyPEM) == 0 {
		logger.Error(ctx, "empty PEM content")
		return nil, nil, errors.NewKeyInvalidError("empty PEM content", nil)
	}

	// 2. 解码 PEM
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		logger.Error(ctx, "failed to decode PEM block")
		return nil, nil, errors.NewKeyInvalidError("failed to decode PEM block", nil)
	}

	logger.Debug(ctx, "decoded PEM block",
		types.Field{Key: "type", Value: block.Type},
		types.Field{Key: "bytes_length", Value: len(block.Bytes)})

	// 3. 解析私钥
	var privateKey *rsa.PrivateKey
	var err error

	if block.Type == "RSA PRIVATE KEY" {
		// 处理 PKCS1 格式
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			logger.Error(ctx, "failed to parse PKCS1 private key",
				types.Field{Key: "error", Value: err.Error()})
			return nil, nil, errors.NewKeyInvalidError("failed to parse private key", err)
		}
	} else if block.Type == "PRIVATE KEY" {
		// 处理 PKCS8 格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			logger.Error(ctx, "failed to parse PKCS8 private key",
				types.Field{Key: "error", Value: err.Error()})
			return nil, nil, errors.NewKeyInvalidError("failed to parse private key", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			logger.Error(ctx, "parsed key is not an RSA private key")
			return nil, nil, errors.NewKeyInvalidError("parsed key is not an RSA private key", nil)
		}
	} else {
		logger.Error(ctx, "unsupported private key type",
			types.Field{Key: "type", Value: block.Type})
		return nil, nil, errors.NewKeyInvalidError("unsupported private key type", nil)
	}

	// 4. 验证密钥
	if err := privateKey.Validate(); err != nil {
		logger.Error(ctx, "invalid RSA private key",
			types.Field{Key: "error", Value: err.Error()})
		return nil, nil, errors.NewKeyInvalidError("invalid RSA private key", err)
	}

	logger.Debug(ctx, "successfully parsed RSA key pair",
		types.Field{Key: "modulus_size", Value: privateKey.N.BitLen()})

	return privateKey, &privateKey.PublicKey, nil
}
