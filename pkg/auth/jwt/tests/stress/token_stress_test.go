package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func TestTokenManager_StressConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 初始化
	log, err := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	require.NoError(t, err)

	tm, err := jwt.NewTokenManager("test-secret",
		jwt.WithLogger(log),
		jwt.WithoutMetrics(),
		jwt.WithoutTracing(),
	)
	require.NoError(t, err)

	// 测试参数
	const (
		numGoroutines = 100
		numOperations = 1000
		timeout       = 30 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	// 启动并发goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				default:
					// 生成token
					claims := jwt.NewStandardClaims(
						jwt.WithUserID("test-user"),
						jwt.WithTokenType(jwt.AccessToken),
						jwt.WithExpiresAt(time.Now().Add(time.Hour)),
					)
					token, err := tm.GenerateToken(ctx, claims)
					if err != nil {
						errCh <- err
						return
					}

					// 验证token
					_, err = tm.ValidateToken(ctx, token)
					if err != nil {
						errCh <- err
						return
					}
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errCh)

	// 检查错误
	for err := range errCh {
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestTokenManager_StressMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 初始化
	log, err := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	require.NoError(t, err)

	tm, err := jwt.NewTokenManager("test-secret",
		jwt.WithLogger(log),
		jwt.WithoutMetrics(),
		jwt.WithoutTracing(),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// 生成大量token并保持在内存中
	tokens := make([]string, 0, 10000)
	for i := 0; i < 10000; i++ {
		claims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(time.Hour)),
			jwt.WithRoles([]string{"user", "admin"}),
			jwt.WithPermissions([]string{"read", "write"}),
		)
		token, err := tm.GenerateToken(ctx, claims)
		require.NoError(t, err)
		tokens = append(tokens, token)
	}

	// 并发验证所有token
	var wg sync.WaitGroup
	for _, tokenStr := range tokens {
		wg.Add(1)
		go func(tokenStr string) {
			defer wg.Done()
			_, err := tm.ValidateToken(ctx, tokenStr)
			if err != nil {
				t.Errorf("token validation failed: %v", err)
			}
		}(tokenStr)
	}

	wg.Wait()
}

// TestTokenManager_StressClaimsValidation 测试大量Claims验证的压力
func TestTokenManager_StressClaimsValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	log, err := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	require.NoError(t, err)

	tm, err := jwt.NewTokenManager("test-secret",
		jwt.WithLogger(log),
		jwt.WithoutMetrics(),
		jwt.WithoutTracing(),
	)
	require.NoError(t, err)

	const (
		numGoroutines = 50
		numClaims     = 1000
	)

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	// 并发生成和验证不同类型的Claims
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numClaims; j++ {
				// 随机生成不同类型的Claims
				claims := jwt.NewStandardClaims(
					jwt.WithUserID(fmt.Sprintf("user-%d-%d", id, j)),
					jwt.WithTokenType(jwt.AccessToken),
					jwt.WithExpiresAt(time.Now().Add(time.Hour)),
					jwt.WithRoles([]string{"user", "admin"}),
					jwt.WithPermissions([]string{"read", "write"}),
					jwt.WithDeviceID(fmt.Sprintf("device-%d", j)),
					jwt.WithIPAddress(fmt.Sprintf("192.168.1.%d", j%255)),
				)

				// 验证Claims
				if err := claims.Validate(); err != nil {
					errCh <- err
					return
				}

				// 生成token
				token, err := tm.GenerateToken(context.Background(), claims)
				if err != nil {
					errCh <- err
					return
				}

				// 验证token
				_, err = tm.ValidateToken(context.Background(), token)
				if err != nil {
					errCh <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("unexpected error: %v", err)
	}
}
