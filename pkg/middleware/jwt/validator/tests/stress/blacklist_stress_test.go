package stress

import (
	"context"
	"fmt"
	"sync"
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

func TestBlacklistValidator_Stress(t *testing.T) {
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

	// 创建黑名单存储
	store, err := blacklist.NewRedisStore(redisClient, blacklist.DefaultOptions())
	require.NoError(t, err)

	// 使用适配器创建 TokenBlacklist
	bl := blacklist.NewStoreAdapter(store)

	// 创建验证器
	v := validator.NewBlacklistValidator(bl)

	// 压力测试参数
	const (
		goroutines = 10  // 并发goroutine数
		requests   = 100 // 每个goroutine的请求数
	)

	// 等待组
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 错误通道
	errCh := make(chan error, goroutines*requests)

	// 开始时间
	start := time.Now()

	// 启动多个goroutine进行并发测试
	for i := 0; i < goroutines; i++ {
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < requests; j++ {
				// 创建测试token和claims
				token := fmt.Sprintf("test-token-%d-%d", routineID, j)
				claims := newTestClaims(
					fmt.Sprintf("user-%d", routineID),
					fmt.Sprintf("device-%d", j),
					"127.0.0.1",
				)

				// 随机将一些token加入黑名单
				if j%2 == 0 {
					if err := bl.Add(context.Background(), token, time.Hour); err != nil {
						errCh <- fmt.Errorf("failed to add token to blacklist: %v", err)
						continue
					}
				}

				// 创建gin上下文
				c, _ := gin.CreateTestContext(nil)
				c.Set("jwt_token", token)

				// 验证token
				if err := v.Validate(c, claims); err != nil {
					// 如果token在黑名单中，这是预期的错误
					if j%2 == 0 {
						continue
					}
					errCh <- fmt.Errorf("unexpected validation error: %v", err)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errCh)

	// 检查错误
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	// 报告结果
	duration := time.Since(start)
	totalRequests := goroutines * requests
	successRate := float64(totalRequests-len(errors)) / float64(totalRequests) * 100

	t.Logf("压力测试结果:")
	t.Logf("总请求数: %d", totalRequests)
	t.Logf("成功请求数: %d", totalRequests-len(errors))
	t.Logf("失败请求数: %d", len(errors))
	t.Logf("成功率: %.2f%%", successRate)
	t.Logf("总耗时: %v", duration)
	t.Logf("平均响应时间: %v", duration/time.Duration(totalRequests))

	// 如果错误率超过1%，测试失败
	maxErrorRate := 0.01
	actualErrorRate := float64(len(errors)) / float64(totalRequests)
	require.Less(t, actualErrorRate, maxErrorRate, "错误率过高")
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
