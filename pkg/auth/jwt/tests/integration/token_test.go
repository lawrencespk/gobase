package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	errorTypes "gobase/pkg/errors/types"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
	"gobase/pkg/trace/jaeger"
)

func TestTokenManager_Integration(t *testing.T) {
	// 创建自定义logger用于测试
	log, err := logger.NewLogger(
		logger.WithLevel(types.DebugLevel),
		logger.WithOutputPaths([]string{"stdout", "./logs/integration_test.log"}),
	)
	require.NoError(t, err)

	// 创建TokenManager
	tm, err := jwt.NewTokenManager(
		"test-secret-key",
		jwt.WithLogger(log), // 需要在TokenManager中添加这个选项
	)
	require.NoError(t, err)

	t.Run("完整的Token生命周期测试", func(t *testing.T) {
		ctx := context.Background()

		// 1. 生成Claims
		claims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithUserName("Test User"),
			jwt.WithRoles([]string{"admin"}),
			jwt.WithPermissions([]string{"read", "write"}),
			jwt.WithDeviceID("device-123"),
			jwt.WithIPAddress("127.0.0.1"),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithTokenID("token-123"),
			jwt.WithExpiresAt(time.Now().Add(time.Hour)),
		)

		// 2. 生成Token
		token, err := tm.GenerateToken(ctx, claims)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// 3. 验证Token
		parsedToken, err := tm.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.NotNil(t, parsedToken)

		// 4. 解析Claims
		parsedClaims, ok := parsedToken.Claims.(*jwt.StandardClaims)
		require.True(t, ok)

		// 5. 验证Claims字段
		assert.Equal(t, "test-user", parsedClaims.GetUserID())
		assert.Equal(t, "Test User", parsedClaims.GetUserName())
		assert.Equal(t, []string{"admin"}, parsedClaims.GetRoles())
		assert.Equal(t, []string{"read", "write"}, parsedClaims.GetPermissions())
		assert.Equal(t, "device-123", parsedClaims.GetDeviceID())
		assert.Equal(t, "127.0.0.1", parsedClaims.GetIPAddress())
		assert.Equal(t, jwt.AccessToken, parsedClaims.GetTokenType())
		assert.Equal(t, "token-123", parsedClaims.GetTokenID())
	})

	t.Run("链路追踪集成测试", func(t *testing.T) {
		ctx := context.Background()

		// 创建根 span
		span, err := jaeger.NewSpan("jwt_test",
			jaeger.WithTag("test_type", "integration"),
			jaeger.WithParent(ctx),
		)
		require.NoError(t, err)
		defer span.Finish()

		ctx = span.Context()

		claims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
		)

		// 生成Token并验证span
		token, err := tm.GenerateToken(ctx, claims)
		require.NoError(t, err)

		// 验证Token并检查span
		_, err = tm.ValidateToken(ctx, token)
		require.NoError(t, err)
	})

	t.Run("监控指标集成测试", func(t *testing.T) {
		ctx := context.Background()
		claims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
		)

		// 执行操作
		token, err := tm.GenerateToken(ctx, claims)
		require.NoError(t, err)

		// 验证生成操作的耗时指标
		metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("generate").Observe(0)
		assert.NotPanics(t, func() {
			metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("generate").Observe(0)
		}, "应该能够记录生成Token的耗时")

		// 验证Token
		_, err = tm.ValidateToken(ctx, token)
		require.NoError(t, err)

		// 验证验证操作的耗时指标
		assert.NotPanics(t, func() {
			metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("validate").Observe(0)
		}, "应该能够记录验证Token的耗时")

		// 验证Token年龄指标
		metrics.DefaultJWTMetrics.TokenAgeGauge.Set(float64(time.Now().Unix()))

		// 生成一个过期的Token来测试错误计数
		expiredClaims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(time.Second)),
		)
		expiredToken, err := tm.GenerateToken(ctx, expiredClaims)
		require.NoError(t, err)

		// 等待Token过期
		time.Sleep(time.Second * 2)

		// 验证过期Token，应该触发错误计数器增加
		_, err = tm.ValidateToken(ctx, expiredToken)
		require.Error(t, err)

		// 验证错误计数器被调用
		assert.NotPanics(t, func() {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "expired").Inc()
		}, "错误计数器应该能够正常工作")
	})

	t.Run("错误处理集成测试", func(t *testing.T) {
		ctx := context.Background()

		// 1. 测试无效Token格式
		_, err := tm.ValidateToken(ctx, "invalid.token.format")
		assert.True(t, errors.IsTokenInvalidError(err), "应该是TokenInvalid错误")

		// 2. 测试过期Token
		expiredClaims := jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(time.Second)),
		)

		// 生成快要过期的token
		expiredToken, err := tm.GenerateToken(ctx, expiredClaims)
		require.NoError(t, err, "生成token失败")

		// 等待token过期
		time.Sleep(time.Second * 2)

		// 验证过期token
		_, err = tm.ValidateToken(ctx, expiredToken)
		require.Error(t, err, "应该返回错误")

		// 添加调试信息
		t.Logf("返回的错误类型: %T", err)
		t.Logf("错误信息: %v", err)

		if customErr, ok := err.(errorTypes.Error); ok {
			t.Logf("自定义错误码: %s", customErr.Code())
			t.Logf("自定义错误信息: %s", customErr.Message())
			t.Logf("原始错误: %v", customErr.Unwrap())
			t.Logf("错误堆栈: %v", customErr.Stack())
		}

		// 检查是否为过期错误
		isExpired := errors.IsTokenExpiredError(err)
		t.Logf("是否为过期错误: %v", isExpired)

		assert.True(t, isExpired, "应该是TokenExpired错误")

		// 3. 测试无效签名
		validToken, _ := tm.GenerateToken(ctx, jwt.NewStandardClaims(
			jwt.WithUserID("test-user"),
			jwt.WithTokenType(jwt.AccessToken),
		))
		invalidSignatureToken := validToken[:len(validToken)-1] + "X"
		_, err = tm.ValidateToken(ctx, invalidSignatureToken)
		assert.True(t, errors.IsSignatureInvalidError(err), "应该是SignatureInvalid错误")
	})
}
