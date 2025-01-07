package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/middleware/jwt/validator"
)

func BenchmarkBlacklistValidator(b *testing.B) {
	// 设置测试环境
	addr, err := testutils.StartRedisSingleContainer()
	require.NoError(b, err)
	defer testutils.CleanupRedisContainers()

	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{addr}),
		redis.WithDialTimeout(time.Second*5),
	)
	require.NoError(b, err)
	defer redisClient.Close()

	// 创建黑名单存储
	store, err := blacklist.NewRedisStore(redisClient, blacklist.DefaultOptions())
	require.NoError(b, err)

	// 使用适配器创建 TokenBlacklist
	bl := blacklist.NewStoreAdapter(store)

	// 创建验证器
	v := validator.NewBlacklistValidator(bl)

	// 设置gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 基准测试场景
	scenarios := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{
			name: "ValidToken",
			fn: func(b *testing.B) {
				// 重置计时器，确保设置过程不计入基准测试时间
				b.StopTimer()

				// 为本次测试创建token和claims
				token := fmt.Sprintf("test-token-valid-%d", b.N)
				claims := newTestClaims(
					fmt.Sprintf("user-%d", b.N),
					fmt.Sprintf("device-%d", b.N),
					"127.0.0.1",
				)

				b.StartTimer()

				// 运行b.N次测试
				for i := 0; i < b.N; i++ {
					c, _ := gin.CreateTestContext(nil)
					c.Set("jwt_token", token)
					_ = v.Validate(c, claims)
				}
			},
		},
		{
			name: "BlacklistedToken",
			fn: func(b *testing.B) {
				// 重置计时器，确保设置过程不计入基准测试时间
				b.StopTimer()

				// 为本次测试创建token和claims
				token := fmt.Sprintf("test-token-blacklisted-%d", b.N)
				claims := newTestClaims(
					fmt.Sprintf("user-%d", b.N),
					fmt.Sprintf("device-%d", b.N),
					"127.0.0.1",
				)

				// 将token加入黑名单
				err := bl.Add(context.Background(), token, time.Hour)
				require.NoError(b, err)

				b.StartTimer()

				// 运行b.N次测试
				for i := 0; i < b.N; i++ {
					c, _ := gin.CreateTestContext(nil)
					c.Set("jwt_token", token)
					_ = v.Validate(c, claims)
				}
			},
		},
	}

	for _, s := range scenarios {
		b.Run(s.name, s.fn)
	}
}

// TestClaims 用于测试的Claims实现
type TestClaims struct {
	*jwt.StandardClaims
}

// newTestClaims 创建测试用Claims
func newTestClaims(userID, deviceID, ipAddress string) *TestClaims {
	return &TestClaims{
		StandardClaims: jwt.NewStandardClaims(
			jwt.WithUserID(userID),
			jwt.WithDeviceID(deviceID),
			jwt.WithIPAddress(ipAddress),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(time.Hour)),
			jwt.WithTokenID("test-token"),
		),
	}
}
