package testutils

import (
	"crypto/rand"
	"crypto/rsa"
)

// GenerateTestRSAKey 生成用于测试的RSA密钥对
func GenerateTestRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}
