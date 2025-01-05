package integration

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtPkg "gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
	"gobase/pkg/auth/jwt/security"
	"gobase/pkg/cache/redis"
	redisTestutils "gobase/pkg/cache/redis/tests/testutils"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger"
	"gobase/pkg/monitor/prometheus/metrics"
)

// testClaims 测试用的Claims实现
type testClaims struct {
	userID      string
	userName    string
	roles       []string
	permissions []string
	deviceID    string
	ipAddress   string
	tokenType   jwtPkg.TokenType
	tokenID     string
	expiresAt   time.Time
	issuedAt    time.Time
	notBefore   time.Time
	audience    []string
	issuer      string
	subject     string
}

// 实现 Claims 接口的所有必需方法
func (c *testClaims) GetUserID() string              { return c.userID }
func (c *testClaims) GetUserName() string            { return c.userName }
func (c *testClaims) GetRoles() []string             { return c.roles }
func (c *testClaims) GetPermissions() []string       { return c.permissions }
func (c *testClaims) GetDeviceID() string            { return c.deviceID }
func (c *testClaims) GetIPAddress() string           { return c.ipAddress }
func (c *testClaims) GetTokenType() jwtPkg.TokenType { return c.tokenType }
func (c *testClaims) GetTokenID() string             { return c.tokenID }
func (c *testClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.expiresAt), nil
}
func (c *testClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.issuedAt), nil
}
func (c *testClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.notBefore), nil
}
func (c *testClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(c.audience), nil
}
func (c *testClaims) GetIssuer() (string, error) {
	return c.issuer, nil
}
func (c *testClaims) GetSubject() (string, error) {
	return c.subject, nil
}
func (c *testClaims) Validate() error { return nil }

func TestSecurity_Integration(t *testing.T) {
	// 设置测试环境
	ctx := context.Background()
	log, err := logger.NewLogger()
	require.NoError(t, err)

	// 启动Redis容器
	cleanup, err := redisTestutils.StartRedisContainer()
	require.NoError(t, err)
	defer cleanup()

	// 创建Redis客户端
	client, err := redisClient.NewClient(
		redisClient.WithAddresses([]string{"localhost:6379"}),
		redisClient.WithLogger(log),
	)
	require.NoError(t, err)
	defer client.Close()

	// 创建Redis缓存
	cache, err := redis.NewCache(redis.Options{
		Client: client,
		Logger: log,
	})
	require.NoError(t, err)

	// 初始化密钥管理器
	keyManager, err := crypto.NewKeyManager(
		jwtPkg.HS256,
		log,
	)
	require.NoError(t, err)

	// 初始化安全策略
	policy := security.NewPolicy(
		security.WithCache(cache),
		security.WithLogger(log),
		security.WithMetrics(metrics.NewJWTMetrics()),
	)

	// 配置策略参数 - 设置较短的重用间隔以便测试
	policy.UpdateConfig(time.Hour, time.Second)

	// 初始化密钥轮换器
	rotator, err := security.NewKeyRotator(keyManager, policy, log)
	require.NoError(t, err)

	// 初始化Token验证器
	validator := security.NewTokenValidator(policy)

	t.Run("完整Token生命周期", func(t *testing.T) {
		// 启动密钥轮换
		err := rotator.Start(ctx)
		require.NoError(t, err)
		defer rotator.Stop()

		// 生成测试Token
		claims := &testClaims{
			userID:      "test-user",
			userName:    "Test User",
			roles:       []string{"user"},
			permissions: []string{"read"},
			deviceID:    "test-device",
			ipAddress:   "192.168.1.1",
			tokenType:   jwtPkg.AccessToken,
			tokenID:     "test-token-id",
			expiresAt:   time.Now().Add(time.Hour),
			issuedAt:    time.Now(),
			notBefore:   time.Now(),
			audience:    []string{"test-audience"},
			issuer:      "test-issuer",
			subject:     "test-subject",
		}

		tokenInfo := &jwtPkg.TokenInfo{
			Raw:       "test-token-raw",
			Type:      jwtPkg.AccessToken,
			Claims:    claims,
			ExpiresAt: claims.expiresAt,
		}

		// 验证Token
		err = validator.ValidateToken(ctx, tokenInfo)
		assert.NoError(t, err)

		// 第一次验证Token重用 - 应该成功
		err = policy.ValidateTokenReuse(ctx, "test-token-1")
		assert.NoError(t, err)

		// 立即再次验证 - 应该失败
		err = policy.ValidateTokenReuse(ctx, "test-token-1")
		assert.Error(t, err)
		assert.True(t, errors.HasErrorCode(err, codes.PolicyViolation))

		// 等待重用间隔后再次验证 - 应该成功
		time.Sleep(time.Second * 2)
		err = policy.ValidateTokenReuse(ctx, "test-token-1")
		assert.NoError(t, err)
	})
}
