package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/errors"
	"gobase/pkg/middleware/jwt/validator"
)

func TestBlacklistValidator_Integration(t *testing.T) {
	// 设置测试环境
	addr, err := testutils.StartRedisSingleContainer()
	require.NoError(t, err)
	defer testutils.CleanupRedisContainers()

	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{addr}),
		redis.WithDialTimeout(time.Second*5),
	)
	require.NoError(t, err)
	defer redisClient.Close()

	// 创建黑名单存储 - 使用 DefaultOptions
	store, err := blacklist.NewRedisStore(redisClient, blacklist.DefaultOptions())
	require.NoError(t, err)

	// 使用适配器创建 TokenBlacklist
	bl := blacklist.NewStoreAdapter(store)

	// 创建验证器
	v := validator.NewBlacklistValidator(bl)

	// 测试用例
	tests := []struct {
		name     string
		setup    func(context.Context) error
		claims   jwt.Claims
		setupCtx func(*gin.Context)
		wantErr  bool
		checkErr func(error) bool
	}{
		{
			name: "验证成功-token不在黑名单中",
			setup: func(ctx context.Context) error {
				return nil // 不需要预设置
			},
			claims: newTestClaims("test-user", "test-device", "127.0.0.1"),
			setupCtx: func(c *gin.Context) {
				c.Set("jwt_token", "valid.token")
			},
			wantErr: false,
		},
		{
			name: "token在黑名单中",
			setup: func(ctx context.Context) error {
				return bl.Add(ctx, "revoked.token", time.Hour)
			},
			claims: newTestClaims("test-user", "test-device", "127.0.0.1"),
			setupCtx: func(c *gin.Context) {
				c.Set("jwt_token", "revoked.token")
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewTokenRevokedError("", nil))
			},
		},
		{
			name: "上下文中没有token",
			setup: func(ctx context.Context) error {
				return nil
			},
			claims:   newTestClaims("test-user", "test-device", "127.0.0.1"),
			setupCtx: func(c *gin.Context) {},
			wantErr:  true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewTokenNotFoundError("", nil))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 执行测试用例
			err := tt.setup(context.Background())
			require.NoError(t, err)

			c, _ := gin.CreateTestContext(nil)
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			err = v.Validate(c, tt.claims)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "错误类型不匹配: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
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
