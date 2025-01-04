package benchmark

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
)

// generateTestRSAKey 生成用于测试的RSA密钥对
func generateTestRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func BenchmarkHMAC_Sign(b *testing.B) {
	alg, _ := crypto.NewHMAC(jwt.HS256)
	key := []byte("test-secret")
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = alg.Sign(data, key)
	}
}

func BenchmarkHMAC_Verify(b *testing.B) {
	alg, _ := crypto.NewHMAC(jwt.HS256)
	key := []byte("test-secret")
	data := []byte("test data")
	sig, _ := alg.Sign(data, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = alg.Verify(data, sig, key)
	}
}

func BenchmarkRSA_Sign(b *testing.B) {
	alg, _ := crypto.NewRSA(jwt.RS256)
	key, _ := generateTestRSAKey()
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = alg.Sign(data, key)
	}
}

func BenchmarkRSA_Verify(b *testing.B) {
	alg, _ := crypto.NewRSA(jwt.RS256)
	key, _ := generateTestRSAKey()
	data := []byte("test data")
	sig, _ := alg.Sign(data, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = alg.Verify(data, sig, &key.PublicKey)
	}
}
