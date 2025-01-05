package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/security"
	redisCache "gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/logger"
)

func TestSecurity_Stress(t *testing.T) {
	// 初始化依赖
	ctx := context.Background()
	log, err := logger.NewLogger()
	require.NoError(t, err)

	// 初始化Redis客户端
	client, err := redisClient.NewClientFromConfig(&redisClient.Config{
		Addresses:    []string{"localhost:6379"},
		Database:     0,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	require.NoError(t, err)

	// 创建Redis缓存
	cache, err := redisCache.NewCache(redisCache.Options{
		Client: client,
		Logger: log,
	})
	require.NoError(t, err)

	// 创建安全策略
	policy := security.NewPolicy(
		security.WithCache(cache),
		security.WithLogger(log),
	)
	policy.UpdateConfig(time.Hour, time.Second)

	validator := security.NewTokenValidator(policy)

	// 并发测试参数
	const (
		goroutines = 100  // 并发goroutine数
		operations = 1000 // 每个goroutine的操作数
	)

	t.Run("并发Token重用检测", func(t *testing.T) {
		var wg sync.WaitGroup
		errCh := make(chan error, goroutines*operations)

		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for i := 0; i < operations; i++ {
					tokenID := fmt.Sprintf("token-%d-%d", routineID, i)

					// 第一次使用token应该成功
					err := policy.ValidateTokenReuse(ctx, tokenID)
					if err != nil {
						errCh <- fmt.Errorf("first use failed: %w", err)
						continue
					}

					// 立即重用token应该失败
					err = policy.ValidateTokenReuse(ctx, tokenID)
					if err == nil {
						errCh <- fmt.Errorf("reuse detection failed for token: %s", tokenID)
					}
				}
			}(g)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Error(err)
		}
	})

	t.Run("并发Token验证", func(t *testing.T) {
		var wg sync.WaitGroup
		errCh := make(chan error, goroutines*operations)

		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for i := 0; i < operations; i++ {
					claims := jwt.NewStandardClaims(
						jwt.WithUserID(fmt.Sprintf("user-%d", routineID)),
						jwt.WithTokenID(fmt.Sprintf("token-%d-%d", routineID, i)),
						jwt.WithDeviceID(fmt.Sprintf("device-%d", routineID)),
						jwt.WithIPAddress("127.0.0.1"),
						jwt.WithTokenType(jwt.AccessToken),
						jwt.WithExpiresAt(time.Now().Add(time.Hour)),
					)

					tokenInfo := &jwt.TokenInfo{
						Type:      jwt.AccessToken,
						Claims:    claims,
						ExpiresAt: time.Now().Add(time.Hour),
					}

					if err := validator.ValidateToken(ctx, tokenInfo); err != nil {
						errCh <- fmt.Errorf("validation failed: %w", err)
					}
				}
			}(g)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Error(err)
		}
	})
}
