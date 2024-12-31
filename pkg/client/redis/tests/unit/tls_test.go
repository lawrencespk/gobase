package unit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSConfig(t *testing.T) {
	// 设置测试证书路径
	certDir := filepath.Join("testdata", "tls")
	certPath := filepath.Join(certDir, "client.crt")
	keyPath := filepath.Join(certDir, "client.key")

	// 确保测试目录存在
	err := os.MkdirAll(certDir, 0755)
	require.NoError(t, err)

	// 生成测试证书
	err = generateTestCertificates(certPath, keyPath)
	require.NoError(t, err)

	// 清理测试文件
	defer func() {
		os.RemoveAll(certDir)
	}()

	t.Run("valid tls config", func(t *testing.T) {
		// 禁用实际连接验证
		redis.DisableConnectionCheck = true
		defer func() {
			redis.DisableConnectionCheck = false
		}()

		client, err := redis.NewClient(
			redis.WithAddress("localhost:6379"),
			redis.WithTLS(true),
			redis.WithTLSCert(certPath),
			redis.WithTLSKey(keyPath),
			redis.WithDialTimeout(time.Second),
			redis.WithConnTimeout(time.Second),
			redis.WithMaxRetries(1),
			redis.WithSkipVerify(true),
		)

		assert.NoError(t, err)
		if client != nil {
			client.Close()
		}
	})

	t.Run("invalid cert path", func(t *testing.T) {
		client, err := redis.NewClient(
			redis.WithAddress("localhost:6379"),
			redis.WithTLS(true),
			redis.WithTLSCert("invalid/path/cert"),
			redis.WithTLSKey("invalid/path/key"),
		)
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.CacheError))
		assert.Nil(t, client)
	})

	t.Run("tls config without cert", func(t *testing.T) {
		// 禁用实际连接验证
		redis.DisableConnectionCheck = true
		defer func() {
			redis.DisableConnectionCheck = false
		}()

		client, err := redis.NewClient(
			redis.WithAddress("localhost:6379"),
			redis.WithTLS(true),
			redis.WithSkipVerify(true),
			redis.WithDialTimeout(time.Second),
			redis.WithConnTimeout(time.Second),
		)
		assert.NoError(t, err)
		if client != nil {
			client.Close()
		}
	})

	t.Run("tls disabled", func(t *testing.T) {
		// 禁用实际连接验证
		redis.DisableConnectionCheck = true
		defer func() {
			redis.DisableConnectionCheck = false
		}()

		client, err := redis.NewClient(
			redis.WithAddress("localhost:6379"),
			redis.WithTLS(false),
			redis.WithDialTimeout(time.Second),
			redis.WithConnTimeout(time.Second),
		)
		assert.NoError(t, err)
		if client != nil {
			client.Close()
		}
	})
}

// generateTestCertificates 生成测试证书
func generateTestCertificates(certPath, keyPath string) error {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 创建自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// 保存证书
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}

	// 保存私钥
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	err = pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	return nil
}
