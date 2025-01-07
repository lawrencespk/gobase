package benchmark

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func BenchmarkTokenManager_GenerateToken(b *testing.B) {
	// 初始化
	log, _ := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	tm, _ := jwt.NewTokenManager("test-secret",
		jwt.WithLogger(log),
		jwt.WithoutMetrics(),
		jwt.WithoutTracing(),
	)
	ctx := context.Background()

	// 准备测试数据
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithUserName("Test User"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
		jwt.WithRoles([]string{"user"}),
		jwt.WithPermissions([]string{"read"}),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = tm.GenerateToken(ctx, claims)
		}
	})
}

func BenchmarkTokenManager_ValidateToken(b *testing.B) {
	// 初始化
	log, _ := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	tm, _ := jwt.NewTokenManager("test-secret",
		jwt.WithLogger(log),
		jwt.WithoutMetrics(),
		jwt.WithoutTracing(),
	)
	ctx := context.Background()

	// 生成测试token
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithUserName("Test User"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
		jwt.WithRoles([]string{"user"}),
		jwt.WithPermissions([]string{"read"}),
	)
	token, _ := tm.GenerateToken(ctx, claims)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = tm.ValidateToken(ctx, token)
		}
	})
}

// BenchmarkStandardClaims_Validate 测试Claims验证性能
func BenchmarkStandardClaims_Validate(b *testing.B) {
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithUserName("Test User"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
		jwt.WithRoles([]string{"user", "admin"}),
		jwt.WithPermissions([]string{"read", "write"}),
		jwt.WithDeviceID("device-123"),
		jwt.WithIPAddress("192.168.1.1"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = claims.Validate()
		}
	})
}

// BenchmarkToken_SignedString 测试token签名性能
func BenchmarkToken_SignedString(b *testing.B) {
	claims := jwt.NewStandardClaims(
		jwt.WithUserID("test-user"),
		jwt.WithTokenType(jwt.AccessToken),
		jwt.WithExpiresAt(time.Now().Add(time.Hour)),
	)
	token, _ := jwt.NewToken(
		jwt.WithClaims(claims),
		jwt.WithSecret("test-secret"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = token.SignedString()
		}
	})
}
