package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/security"
	redisCache "gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/logger"
)

func BenchmarkSecurity(b *testing.B) {
	ctx := context.Background()
	log, err := logger.NewLogger()
	if err != nil {
		b.Fatal(err)
	}

	// 初始化Redis客户端
	client, err := redisClient.NewClientFromConfig(&redisClient.Config{
		Addresses:    []string{"localhost:6379"},
		Database:     0,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	if err != nil {
		b.Fatal(err)
	}

	// 创建Redis缓存
	cache, err := redisCache.NewCache(redisCache.Options{
		Client: client,
		Logger: log,
	})
	if err != nil {
		b.Fatal(err)
	}

	// 创建安全策略
	policy := security.NewPolicy(
		security.WithCache(cache),
		security.WithLogger(log),
	)
	policy.UpdateConfig(time.Hour, time.Second)

	validator := security.NewTokenValidator(policy)

	// 准备测试数据
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithTokenID("test-token"),
		jwt.WithDeviceID("test-device"),
		jwt.WithIPAddress("127.0.0.1"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
	)

	tokenInfo := &jwt.TokenInfo{
		Type:      jwt.AccessToken,
		Claims:    claims,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	b.Run("ValidateTokenReuse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenID := fmt.Sprintf("token-%d", i)
			_ = policy.ValidateTokenReuse(ctx, tokenID)
		}
	})

	b.Run("ValidateToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.ValidateToken(ctx, tokenInfo)
		}
	})

	b.Run("ValidateTokenParallel", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = validator.ValidateToken(ctx, tokenInfo)
			}
		})
	})
}
