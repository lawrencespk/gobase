package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/client/redis"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// TestClaims 用于测试的Claims实现
type TestClaims struct {
	jwtv5.RegisteredClaims               // 嵌入标准Claims
	UserID                 string        `json:"user_id"`
	UserName               string        `json:"user_name"`
	Roles                  []string      `json:"roles"`
	Permissions            []string      `json:"permissions"`
	DeviceID               string        `json:"device_id"`
	IPAddress              string        `json:"ip_address"`
	TokenType              jwt.TokenType `json:"token_type"`
	TokenID                string        `json:"token_id"`
}

// 实现 jwt.Claims 接口
func (c *TestClaims) GetUserID() string           { return c.UserID }
func (c *TestClaims) GetUserName() string         { return c.UserName }
func (c *TestClaims) GetRoles() []string          { return c.Roles }
func (c *TestClaims) GetPermissions() []string    { return c.Permissions }
func (c *TestClaims) GetDeviceID() string         { return c.DeviceID }
func (c *TestClaims) GetIPAddress() string        { return c.IPAddress }
func (c *TestClaims) GetTokenType() jwt.TokenType { return c.TokenType }
func (c *TestClaims) GetTokenID() string          { return c.TokenID }
func (c *TestClaims) Validate() error             { return nil }

// newTestClaims 创建测试用Claims
func newTestClaims(userID, deviceID string) *TestClaims {
	return &TestClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			NotBefore: jwtv5.NewNumericDate(time.Now()),
			Issuer:    "test",
			Subject:   userID,
			ID:        deviceID,
			Audience:  []string{"test"},
		},
		UserID:      userID,
		UserName:    "test-user",
		Roles:       []string{"user"},
		Permissions: []string{"read"},
		DeviceID:    deviceID,
		IPAddress:   "127.0.0.1",
		TokenType:   jwt.AccessToken,
		TokenID:     deviceID,
	}
}

func TestMain(m *testing.M) {
	// 在所有测试开始前设置
	cleanup := setupTestEnvironment()

	// 运行测试
	code := m.Run()

	// 清理资源
	cleanup()

	os.Exit(code)
}

func setupTestEnvironment() func() {
	// 初始化并注册 metrics collector
	if err := binding.RegisterCollector(); err != nil {
		panic(err)
	}

	// 返回清理函数
	return func() {
		// 清理所有测试资源
	}
}

func setupRedisStore(t *testing.T) (binding.Store, func()) {
	// 创建Redis客户端
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{"localhost:6379"}),
		redis.WithDB(1),          // 使用单独的DB做测试
		redis.WithMetrics(false), // 禁用metrics
	)
	require.NoError(t, err)

	// 清理测试数据
	_, err = redisClient.Del(context.Background(), "*")
	require.NoError(t, err)

	// 创建Redis存储
	store, err := binding.NewRedisStore(redisClient)
	require.NoError(t, err)

	// 返回清理函数
	cleanup := func() {
		// 确保先关闭 store
		if err := store.Close(); err != nil {
			t.Logf("failed to close store: %v", err)
		}
		// 清理 Redis 数据
		if _, err := redisClient.Del(context.Background(), "*"); err != nil {
			t.Logf("failed to cleanup redis data: %v", err)
		}
	}

	return store, cleanup
}

func TestBinding_Integration(t *testing.T) {
	// 设置存储并获取清理函数
	store, cleanup := setupRedisStore(t)
	// 确保在测试结束时清理资源
	t.Cleanup(cleanup)

	// 创建设备验证器
	deviceValidator, err := binding.NewDeviceValidator(store)
	require.NoError(t, err)

	// 创建IP验证器
	ipValidator, err := binding.NewIPValidator(store)
	require.NoError(t, err)

	t.Run("ValidateDevice", func(t *testing.T) {
		// 创建测试数据
		claims := newTestClaims("test-user", "test-device")

		device := &binding.DeviceInfo{
			ID:          "test-device",
			Type:        "mobile",
			Name:        "iPhone",
			OS:          "iOS",
			Browser:     "Safari",
			Fingerprint: "test-fp",
		}

		// 保存设备绑定
		err = store.SaveDeviceBinding(context.Background(), claims.UserID, claims.DeviceID, device)
		require.NoError(t, err)

		// 验证设备绑定
		err = deviceValidator.ValidateDevice(context.Background(), claims, device)
		require.NoError(t, err)
	})

	t.Run("ValidateIP", func(t *testing.T) {
		// 创建测试数据
		claims := newTestClaims("test-user", "test-device")
		ip := "127.0.0.1"

		// 保存IP绑定
		err = store.SaveIPBinding(context.Background(), claims.UserID, claims.DeviceID, ip)
		require.NoError(t, err)

		// 验证IP绑定
		err = ipValidator.ValidateIP(context.Background(), claims, ip)
		require.NoError(t, err)
	})
}
