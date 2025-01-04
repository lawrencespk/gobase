package stress

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
)

func TestConcurrentSignAndVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	alg, err := crypto.NewHMAC(jwt.HS256)
	require.NoError(t, err)

	key := []byte("test-secret")
	data := []byte("test data")

	// 并发签名和验证
	workers := 100
	iterations := 1000
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sig, err := alg.Sign(data, key)
				require.NoError(t, err)
				err = alg.Verify(data, sig, key)
				require.NoError(t, err)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Completed %d operations in %v", workers*iterations, duration)
	t.Logf("Average operation time: %v", duration/time.Duration(workers*iterations))
}
